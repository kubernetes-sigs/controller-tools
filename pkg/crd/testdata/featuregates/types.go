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

//go:generate ../../../../.run-controller-gen.sh crd:featureGates=alpha=true paths=. output:dir=./output_alpha
//go:generate ../../../../.run-controller-gen.sh crd:featureGates=beta=true paths=. output:dir=./output_beta  
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
	// +kubebuilder:feature-gate=alpha
	AlphaFeature *string `json:"alphaFeature,omitempty"`

	// Beta-gated field - only included when beta gate is enabled  
	// +kubebuilder:feature-gate=beta
	BetaFeature *string `json:"betaFeature,omitempty"`
}

// FeatureGateTestStatus defines the observed state with feature-gated fields
type FeatureGateTestStatus struct {
	// Standard status field
	Ready bool `json:"ready"`

	// Alpha-gated status field
	// +kubebuilder:feature-gate=alpha
	AlphaStatus *string `json:"alphaStatus,omitempty"`

	// Beta-gated status field
	// +kubebuilder:feature-gate=beta
	BetaStatus *string `json:"betaStatus,omitempty"`
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
