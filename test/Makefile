
all: test manager

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager sigs.k8s.io/controller-tools/test/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/rbac
	kubectl apply -f config/crds
	kubectl apply -f config/manager

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go build -o /tmp/controller-gen sigs.k8s.io/controller-tools/cmd/controller-gen && /tmp/controller-gen all

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
docker-build: test
	docker build . -t controller:latest

# Push the docker image
docker-push:
	docker push controller:latest
