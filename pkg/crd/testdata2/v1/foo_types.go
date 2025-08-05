// +kubebuilder:object:generate=true
// +groupName=example.com
// +versionName=v1
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type Foo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              FooSpec `json:"spec"`
}

type FooSpec struct {
	Bar any `json:"bar"`
}
