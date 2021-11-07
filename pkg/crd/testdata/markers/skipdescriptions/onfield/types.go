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

//go:generate ../../../../../../.run-controller-gen.sh crd:crdVersions=v1 paths=. output:dir=.

// +groupName=example.com
// +versionName=v1
package onfield

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestResource type description.
type TestResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// HavingDescription field description.
	HavingDescription HavingDescription `json:"havingDescription"`

	// NotHavingDescription field description.
	// +kubebuilder:skipDescriptions
	NotHavingDescription NotHavingDescription `json:"notHavingDescription"`
}

// HavingDescription type description.
type HavingDescription struct {
	// Foo field description.
	Foo string `json:"foo"`
}

// NotHavingDescription type description.
type NotHavingDescription struct {
	// Bar field description.
	Bar string `json:"bar"`

	// NotHavingDescriptionEmbedded field description.
	NotHavingDescriptionEmbedded NotHavingDescriptionEmbedded `json:"notHavingDescriptionEmbedded"`
}

// NotHavingDescriptionEmbedded type description.
type NotHavingDescriptionEmbedded struct {
	// Baz field description.
	Baz string `json:"baz"`
}
