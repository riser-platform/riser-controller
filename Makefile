IMG ?= riserplatform/riser-controller:latest

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

all: manager

# Run tests
test: fmt lint
	go test ./...

# Build manager binary
manager: generate fmt lint
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config.
run: generate
	go run ./main.go -metrics-addr=localhost:8080

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crd/bases
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
      $(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

lint:
	golangci-lint run

# Outputs kube yaml for installing the controller
kube-resources:
	@kustomize build config/default/

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./api/...

# Build the docker image
docker-build:
	docker build . -t ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.1
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# compile and run unit tests on change. Always "make test" before comitting.
# requires filewatcher and gotestsum
watch:
	filewatcher gotestsum

# Note: As of go 1.13 GOSUMDB returns a 410. Disabling until we figure out why.
update-sdk:
	GOSUMDB=off go get -u github.com/riser-platform/riser/sdk
	GOSUMDB=off go get -u github.com/riser-platform/riser-server/api/v1/model
	go mod tidy