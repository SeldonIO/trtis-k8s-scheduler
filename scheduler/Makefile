# Image URL to use all building/pushing image targets
SCHEDULER_IMG ?= seldonio/trtis-scheduler:0.1

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
bin/scheduler: fmt vet
	go build -o bin/scheduler cmd/scheduler/main.go


# Build the docker image
docker-build: 
	docker build . -f Dockerfile.scheduler -t ${SCHEDULER_IMG}

# Push the docker image
docker-push:
	docker push ${SCHEDULER_IMG}

docker-save:
	docker save ${SCHEDULER_IMG} > scheduler.tar

kind-image-install: docker-build
	kind load docker-image ${SCHEDULER_IMG}
