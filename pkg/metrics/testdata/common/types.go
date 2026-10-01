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

// Package common contains types which are used by multiple versions of Foo.
// It is used to test that metrics markers on types from another package get considered.
package common

// Settings are settings of a Foo. It gets used as named field.
type Settings struct {
	// Replicas is the number of replicas.
	// +k8s:controller-gen:metrics:gauge:name="settings_replicas",help="The number of replicas of a foo."
	Replicas int32 `json:"replicas"`
}

// Limits are limits of a Foo. It gets used as inline field.
type Limits struct {
	// MaxItems is the maximum number of items.
	// +k8s:controller-gen:metrics:gauge:name="limits_max_items",help="The maximum number of items of a foo."
	MaxItems int32 `json:"maxItems"`
}
