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

// Package v1 contains a Foo of another API group than the Foo in ../../v1 and ../../v2.
// It is used to test that kinds with the same name in multiple API groups are kept apart
// and that the plural name of the kubebuilder:resource marker is used.
// The group is sorted before bar.example.com on purpose: the kinds of all parsed packages are
// looked up sorted by group, so without filtering by group the Foo of bar.example.com would
// end up in this group.
//
// +groupName=another.example.com
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FooSpec is the spec of Foo.
type FooSpec struct {
	// Size is the size of a Foo.
	// +k8s:controller-gen:state-metrics:gauge:name="size",help="The size of a foo."
	Size int32 `json:"size"`
}

// Foo is a test object, which has the same kind as the Foo of the group bar.example.com
// and a plural name which is not derived from the kind.
// +k8s:controller-gen:state-metrics:store
// +kubebuilder:resource:path=foobars
type Foo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FooSpec `json:"spec,omitempty"`
}
