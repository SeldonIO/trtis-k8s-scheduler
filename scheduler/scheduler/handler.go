package scheduler

import (
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/api/core/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v12 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	log2 "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"time"
)

const (
	schedulerName                     = "trtis-scheduler"
	ANNOTATION_TRTIS_GPU_MEMORY_USED  = "seldon.io/trtis-gpu-mem-used"
	ANNOTATION_TRTIS_GPU_MEMORY_TOTAL = "seldon.io/trtis-gpu-mem-total"
	MAX_SCHEDULE_WAIT                 = 2*time.Minute + 2*time.Second
)

type predicateFunc func(clientSet *kubernetes.Clientset, node *v1.Node, pod *v1.Pod, logger logr.Logger) bool
type priorityFunc func(node *v1.Node, pod *v1.Pod, logger logr.Logger) int

type PodJob struct {
	Pod              *v1.Pod
	nextScheduleTime time.Duration
}

type Scheduler struct {
	clientset  *kubernetes.Clientset
	podQueue   chan *PodJob
	nodeLister v12.NodeLister
	predicates []predicateFunc
	priorities []priorityFunc
	logger     logr.Logger
}

func NewScheduler(podQueue chan *PodJob, quit chan struct{}) Scheduler {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	log2.SetLogger(log2.ZapLogger(false))
	logger := log2.Log.WithName("entrypoint")

	return Scheduler{
		clientset:  clientset,
		podQueue:   podQueue,
		nodeLister: initInformers(clientset, podQueue, quit, logger),
		predicates: []predicateFunc{
			trtisPredicate,
		},
		priorities: []priorityFunc{
			randomPriority,
		},
		logger: logger,
	}
}

func initInformers(clientset *kubernetes.Clientset, podQueue chan *PodJob, quit chan struct{}, logger logr.Logger) v12.NodeLister {
	factory := informers.NewSharedInformerFactory(clientset, 0)

	nodeInformer := factory.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*v1.Node)
			if !ok {
				logger.Info("Not a node")
				return
			}
			logger.Info("New Node Added to Store", "name", node.GetName())
		},
	})

	podInformer := factory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				logger.Info("this is not a pod")
				return
			}
			if pod.Spec.NodeName == "" && pod.Spec.SchedulerName == schedulerName {
				logger.Info("Adding pod to queue", "pod name", pod.Name)
				podQueue <- &PodJob{
					Pod:              pod,
					nextScheduleTime: 500 * time.Millisecond,
				}
			} else {
				logger.Info("Ignoring pod", "name", pod.Name)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*v1.Pod)
			if !ok {
				logger.Info("this is not a pod")
				return
			}
			if pod.Spec.SchedulerName == schedulerName && pod.Status.Phase == v1.PodRunning {
				logger.Info("Scheduled pod is running", "name", pod.Name, "node", pod.Spec.NodeName)
			}
		},
	})

	factory.Start(quit)
	return nodeInformer.Lister()
}

func (s *Scheduler) Run(quit chan struct{}) {
	wait.Until(s.ScheduleOne, 0, quit)
}

func (s *Scheduler) requeuePod(podJob *PodJob) {
	nextScheduleTime := podJob.nextScheduleTime * 2
	if nextScheduleTime > MAX_SCHEDULE_WAIT {
		nextScheduleTime = MAX_SCHEDULE_WAIT
	}
	s.logger.Info("Rescheduling pod ","wait time", podJob.nextScheduleTime)
	go func() {
		time.Sleep(podJob.nextScheduleTime)
		s.podQueue <- &PodJob{
			Pod:              podJob.Pod,
			nextScheduleTime: nextScheduleTime,
		}
	}()
}

func (s *Scheduler) ScheduleOne() {

	pj := <-s.podQueue
	p := pj.Pod
	s.logger.Info("found a pod to schedule", "namespace", p.Namespace, "name", p.Name)

	node, err := s.findFit(p)
	if err != nil {
		s.logger.Error(err, "cannot find node that fits pod")
		s.requeuePod(pj)
		return
	}

	err = s.bindPod(p, node)
	if err != nil {
		s.logger.Error(err, "failed to bind pod")
		s.requeuePod(pj)
		return
	}

	message := fmt.Sprintf("Placed pod [%s/%s] on %s\n", p.Namespace, p.Name, node)

	err = s.emitEvent(p, message)
	if err != nil {
		s.logger.Error(err, "failed to emit scheduled event")
		return
	}

	s.logger.Info(message)
}

func (s *Scheduler) findFit(pod *v1.Pod) (string, error) {
	nodes, err := s.nodeLister.List(labels.Everything())
	if err != nil {
		return "", err
	}

	filteredNodes := s.runPredicates(nodes, pod)
	if len(filteredNodes) == 0 {
		return "", errors.New("failed to find node that fits pod")
	}
	priorities := s.prioritize(filteredNodes, pod)
	return s.findBestNode(priorities), nil
}

func (s *Scheduler) bindPod(p *v1.Pod, node string) error {
	return s.clientset.CoreV1().Pods(p.Namespace).Bind(&v1.Binding{
		ObjectMeta: v13.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
		},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node,
		},
	})
}

func (s *Scheduler) emitEvent(p *v1.Pod, message string) error {
	timestamp := time.Now().UTC()
	_, err := s.clientset.CoreV1().Events(p.Namespace).Create(&v1.Event{
		Count:          1,
		Message:        message,
		Reason:         "Scheduled",
		LastTimestamp:  v13.NewTime(timestamp),
		FirstTimestamp: v13.NewTime(timestamp),
		Type:           "Normal",
		Source: v1.EventSource{
			Component: schedulerName,
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Pod",
			Name:      p.Name,
			Namespace: p.Namespace,
			UID:       p.UID,
		},
		ObjectMeta: v13.ObjectMeta{
			GenerateName: p.Name + "-",
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) runPredicates(nodes []*v1.Node, pod *v1.Pod) []*v1.Node {
	filteredNodes := make([]*v1.Node, 0)
	for _, node := range nodes {
		if s.predicatesApply(node, pod) {
			filteredNodes = append(filteredNodes, node)
		}
	}
	for _, n := range filteredNodes {
		s.logger.Info("Node fits: ", "name", n.Name)
	}
	return filteredNodes
}

func (s *Scheduler) predicatesApply(node *v1.Node, pod *v1.Pod) bool {
	for _, predicate := range s.predicates {
		if !predicate(s.clientset, node, pod, s.logger.WithName(node.Name)) {
			return false
		}
	}
	return true
}

func (s *Scheduler) prioritize(nodes []*v1.Node, pod *v1.Pod) map[string]int {
	priorities := make(map[string]int)
	for _, node := range nodes {
		for _, priority := range s.priorities {
			priorities[node.Name] += priority(node, pod, s.logger)
		}
	}
	s.logger.Info("calculated priorities:", "pritorities", priorities)
	return priorities
}

func (s *Scheduler) findBestNode(priorities map[string]int) string {
	var maxP int
	var bestNode string
	for node, p := range priorities {
		if p > maxP {
			maxP = p
			bestNode = node
		}
	}
	return bestNode
}
