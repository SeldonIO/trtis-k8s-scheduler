
gpu_daemonset:
	kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/container-engine-accelerators/master/nvidia-driver-installer/cos/daemonset-preloaded.yaml


# Needs a disk names nfs-disk in same region as cluster
create-nfs:
	kubectl apply -f pvc-nfs.yaml
	kubectl apply -f svc-nfs.yaml
	kubectl apply -f deployment-nfs.yaml
	kubectl rollout status deploy/nfs-server

create-trtis:
	kubectl create clusterrolebinding default-cluster-admin --clusterrole=cluster-admin --serviceaccount=default:default
	kubectl apply -f daemonset_trtis.yaml
	kubectl rollout status daemonset/trtis

create-scheduler:
	kubectl apply -f trtis-scheduler-rbac.yaml
	kubectl apply -f deployment-scheduler.yaml
	kubectl rollout status deploy/trtis-scheduler

create-loadbalancer:
	kubectl apply -f svc-trtis-demo.yaml

teardown-demo:
	kubectl delete -f daemonset_trtis.yaml
	kubectl delete clusterrolebinding default-cluster-admin 
	kubectl delete -f deployment-scheduler.yaml
	kubectl delete -f trtis-scheduler-rbac.yaml
	kubectl delete -f svc-nfs.yaml
	kubectl delete -f deployment-nfs.yaml
	kubectl delete -f pvc-nfs.yaml



deploy-simple-model:
	kubectl apply -f deployment-model-simple.yaml
	kubectl rollout status deploy/trtis-model-simple

undeploy-simple-model:
	kubectl delete -f deployment-model-simple.yaml

deploy-resnet-model:
	kubectl apply -f deployment-model-resnet.yaml
	kubectl rollout status deploy/trtis-model-resnet

undeploy-resnet-model:
	kubectl delete -f deployment-model-resnet.yaml

deploy-resnet-model-toobig:
	kubectl apply -f deployment-model-resnet-toobig.yaml

undeploy-resnet-model-toobig:
	kubectl delete -f deployment-model-resnet-toobig.yaml




port-forward-proxy-simple:
	kubectl port-forward $(kubectl get pods -n default -l app=trtis-model-simple -o jsonpath='{.items[0].metadata.name}') -n default 9000:9000

# example tests - need external ip for loadbalancer
# simple_client -u 35.241.132.4:8000
# image_client -u 35.241.132.4:8000 -m resnet50_netdef -s INCEPTION images/mug.jpg
run-trtis-client:
	docker run -it --rm --net=host nvcr.io/nvidia/tensorrtserver:20.01-py3-clientsdk
