package crd

import (
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// v1AliasCRD is an alias of v1.CustomResourceDefinition.
// It has identical fields, the only difference is the json tag
// on spec.PreserveUnknownFields does not contain omitempty.
//
// The source of truth of v1.CustomResourceDefinition lives at
// https://github.com/kubernetes/apiextensions-apiserver/blob/master/pkg/apis/apiextensions/v1/types.go
//
// This is used in order to force controller-gen to generate
// v1 CRD yamls with preserveUnknownFields=false explicitly, because
// otherwise controller will omit it which becomes problematic for users
// upgrading existing CRDs from v1beta to v1
//
// See https://github.com/kubernetes-sigs/controller-tools/issues/476
type v1AliasCRD struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec describes how the user wants the resources to appear
	Spec v1AliasCRDSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	// status indicates the actual state of the CustomResourceDefinition
	// +optional
	Status apiext.CustomResourceDefinitionStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// v1AliasCRDSpec is an alias of v1.CustomResourceDefinitionSpec.
// It has identical fields, the only difference is the json tag
// on PreserveUnknownFields does not contain omitempty.
//
// The source of truth of v1.CustomResourceDefinitionSpec lives at
// https://github.com/kubernetes/apiextensions-apiserver/blob/master/pkg/apis/apiextensions/v1/types.go
type v1AliasCRDSpec struct {
	// group is the API group of the defined custom resource.
	// The custom resources are served under `/apis/<group>/...`.
	// Must match the name of the CustomResourceDefinition (in the form `<names.plural>.<group>`).
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	// names specify the resource and kind names for the custom resource.
	Names apiext.CustomResourceDefinitionNames `json:"names" protobuf:"bytes,3,opt,name=names"`
	// scope indicates whether the defined custom resource is cluster- or namespace-scoped.
	// Allowed values are `Cluster` and `Namespaced`.
	Scope apiext.ResourceScope `json:"scope" protobuf:"bytes,4,opt,name=scope,casttype=ResourceScope"`
	// versions is the list of all API versions of the defined custom resource.
	// Version names are used to compute the order in which served versions are listed in API discovery.
	// If the version string is "kube-like", it will sort above non "kube-like" version strings, which are ordered
	// lexicographically. "Kube-like" versions start with a "v", then are followed by a number (the major version),
	// then optionally the string "alpha" or "beta" and another number (the minor version). These are sorted first
	// by GA > beta > alpha (where GA is a version with no suffix such as beta or alpha), and then by comparing
	// major version, then minor version. An example sorted list of versions:
	// v10, v2, v1, v11beta2, v10beta3, v3beta1, v12alpha1, v11alpha2, foo1, foo10.
	Versions []apiext.CustomResourceDefinitionVersion `json:"versions" protobuf:"bytes,7,rep,name=versions"`

	// conversion defines conversion settings for the CRD.
	// +optional
	Conversion *apiext.CustomResourceConversion `json:"conversion,omitempty" protobuf:"bytes,9,opt,name=conversion"`

	// preserveUnknownFields indicates that object fields which are not specified
	// in the OpenAPI schema should be preserved when persisting to storage.
	// apiVersion, kind, metadata and known fields inside metadata are always preserved.
	// This field is deprecated in favor of setting `x-preserve-unknown-fields` to true in `spec.versions[*].schema.openAPIV3Schema`.
	// See https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#pruning-versus-preserving-unknown-fields for details.
	// +optional
	PreserveUnknownFields bool `json:"preserveUnknownFields" protobuf:"varint,10,opt,name=preserveUnknownFields"`
}
