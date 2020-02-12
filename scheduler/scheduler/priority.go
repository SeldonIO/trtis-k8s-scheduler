package scheduler

import (
	"github.com/go-logr/logr"
	"k8s.io/api/core/v1"
	"math/rand"
)

func randomPriority(node *v1.Node, pod *v1.Pod, logger logr.Logger) int {
	return rand.Intn(100)
}
