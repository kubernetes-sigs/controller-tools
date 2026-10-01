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

// Changes to this file or the packages below this directory may require to regenerate the
// `metrics.yaml`, `rbac.yaml` and `bar.example.com_foos.yaml` files. Otherwise the tests in
// ../generate_integration_test.go may fail.
// The below marker can be used to regenerate the files by running `hack/update-generated.sh`
// or the following command:
// $ go generate ./pkg/metrics/testdata
//go:generate sh -c "go run ../../../cmd/controller-gen crd metrics:experimental=true paths=./... output:dir=."

// +groupName=bar.example.com
package foo

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FooSpec is the spec of Foo.
type FooSpec struct {
	// SomeString is a string.
	SomeString string `json:"someString"`
}

// FooStatus is the status of Foo.
type FooStatus struct {
	// +k8s:controller-gen:metrics:stateset:name="status_condition",help="The condition of a foo.",labelName="status",value=".status",list={"True","False","Unknown"},labels={"type":".type"}
	// +k8s:controller-gen:metrics:gauge:name="status_condition_last_transition_time",help="The condition last transition time of a foo.",value=.lastTransitionTime,labels={"type":".type","status":".status"}
	Conditions []Condition `json:"conditions,omitempty"`
}

// Foo is a test object.
// +k8s:controller-gen:metrics:store:namePrefix="foo"
// +k8s:controller-gen:metrics:label:name="name",JSONPath=".metadata.name"
// +k8s:controller-gen:metrics:label:name="app_name",JSONPath=.metadata.labels.app\.kubernetes\.io/name
// +k8s:controller-gen:metrics:gauge:name="created",JSONPath=".metadata.creationTimestamp",help="Unix creation timestamp."
// +k8s:controller-gen:metrics:info:name="owner",JSONPath=".metadata.ownerReferences",help="Owner references.",labels={owner_is_controller:".controller",owner_kind:".kind",owner_name:".name",owner_uid:".uid"}
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

// Condition is a test condition.
type Condition struct {
	// Type of condition.
	Type string `json:"type"`
	// Status of condition.
	Status string `json:"status"`
	// LastTransitionTime of condition.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}
