
all: test manager

test: generate
	go test sigs.k8s.io/controller-tools/test/pkg/... sigs.k8s.io/controller-tools/test/cmd/...

manager: generate
	go build -o bin/manager sigs.k8s.io/controller-tools/test/cmd/manager

generate: 
	go generate sigs.k8s.io/controller-tools/test/pkg/... sigs.k8s.io/controller-tools/test/cmd/...
