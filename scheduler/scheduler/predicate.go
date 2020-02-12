package scheduler

import (
	"github.com/go-logr/logr"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"math/rand"
	"strconv"
)

const RESOURCES_TRTIS_GPU_MEMORY = "seldon.io/trtis-gpu-mem"
const ANNOTATION_MODEL_ID = "seldon.io/trtis-model-id" // ID to ensure model loaded once on each node

func getUsedGpuMemoryOnNode(clientSet *kubernetes.Clientset, node *v1.Node, logger logr.Logger) (*int64, map[string]bool, error) {
	//Get pods on node
	pods, err := clientSet.CoreV1().Pods("").List(metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node.Name,
	})
	if err != nil {
		return nil, nil, err
	}
	modelIds := make(map[string]bool)
	var requestedGpuMemory int64
	for _, pod := range pods.Items {
		logger.Info("Looking at pod ", "name", pod.Name)
		if modelId := pod.Annotations[ANNOTATION_MODEL_ID]; modelId != "" {
			modelIds[modelId] = true
		}
		for _, c := range pod.Spec.Containers {
			// There always needs to be a limit for non default resource types
			if limitMem, ok := c.Resources.Limits[RESOURCES_TRTIS_GPU_MEMORY]; ok {
				requestedGpuMemory += limitMem.Value()
			}
		}
	}
	return &requestedGpuMemory, modelIds, nil
}

func trtisPredicate(clientSet *kubernetes.Clientset, node *v1.Node, pod *v1.Pod, logger logr.Logger) bool {
	if memNode, ok := node.Annotations[ANNOTATION_TRTIS_GPU_MEMORY_TOTAL]; ok {
		totalNodeGPUMemory, err := strconv.ParseInt(memNode, 0, 64)
		if err != nil {
			logger.Error(err, "Failed to parse node memory")
			return false
		} else {
			logger.Info("Total GPU memory on node", "node", node.Name, ANNOTATION_TRTIS_GPU_MEMORY_TOTAL, totalNodeGPUMemory)
		}

		usedGpuMemory, modelIds, err := getUsedGpuMemoryOnNode(clientSet, node, logger)
		if err != nil {
			logger.Error(err, "Failed to get GPU Memory used on node")
			return false
		} else {
			logger.Info("Memory already requested on node", "node", node.Name, "GPU memory used", usedGpuMemory, "modelIds", modelIds)
		}

		modelId := pod.Annotations[ANNOTATION_MODEL_ID]
		if modelId != "" {
			if modelIds[modelId] {
				logger.Info("Model already on node", "id", modelId)
				return false
			}
		} else {
			logger.Info("Failed to find model name : continuning with anonymous model")
		}

		availableGPUMemory := totalNodeGPUMemory - *usedGpuMemory

		// Calculate the GPU memory limit from container limits
		var limitMemorySum int64
		for _, c := range pod.Spec.Containers {
			// There always needs to be a limit for non default resource types
			if limitMem, ok := c.Resources.Limits[RESOURCES_TRTIS_GPU_MEMORY]; ok {
				limitMemorySum += limitMem.Value()
			}
		}

		logger.Info("Requested memory ", RESOURCES_TRTIS_GPU_MEMORY, limitMemorySum)
		if availableGPUMemory > limitMemorySum {
			remaining := availableGPUMemory - limitMemorySum
			logger.Info("found fitting node","requested",limitMemorySum,"available",availableGPUMemory,"total", totalNodeGPUMemory, "used",*usedGpuMemory,"remaining",remaining)
			return true
		} else {
			logger.Info("no space on node","requested", limitMemorySum, "available",availableGPUMemory,"total", totalNodeGPUMemory, "used",*usedGpuMemory)
		}
	}
	log.Println("Failed node placement")
	return false
}

func randomPredicate(node *v1.Node, pod *v1.Pod, logger logr.Logger) bool {
	r := rand.Intn(2)
	return r == 0
}
