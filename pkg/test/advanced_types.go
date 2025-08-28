// +groupName=advanced.example.com
package test

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AdvancedExample demonstrates advanced feature gate usage
// +kubebuilder:object:root=true
type AdvancedExample struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdvancedExampleSpec   `json:"spec,omitempty"`
	Status AdvancedExampleStatus `json:"status,omitempty"`
}

// AdvancedExampleSpec demonstrates OR and AND combinations
type AdvancedExampleSpec struct {
	// AlphaOrBetaField is included when either alpha OR beta is enabled
	// +kubebuilder:featuregate=alpha|beta
	AlphaOrBetaField string `json:"alphaOrBetaField,omitempty"`

	// AlphaAndBetaField is included when both alpha AND beta are enabled
	// +kubebuilder:featuregate=alpha&beta
	AlphaAndBetaField string `json:"alphaAndBetaField,omitempty"`

	// ComplexExpressionField uses parentheses for precedence
	// +kubebuilder:featuregate=(alpha&beta)|gamma
	ComplexExpressionField string `json:"complexExpressionField,omitempty"`

	// GammaOnlyField is included only when gamma is enabled
	// +kubebuilder:featuregate=gamma
	GammaOnlyField string `json:"gammaOnlyField,omitempty"`

	// AlwaysIncludedField has no feature gate
	AlwaysIncludedField string `json:"alwaysIncludedField,omitempty"`
}

// AdvancedExampleStatus defines the observed state
type AdvancedExampleStatus struct {
	Phase string `json:"phase,omitempty"`
}

// AdvancedExampleList contains a list of AdvancedExample
// +kubebuilder:object:root=true
type AdvancedExampleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdvancedExample `json:"items"`
}
