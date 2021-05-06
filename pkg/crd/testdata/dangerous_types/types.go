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

//go:generate ../../../../.run-controller-gen.sh crd:crdVersions=v1 paths=. output:dir=.

// +groupName=dangerous.example.com
// +versionName=v1
package dangerous_types

type DangerousType struct {

	f32 float32 `json:"f32,omitempty"`
	f64 float64 `json:"f64,omitempty"`

	// Checks that nested maps work
	NestedMap map[string]map[string]string `json:"nestedMap,omitempty"`

	// Checks that multiply-nested maps work
	NestedNestedMap map[string]map[string]map[string]string `json:"nestedNestedMap,omitempty"`

	// Checks that maps containing types that contain maps work
	ContainsNestedMapMap map[string]ContainsNestedMap `json:"nestedMapInStruct,omitempty"`
}

type ContainsNestedMap struct {
	InnerMap map[string]string `json:"innerMap,omitempty"`
}
