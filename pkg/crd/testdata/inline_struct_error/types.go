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

// +groupName=testdata.kubebuilder.io
package inlinestructerror

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// InlineStruct tests that inline struct literals are properly rejected.
// This should fail with a clear error message instead of panicking.
type InlineStruct struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InlineStructSpec `json:"spec,omitempty"`
}

// InlineStructSpec contains an inline struct field which should be rejected.
type InlineStructSpec struct {
	// TestStruct is an inline struct literal which is not supported.
	TestStruct struct {
		Test1 string `json:"test1,omitempty"`
	} `json:"test_struct"`
}
