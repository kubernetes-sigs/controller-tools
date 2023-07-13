package multiple_versions

type InnerStruct struct {
	Foo string `json:"foo,omitempty"`
}

type OuterStruct struct {
	// +structType=atomic
	Struct InnerStruct `json:"struct,omitempty"`
}
