package repro1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testdata.kubebuilder.io/cronjob/issue1063/repro2"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,path=reproes
type Repro struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ReproSpec `json:"spec,omitempty"`
}

type ReproSpec struct {
	Reproducer repro2.MyReproAlias `json:"reproducer"`
}

// +kubebuilder:object:root=true
type ReproList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repro `json:"items"`
}
