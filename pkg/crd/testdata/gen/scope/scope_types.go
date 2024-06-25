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

//go:generate ../../../../../.run-controller-gen.sh crd:crdVersions=v1 paths=. output:dir=.

// +groupName=scope.example.com
package scope

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScopeSpec struct {
	// This field appears regardless of scope.
	AlwaysExists string `json:"alwaysExists"`
	// This field only appears for cluster-scoped objects.
	// +kubebuilder:field:scope=Cluster
	ExistsInCluster string `json:"existsInCluster"`
	// This field only appears for namespace-scoped objects.
	// +kubebuilder:field:scope=Namespaced
	ExistsInNamespaced string `json:"existsInNamespaced"`
}

type NamespacedFoo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ScopeSpec `json:"spec,omitempty"`
}

// +kubebuilder:resource:scope=Cluster
type ClusterFoo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ScopeSpec `json:"spec,omitempty"`
}
