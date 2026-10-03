/*
Copyright 2026 The Kubernetes Authors.

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

// +groupName=bar.example.com
package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-tools/pkg/metrics/testdata/common"
)

// FooSpec is the spec of Foo.
type FooSpec struct {
	// SomeString is a string.
	SomeString string `json:"someString"`

	// Settings are the settings of a Foo.
	// This uses a type from another package.
	Settings common.Settings `json:"settings"`

	// Limits are inlined, no additional path element must get added for them.
	common.Limits `json:",inline"`
}

// Foo is a test object.
// +k8s:controller-gen:metrics:store:namePrefix="foo"
// +k8s:controller-gen:metrics:label:name="name",JSONPath=".metadata.name"
// +k8s:controller-gen:metrics:gauge:name="created",JSONPath=".metadata.creationTimestamp",help="Unix creation timestamp."
// +kubebuilder:storageversion
type Foo struct {
	// TypeMeta comments should NOT appear in the CRD spec
	metav1.TypeMeta `json:",inline"`
	// ObjectMeta comments should NOT appear in the CRD spec
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec comments SHOULD appear in the CRD spec
	Spec FooSpec `json:"spec,omitempty"`
}
