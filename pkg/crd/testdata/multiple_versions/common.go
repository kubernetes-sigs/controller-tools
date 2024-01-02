//go:generate ../../../../.run-controller-gen.sh crd paths=./v1beta1;./v1beta2 output:dir=.

package multiple_versions

type InnerStruct struct {
	Foo string `json:"foo,omitempty"`
}

type OuterStruct struct {
	// +structType=atomic
	Struct InnerStruct `json:"struct,omitempty"`
}
