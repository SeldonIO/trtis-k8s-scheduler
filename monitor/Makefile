# Image URL to use all building/pushing image targets
MONITOR_IMG ?= seldonio/trtis-monitor:0.1


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
bin/monitor: fmt vet
	go build -o bin/monitor cmd/monitor/main.go


# Build the docker image
docker-build: 
	docker build . -f Dockerfile.monitor -t ${MONITOR_IMG}

# Push the docker image
docker-push:
	docker push ${MONITOR_IMG}

docker-save:
	docker save ${MONITOR_IMG} > monitor.tar

kind-image-install: docker-build
	kind load docker-image ${MONITOR_IMG}
