/*
Copyright 2018 The Kubernetes authors.

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

package v1beta1

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CRD Generation
// func getFloat(f float64) *float64 {
// 	return &f
// }

// func getInt(i int64) *int64 {
// 	return &i
// }

var (
	// FrigateCRD resource
	FrigateCRD = v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "frigates.ship.testproject.org",
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   "ship.testproject.org",
			Version: "v1beta1",
			Names: v1beta1.CustomResourceDefinitionNames{
				Kind:   "Frigate",
				Plural: "frigates",
			},
			Scope: "Namespaced",
			Validation: &v1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &v1beta1.JSONSchemaProps{
					Properties: map[string]v1beta1.JSONSchemaProps{
						"apiVersion": v1beta1.JSONSchemaProps{
							Type: "string",
						},
						"kind": v1beta1.JSONSchemaProps{
							Type: "string",
						},
						"metadata": v1beta1.JSONSchemaProps{
							Type: "object",
						},
						"spec": v1beta1.JSONSchemaProps{
							Type:       "object",
							Properties: map[string]v1beta1.JSONSchemaProps{},
						},
						"status": v1beta1.JSONSchemaProps{
							Type:       "object",
							Properties: map[string]v1beta1.JSONSchemaProps{},
						},
					},
				},
			},
			Subresources: &v1beta1.CustomResourceSubresources{},
		},
	}
)
