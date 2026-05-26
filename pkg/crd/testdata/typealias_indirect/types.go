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
// +versionName=v1
package typealias_indirect

import "testdata.kubebuilder.io/cronjob/typealias_middle"

type IndirectAliasSpec struct {
	Field typealias_middle.AliasedStruct `json:"field"`
}
