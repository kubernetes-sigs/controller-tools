/*
Copyright The Kubernetes Authors.

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

// Package rawmessage_error has a list marker on a json.RawMessage field. A
// byte slice is a base64-encoded string, not an array, so the marker must be
// rejected.
//
// +groupName=testdata.kubebuilder.io
// +versionName=v1
package rawmessage_error

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type Extension struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ExtensionSpec `json:"spec"`
}

type ExtensionSpec struct {
	// Config is a base64-encoded string, so a list marker does not apply.
	// +listType=set
	Config json.RawMessage `json:"config"`
}
