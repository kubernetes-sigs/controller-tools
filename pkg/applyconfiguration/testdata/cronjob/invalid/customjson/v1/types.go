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

// Package v1 has a struct type with custom JSON methods that serialize it as a
// string, for which applyconfiguration-gen generates an apply configuration
// that is serialized as an object:
// https://github.com/kubernetes/kubernetes/issues/143415
// +groupName=testdata.kubebuilder.io
// +versionName=v1
// +kubebuilder:ac:generate=true
package v1

import (
	"encoding/json"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource

type Widget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec WidgetSpec `json:"spec,omitempty"`
}

type WidgetSpec struct {
	Address  Address  `json:"address,omitempty"`
	Interval Interval `json:"interval,omitempty"`
}

// Address is serialized as "host:port", so it must be reported.
// +kubebuilder:validation:Type=string
type Address struct {
	Host string `json:"host,omitempty"`
	Port string `json:"port,omitempty"`
}

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Host + ":" + a.Port)
}

func (a *Address) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	a.Host, a.Port, _ = strings.Cut(s, ":")
	return nil
}

// Interval has a custom MarshalJSON that keeps the object form, so its apply
// configuration works and it must not be reported.
type Interval struct {
	Value time.Duration `json:"value"`
}

func (i Interval) MarshalJSON() ([]byte, error) {
	type intervalWithoutMarshaler Interval
	return json.Marshal(intervalWithoutMarshaler(i))
}
