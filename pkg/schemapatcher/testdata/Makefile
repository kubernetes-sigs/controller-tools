all:
	go run -v sigs.k8s.io/controller-tools/cmd/controller-gen schemapatch:manifests=./manifests output:dir=./expected paths=./apis/...
	egrep -v -- "- foo" < expected/kubebuilder-example-crd.yaml > manifests/kubebuilder-example-crd.yaml
	egrep -v -- "- foo" < expected/legacy-example-crd.yaml > manifests/legacy-example-crd.yaml

.PHONY: all
