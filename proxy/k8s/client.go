package k8s

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

const (
	POD_NAME_ENV = "POD_NAME"
	POD_NAMESPACE_ENV = "POD_NAMESPACE"
)

type K8sManager struct {
	log logr.Logger
	client *kubernetes.Clientset
	podName string
	podNamespace string
}

func NewK8sManager(log logr.Logger) (*K8sManager, error) {
	podName := os.Getenv(POD_NAME_ENV)
	if podName == "" {
		log.Info("Failed to find pod name from environment","env name", POD_NAME_ENV)
		return nil, nil
	}
	podNamespace := os.Getenv(POD_NAMESPACE_ENV)
	if podName == "" {
		log.Info("Failed to find pod namespace from environment","env name", POD_NAMESPACE_ENV)
		return nil, nil
	}
	client, err := getK8sClient(log)
	if client == nil || err != nil{
		return nil, err
	} else {
		return &K8sManager{
			log:    log,
			client: client,
			podName: podName,
			podNamespace: podNamespace,
		}, nil
	}
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



