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

//go:generate ../../../../.run-controller-gen.sh crd paths=. output:dir=./output_none
//go:generate ../../../../.run-controller-gen.sh crd:featureGates="alpha=true" paths=. output:dir=./output_alpha
//go:generate ../../../../.run-controller-gen.sh crd:featureGates="beta=true" paths=. output:dir=./output_beta  
//go:generate ../../../../.run-controller-gen.sh crd:featureGates="alpha=true,beta=true" paths=. output:dir=./output_both
//go:generate ../../../../.run-controller-gen.sh crd:featureGates="gamma=true" paths=. output:dir=./output_gamma
//go:generate ../../../../.run-controller-gen.sh crd:featureGates="alpha=true,beta=true,gamma=true" paths=. output:dir=./output_all

package typelevelfeaturegates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AlwaysOnSpec defines a CRD that's always generated (no feature gate)
type AlwaysOnSpec struct {
	Name string `json:"name"`
}

// AlwaysOnStatus defines the observed state of AlwaysOn
type AlwaysOnStatus struct {
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespace

// AlwaysOn is always generated since it has no feature gate marker
type AlwaysOn struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlwaysOnSpec   `json:"spec,omitempty"`
	Status AlwaysOnStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlwaysOnList contains a list of AlwaysOn
type AlwaysOnList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlwaysOn `json:"items"`
}

// AlphaGatedSpec defines a CRD that's only generated when alpha gate is enabled
type AlphaGatedSpec struct {
	AlphaField string `json:"alphaField"`
}

// AlphaGatedStatus defines the observed state of AlphaGated
type AlphaGatedStatus struct {
	AlphaReady bool `json:"alphaReady"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status  
// +kubebuilder:resource:scope=Namespace
// +kubebuilder:featuregate=alpha

// AlphaGated is only generated when alpha feature gate is enabled
type AlphaGated struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlphaGatedSpec   `json:"spec,omitempty"`
	Status AlphaGatedStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlphaGatedList contains a list of AlphaGated
type AlphaGatedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlphaGated `json:"items"`
}

// BetaGatedSpec defines a CRD that's only generated when beta gate is enabled
type BetaGatedSpec struct {
	BetaField string `json:"betaField"`
}

// BetaGatedStatus defines the observed state of BetaGated
type BetaGatedStatus struct {
	BetaReady bool `json:"betaReady"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespace
// +kubebuilder:featuregate=beta

// BetaGated is only generated when beta feature gate is enabled
type BetaGated struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BetaGatedSpec   `json:"spec,omitempty"`
	Status BetaGatedStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BetaGatedList contains a list of BetaGated
type BetaGatedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BetaGated `json:"items"`
}

// OrGatedSpec defines a CRD that's generated when either alpha OR beta is enabled
type OrGatedSpec struct {
	OrField string `json:"orField"`
}

// OrGatedStatus defines the observed state of OrGated
type OrGatedStatus struct {
	OrReady bool `json:"orReady"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespace
// +kubebuilder:featuregate=alpha|beta

// OrGated is generated when either alpha OR beta feature gate is enabled
type OrGated struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrGatedSpec   `json:"spec,omitempty"`
	Status OrGatedStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OrGatedList contains a list of OrGated
type OrGatedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OrGated `json:"items"`
}

// AndGatedSpec defines a CRD that's generated when both alpha AND beta are enabled
type AndGatedSpec struct {
	AndField string `json:"andField"`
}

// AndGatedStatus defines the observed state of AndGated
type AndGatedStatus struct {
	AndReady bool `json:"andReady"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespace
// +kubebuilder:featuregate=alpha&beta

// AndGated is generated when both alpha AND beta feature gates are enabled
type AndGated struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AndGatedSpec   `json:"spec,omitempty"`
	Status AndGatedStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AndGatedList contains a list of AndGated
type AndGatedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AndGated `json:"items"`
}

// ComplexGatedSpec defines a CRD with complex precedence
type ComplexGatedSpec struct {
	ComplexField string `json:"complexField"`
}

// ComplexGatedStatus defines the observed state of ComplexGated
type ComplexGatedStatus struct {
	ComplexReady bool `json:"complexReady"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespace
// +kubebuilder:featuregate=(alpha&beta)|gamma

// ComplexGated is generated when (alpha AND beta) OR gamma is enabled
type ComplexGated struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComplexGatedSpec   `json:"spec,omitempty"`
	Status ComplexGatedStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ComplexGatedList contains a list of ComplexGated
type ComplexGatedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComplexGated `json:"items"`
}
