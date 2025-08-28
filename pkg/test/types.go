// +groupName=test.example.com
package test

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Example demonstrates basic feature gate usage
// +kubebuilder:object:root=true
type Example struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExampleSpec   `json:"spec,omitempty"`
	Status ExampleStatus `json:"status,omitempty"`
}

// ExampleSpec defines the desired state of Example
type ExampleSpec struct {
	// AlphaField is only included when alpha feature gate is enabled
	// +kubebuilder:featuregate=alpha
	AlphaField string `json:"alphaField,omitempty"`

	// BetaField is only included when beta feature gate is enabled
	// +kubebuilder:featuregate=beta
	BetaField string `json:"betaField,omitempty"`

	// StableField is always included (no feature gate)
	StableField string `json:"stableField,omitempty"`

	// ComplexFeatureField requires both alpha and beta gates
	// +kubebuilder:featuregate=alpha&beta
	ComplexFeatureField string `json:"complexFeatureField,omitempty"`

	// RegularValidationField has normal validation (for comparison)
	// +kubebuilder:validation:Pattern="^[a-z]+$"
	RegularValidationField string `json:"regularValidationField,omitempty"`
}

// ExampleStatus defines the observed state of Example
type ExampleStatus struct {
	// Ready indicates if the resource is ready
	Ready bool `json:"ready,omitempty"`
}

// ExampleList contains a list of Example
// +kubebuilder:object:root=true
type ExampleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Example `json:"items"`
}
