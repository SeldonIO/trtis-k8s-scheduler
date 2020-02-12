# GPU Sharing Demo on GKE

## Create Infrastructure

Set project settings.

```
PROJECT=seldon-demos
ZONE=europe-west1-b
```

Create a disk for use with nfs. Call it nfs-disk.

```
gcloud beta compute disks create nfs-disk --project=${PROJECT} --type=pd-standard --size=500GB --zone=${ZONE} --physical-block-size=4096
```

Create a kubernetes cluster

```
gcloud beta container --project ${PROJECT} clusters create "gpu-sharing-demo" --zone ${ZONE} --no-enable-basic-auth --cluster-version "1.13.11-gke.23" --machine-type "n1-standard-8" --accelerator "type=nvidia-tesla-k80,count=1" --image-type "COS" --disk-type "pd-standard" --disk-size "100" --scopes "https://www.googleapis.com/auth/devstorage.read_only","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly","https://www.googleapis.com/auth/trace.append" --num-nodes "2" --enable-cloud-logging --enable-cloud-monitoring --enable-ip-alias --network "projects/seldon-demos/global/networks/default" --subnetwork "projects/seldon-demos/regions/europe-west1/subnetworks/default" --default-max-pods-per-node "110" --addons HorizontalPodAutoscaling,HttpLoadBalancing --enable-autoupgrade --enable-autorepair
```

Connect to cluster

```
gcloud container clusters get-credentials gpu-sharing-demo --zone ${ZONE} --project ${PROJECT}
```

## Setup

Install GPU daemonset

```
kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/container-engine-accelerators/master/nvidia-driver-installer/cos/daemonset-preloaded.yaml
```

Wait until ready.

Create NFS PVC that will utilize the disk created above as the model repo for all TRTIS servers.


```
make create-nfs
```

Install NVIDIA TRTIS as a daemonset on GPU nodes.

```
make create-trtis
```

Start scheduler.

```
make create-scheduler
```

Create a service for this demo to allow us to connect to TRTIS server. In production you would need to either use proxy or create a specific service for scheduled pods based on which servers they are running on.

```
make create-loadbalancer
```

Get the IP of the loadbalancer:

```
INGRESS=`kubectl get svc trtis-svc-demo -o jsonpath='{.status.loadBalancer.ingress[0].ip}'`
echo $INGRESS
```

Deploy a simple model

```
make deploy-simple-model
```

Test API

```
docker run  --rm --net=host nvcr.io/nvidia/tensorrtserver:20.01-py3-clientsdk simple_client -u ${INGRESS}:8000
```

You should see something similar to:

```
Health for model simple:
Live: 1
Ready: 1
Status for model simple:
id: "inference:0"
version: "1.10.0"
uptime_ns: 2630674624829
model_status {
  key: "simple"
  value {
    config {
      name: "simple"
      platform: "tensorflow_graphdef"
      version_policy {
        latest {
          num_versions: 1
        }
      }
      max_batch_size: 8
      input {
        name: "INPUT0"
        data_type: TYPE_INT32
        dims: 16
      }
      input {
        name: "INPUT1"
        data_type: TYPE_INT32
        dims: 16
      }
      output {
        name: "OUTPUT0"
        data_type: TYPE_INT32
        dims: 16
      }
      output {
        name: "OUTPUT1"
        data_type: TYPE_INT32
        dims: 16
      }
      instance_group {
        name: "simple"
        count: 1
        kind: KIND_CPU
      }
      default_model_filename: "model.graphdef"
    }
    version_status {
      key: 1
      value {
        ready_state: MODEL_READY
        ready_state_reason {
        }
      }
    }
  }
}
ready_state: SERVER_READY

0 + 1 = 1
0 - 1 = -1
1 + 1 = 2
1 - 1 = 0
2 + 1 = 3
2 - 1 = 1
3 + 1 = 4
3 - 1 = 2
4 + 1 = 5
4 - 1 = 3
5 + 1 = 6
5 - 1 = 4
6 + 1 = 7
6 - 1 = 5
7 + 1 = 8
7 - 1 = 6
8 + 1 = 9
8 - 1 = 7
9 + 1 = 10
9 - 1 = 8
10 + 1 = 11
10 - 1 = 9
11 + 1 = 12
11 - 1 = 10
12 + 1 = 13
12 - 1 = 11
13 + 1 = 14
13 - 1 = 12
14 + 1 = 15
14 - 1 = 13
15 + 1 = 16
15 - 1 = 14
```

Launch resnet model. This yaml asks for 11G of GPU memory which should fit on only 1 of the two nodes.

```
make deploy-resnet-model
```

Test API

```
docker run  --rm --net=host nvcr.io/nvidia/tensorrtserver:20.01-py3-clientsdk image_client -u ${INGRESS}:8000 -m resnet50_netdef -s INCEPTION images/mug.jpg
```

You should see something like:

```
Request 0, batch size 1
Image 'images/mug.jpg':
    504 (COFFEE MUG) = 0.723992
```

Now launch a model which can't be scheduled as it asks for 20Gi.

```
make deploy-resnet-model-toobig
```

You will see the pod remains in "Pending":

```
kubectl get pods -l app=trtis-model-resnet-big
```

You should see something like:

```
NAME                                     READY   STATUS    RESTARTS   AGE
trtis-model-resnet-big-ff95cf899-9vkgm   0/1     Pending   0          2m8s
```

