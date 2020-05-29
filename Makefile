.PHONY: build_container build tag_release test_unit

TAG = $(shell git rev-parse --short HEAD)

REGISTRY_NAMESPACE = scwserverlessruntimes
REGISTRY_DOCKER_ENDPOINT = rg.fr-par.scw.cloud/$(REGISTRY_NAMESPACE)/
RUNTIME_DOCKER_IMAGE_NAME = $(REGISTRY_DOCKER_ENDPOINT)core
RUNTIME_DOCKER_IMAGE = $(RUNTIME_DOCKER_IMAGE_NAME):$(TAG)
GOBIN = $(shell go env GOPATH)/bin

RELEASE_TAG = v1.2.2

build:
	go build -o runtime main.go

lint:
	command $(GOBIN)/golint || (cd /tmp ; go get -u golang.org/x/lint/golint && go install -i golang.org/x/lint/golint)
	go list ./... | grep -v /vendor/ | xargs -L1 $(GOBIN)/golint -set_exit_status

build_container:
	go mod vendor
	docker build -t $(RUNTIME_DOCKER_IMAGE) .

tag_release:
	docker tag $(RUNTIME_DOCKER_IMAGE) $(RUNTIME_DOCKER_IMAGE_NAME):$(RELEASE_TAG)

test:
	go test ./...
