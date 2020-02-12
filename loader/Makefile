TEST_MODEL_REPO=/home/clive/work/gpushare/testModelRepo
# Image URL to use all building/pushing image targets
LOADER_IMG ?= seldonio/trtis-loader:0.1

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif


# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...


# Build manager binary
bin/loader: fmt vet
	go build -o bin/loader cmd/loader/main.go


.PHONY: generate_protos
generate_protos:
	git clone -b r19.12 https://github.com/NVIDIA/tensorrt-inference-server.git
	protoc -I tensorrt-inference-server/src/core --go_out=plugins=grpc:proto/trtis tensorrt-inference-server/src/core/*.proto


.PHONY: start_client
start_client:
	docker run -it --rm --net=host nvcr.io/nvidia/tensorrtserver:19.12-py3-clientsdk


start_test_trtis:
	docker run --rm --shm-size=1g --ulimit memlock=-1 --ulimit stack=67108864 -p8000:8000 -p8001:8001 -p8002:8002 -v ${TEST_MODEL_REPO}:/models nvcr.io/nvidia/tensorrtserver:19.12-py3 trtserver --model-repository=/models --model-control-mode=poll --repository-poll-secs=2

run_loader_test_model:
	./bin/loader --model-src /home/clive/work/gpushare/testModelSrc/simple --trtis-model-repo /home/clive/work/gpushare/testModelRepo

remove_loaded_test_model:
	rm -r /home/clive/work/gpushare/testModelRepo/simple


# Build the docker image
docker-build: 
	docker build . -f Dockerfile.loader -t ${LOADER_IMG}

# Push the docker image
docker-push:
	docker push ${LOADER_IMG}

docker-save:
	docker save ${LOADER_IMG} > loader.tar

kind-image-install: docker-build
	kind load docker-image ${LOADER_IMG}

