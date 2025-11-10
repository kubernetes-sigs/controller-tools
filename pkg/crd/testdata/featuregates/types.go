/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//go:generate ../../../../.run-controller-gen.sh crd:featureGates="alpha=true" paths=. output:dir=./output_alpha
//go:generate ../../../../.run-controller-gen.sh crd:featureGates="beta=true" paths=. output:dir=./output_beta
//go:generate ../../../../.run-controller-gen.sh crd:featureGates="alpha=true,beta=true" paths=. output:dir=./output_both
//go:generate ../../../../.run-controller-gen.sh crd:featureGates="gamma=true" paths=. output:dir=./output_gamma
//go:generate ../../../../.run-controller-gen.sh crd:featureGates="alpha=true,beta=true,gamma=true" paths=. output:dir=./output_all
//go:generate ../../../../.run-controller-gen.sh crd paths=. output:dir=./output_none

package featuregates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FeatureGateTestSpec defines the desired state with feature-gated fields
type FeatureGateTestSpec struct {
	// Standard field - always included
	Name string `json:"name"`

	// Alpha-gated field - only included when alpha gate is enabled
	// +kubebuilder:featuregate=alpha
	AlphaFeature *string `json:"alphaFeature,omitempty"`

	// Beta-gated field - only included when beta gate is enabled
	// +kubebuilder:featuregate=beta
	BetaFeature *string `json:"betaFeature,omitempty"`

	// OR-gated field - included when either alpha OR beta gate is enabled
	// +kubebuilder:featuregate=alpha|beta
	OrFeature *string `json:"orFeature,omitempty"`

	// AND-gated field - included only when both alpha AND beta gates are enabled
	// +kubebuilder:featuregate=alpha&beta
	AndFeature *string `json:"andFeature,omitempty"`

	// Complex precedence field - included when (alpha AND beta) OR gamma is enabled
	// +kubebuilder:featuregate=(alpha&beta)|gamma
	ComplexOrFeature *string `json:"complexOrFeature,omitempty"`

	// Complex precedence field - included when (alpha OR beta) AND gamma is enabled
	// +kubebuilder:featuregate=(alpha|beta)&gamma
	ComplexAndFeature *string `json:"complexAndFeature,omitempty"`
}

// FeatureGateTestStatus defines the observed state with feature-gated fields
type FeatureGateTestStatus struct {
	// Standard status field
	Ready bool `json:"ready"`

	// Alpha-gated status field
	// +kubebuilder:featuregate=alpha
	AlphaStatus *string `json:"alphaStatus,omitempty"`

	// Beta-gated status field
	// +kubebuilder:featuregate=beta
	BetaStatus *string `json:"betaStatus,omitempty"`

	// OR-gated status field - included when either alpha OR beta gate is enabled
	// +kubebuilder:featuregate=alpha|beta
	OrStatus *string `json:"orStatus,omitempty"`

	// AND-gated status field - included only when both alpha AND beta gates are enabled
	// +kubebuilder:featuregate=alpha&beta
	AndStatus *string `json:"andStatus,omitempty"`

	// Complex precedence status field - included when (alpha AND beta) OR gamma is enabled
	// +kubebuilder:featuregate=(alpha&beta)|gamma
	ComplexOrStatus *string `json:"complexOrStatus,omitempty"`

	// Complex precedence status field - included when (alpha OR beta) AND gamma is enabled
	// +kubebuilder:featuregate=(alpha|beta)&gamma
	ComplexAndStatus *string `json:"complexAndStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// FeatureGateTest is the Schema for testing feature gates
type FeatureGateTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FeatureGateTestSpec   `json:"spec,omitempty"`
	Status FeatureGateTestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FeatureGateTestList contains a list of FeatureGateTest
type FeatureGateTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FeatureGateTest `json:"items"`
}
