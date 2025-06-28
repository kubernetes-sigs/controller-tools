/*
Copyright 2022 The Kubernetes Authors.

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
// +versionName=v1
// +kubebuilder:ac:generate=true
// +kubebuilder:ac:output:package="applyconfiguration"
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	GroupName          = "testdata.kubebuilder.io"
	GroupVersion       = schema.GroupVersion{Group: GroupName, Version: "v1"}
	SchemeGroupVersion = GroupVersion
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme exists solely to keep the old generators creating valid code
	AddToScheme = SchemeBuilder.AddToScheme
)

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	metav1.AddToGroupVersion(scheme, GroupVersion)

	return nil
}
