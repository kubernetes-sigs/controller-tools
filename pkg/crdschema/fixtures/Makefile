update:
	cd ../../.. && \
	go run ./cmd/controller-gen crdschema:manifests=./pkg/crdschema/fixtures/manifests output:dir=./pkg/crdschema/fixtures/manifests paths=./pkg/crdschema/fixtures/pkg/apis/...
	cp manifests/*-example-crd.yaml expected
	egrep -v -- "- foo" < expected/kubebuilder-example-crd.yaml > manifests/kubebuilder-example-crd.yaml
	egrep -v -- "- foo" < expected/legacy-example-crd.yaml > manifests/legacy-example-crd.yaml
