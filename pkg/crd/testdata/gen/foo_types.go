/*

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

//go:generate ../../../../.run-controller-gen.sh crd:crdVersions=v1beta1 paths=. output:dir=.
//go:generate mv bar.example.com_foos.yaml bar.example.com_foos.v1beta1.yaml
//go:generate ../../../../.run-controller-gen.sh crd:crdVersions=v1 paths=. output:dir=.

// +groupName=bar.example.com
package foo

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FooSpec struct {
	// This tests that defaulted fields are stripped for v1beta1,
	// but not for v1
	// +kubebuilder:default=fooDefaultString
	// +kubebuilder:example=fooExampleString
	DefaultedString string `json:"defaultedString"`

	// This verifies that the format annotation is applied in the CRD.
	// +kubebuilder:validation:Format=ipv4
	Address string `json:"address"`

	// This verifies that the format annotation is applied in the CRD.
	// +kubebuilder:validation:Format=ipv4
	Addresses []string `json:"addresses"`
}
type FooStatus struct{}

type Foo struct {
	// TypeMeta comments should NOT appear in the CRD spec
	metav1.TypeMeta `json:",inline"`
	// ObjectMeta comments should NOT appear in the CRD spec
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec comments SHOULD appear in the CRD spec
	Spec FooSpec `json:"spec,omitempty"`
	// Status comments SHOULD appear in the CRD spec
	Status FooStatus `json:"status,omitempty"`
}
