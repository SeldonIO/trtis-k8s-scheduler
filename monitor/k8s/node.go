package k8s

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/seldonio/trtis-scheduler/monitor/metric"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	kclient "kmodules.xyz/client-go/core/v1"
)

const (
	ANNOTATION_TRTIS_GPU_MEMORY_USED     = "seldon.io/trtis-gpu-mem-used"
	ANNOTATION_TRTIS_GPU_MEMORY_TOTAL    = "seldon.io/trtis-gpu-mem-total"
	ANNOTATION_TRTIS_GPU_MEMORY_UTIL    = "seldon.io/trtis-gpu-util"
)

type NodeAnnotator struct {
	client   *kubernetes.Clientset
	nodeName string
	log      logr.Logger
	amap map[string]string
}

func NewNodeAnnotator(nodeName string, log logr.Logger) (*NodeAnnotator, error) {
	client, err := getK8sClient(log)
	if err != nil {
		return nil, err
	}
	return &NodeAnnotator{
		client:   client,
		nodeName: nodeName,
		log:      log.WithName("NodeAnnotator"),
		amap: map[string]string{metric.Nv_gpu_memory_total_bytes:ANNOTATION_TRTIS_GPU_MEMORY_TOTAL,metric.Nv_gpu_memory_used_bytes:ANNOTATION_TRTIS_GPU_MEMORY_USED,metric.Nv_gpu_utilization:ANNOTATION_TRTIS_GPU_MEMORY_UTIL},
	}, nil
}

func getK8sClient(log logr.Logger) (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "failed to get in cluster config")
		return nil, nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "Failed to get k8s clientset")
		return nil, err
	}

	return clientset, nil
}

func (n *NodeAnnotator) PatchNodeAnnotation(gpuMap map[string]*float64) error {
	// Get Node
	node, err := n.client.CoreV1().Nodes().Get(n.nodeName, metav1.GetOptions{})
	if err != nil {
		n.log.Error(err, "Failed to get node", "nodeName", n.nodeName)
		return err
	}
	//Patch Node
	_, _, err = kclient.PatchNode(n.client, node, func(nIn *v1.Node) *v1.Node {
		nOut := nIn.DeepCopy()
		for k,v := range(gpuMap) {
			if v != nil {
				convertedKey := n.amap[k]
				if convertedKey != "" {
					n.log.Info("Updating annotation", "key", convertedKey, "value", v)
					nOut.Annotations[convertedKey] = fmt.Sprintf("%d", int(*v))
				} else {
					n.log.Info("Failed to find gpu annotation mapping", "key", k, "value", v)
				}
			} else {
				n.log.Info("Skipping null value","key",k)
			}
		}
		return nOut
	})
	return err
}
