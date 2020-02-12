# TRTIS K8S Scheduler

Proof of concept to schedule ML models onto [NVIDIA TensorRT Inference Servers](https://github.com/NVIDIA/tensorrt-inference-server) running as Kubernetes DaemonSets.

## Motivation
Provide the ability for ML models to share GPUs to save on infrastructure costs.

## GPU Sharing Goals
GPU sharing has several sub requirements.

 * **Scheduling**
    * Decide which GPU to attach a model
 * **Isolation**
    * Ensure multi-tenant users can not interfere with each other.
 * **Fairness**
    * Ensure each user has fair access to GPU resources.

We will mostly be concerned with scheduling.

## Existing Resources

 * [Ongoing discussion on GPU sharing in kubernetes issue](https://github.com/kubernetes/kubernetes/issues/52757)
 * [Fractional GPUs](https://github.com/sakjain92/Fractional-GPUs)
 * [Alibaba cloud GPU sharing](https://www.alibabacloud.com/blog/gpu-sharing-scheduler-extender-now-supports-fine-grained-kubernetes-clusters_594926)

## Proposal
Follow the work of Alibaba to provide a custom scheduler but rather than use a low level NVIDIA device plugin utilize TRTIS servers to run models.


## Components

 * TRTIS Daemonset
   * TRTIS running on each GPU node as a k8s Daemonset using an NFS back model repository.
 * Scheduler
   * Custom scheduler that looks for pods assigned to decide which node to place them.
 * Monitor
   * Runs alongside TRTIS server to expose GPU metrics onto node as annotations
 * Loader (initContainer)
   * Loads model onto TRTIS model repository for a node.
 * Proxy/Unloader
   * Optional proxy that forwards API requests to TRTIS server on node
   * Unloads model from server when terminated

## ML Pod Scheduling Requirements

To be scheduled a pod must:
  * Have a custom resource limit `seldon.io/trtis-gpu-mem`
    * This will specify the GPU memory required
  * Have a annotation for the model ID: `seldon.io/trtis-model-id`
    * This will ensure a model is not scheduled more than once on any node
  * Have custom schedulerName set: `schedulerName: trtis-scheduler`

In this demo the pod will be defined via a Deployment with the following containers

  * A 1st initContainer `gcr.io/kfserving/storage-initializer:0.2.1` to download model from cloud storage to local
  * A 2nd initContainer `seldonio/trtis-loader:0.1` to load model onto TRTIS model repo and wait for TRTIS to show its loaded
  * A container `seldonio/trtis-proxy:0.1`
    * Acts as an optional proxy for REST and GRPC requests to server as well as possible isolation enforcer to only allow requests to loaded model on server.
    * Unloads model on termination

## Scheduling Steps

  1. A pod with appropriate settings as discussed above is created. This could be done via an operator using a CRD for model definition, e.g. KFServing or Seldon.
  1. The TRTIS-Scheduler will currently:
     * For each node
        * Calculate the total memory for pods assigned to that node from their `seldon.io/trtis-gpu-mem`
        * Get the total available memory on node via node annotation `seldon.io/trtis-gpu-mem-total`
        * Check the running model IDs via the pod annotations `seldon.io/trtis-model-id`
     * A pod can be scheduled if there is enough memory and same model ID is not already on node
     * Choose a random node for available nodes to schedule pod and bind the pod to that node.
     * If no node satisfies the constraints the pod is placed back in the scheduling queue with an exponential backoff (max 2 mins). It will remain “Pending” in status field until scheduled.
  1. When the pod starts on the node it will
     * Download model from cloud storage
     * Upload model to TRTIS model repository on that node
        * Optionally in future update an “Endpoint” to add this Node to a Service for this model.
     * Wait for TRTIS server to say its loaded. (TRTIS server running in POLL mode)
     * The main container is a proxy to forward REST and GRPC requests
  1. On termination the pod deletes its folder from the TRTIS model repository.
     * Optionally in future remove this TRTIS node from the “Endpoint” for the Service for this model.

## API Requests

There are two options:

  1. Use the proxy service running on nodes.
     * Advantages
       * Allows for isolation of requests to ensure other models on the TRTIS server can not be called.
       * Easy to create a service which automatically load balances as more replicas of the model are created (manually or via auto-scaling).
     * Disadvantages
       * Adds an extra network hop and proxy step.
  1. Use custom Service and Endpoint.
     * Advantages
       * No proxy step
     * Disadvantages
       * No immediate model isolation
       * Needs custom controller to update endpoint as models are added to TRTIS nodes or removed. A standard Service is not possible as all TRTIS DaemonSets will have same labels.


## Demo

 * [A GKE Demo](samples/gke/README.md)