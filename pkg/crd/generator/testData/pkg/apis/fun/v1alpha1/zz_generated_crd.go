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

package v1alpha1

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CRD Generation
func getFloat(f float64) *float64 {
	return &f
}

func getInt(i int64) *int64 {
	return &i
}

var (
	// ToyCRD resource
	ToyCRD = v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "toys.fun.myk8s.io",
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   "fun.myk8s.io",
			Version: "v1alpha1",
			Names: v1beta1.CustomResourceDefinitionNames{
				Kind:   "Toy",
				Plural: "toys",
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
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"alias": v1beta1.JSONSchemaProps{
									Type: "string",
									Enum: []v1beta1.JSON{v1beta1.JSON{Raw: []byte{34, 76, 105, 111, 110, 34}}, v1beta1.JSON{Raw: []byte{34, 87, 111, 108, 102, 34}}, v1beta1.JSON{Raw: []byte{34, 68, 114, 97, 103, 111, 110, 34}}},
								},
								"bricks": v1beta1.JSONSchemaProps{
									Type:   "integer",
									Format: "int32",
								},
								"claim": v1beta1.JSONSchemaProps{
									Type:       "object",
									Properties: map[string]v1beta1.JSONSchemaProps{},
								},
								"comment": v1beta1.JSONSchemaProps{
									Type:   "string",
									Format: "byte",
								},
								"knights": v1beta1.JSONSchemaProps{
									Type:     "array",
									MaxItems: getInt(500),
									MinItems: getInt(1),
									Items: &v1beta1.JSONSchemaPropsOrArray{
										Schema: &v1beta1.JSONSchemaProps{
											Type: "string",
										},
									},
								},
								"name": v1beta1.JSONSchemaProps{
									Type:      "string",
									MaxLength: getInt(15),
									MinLength: getInt(1),
								},
								"power": v1beta1.JSONSchemaProps{
									Maximum:          getFloat(100),
									Minimum:          getFloat(1),
									ExclusiveMinimum: true,
									Type:             "number",
									Format:           "float",
								},
								"rank": v1beta1.JSONSchemaProps{
									Type:   "integer",
									Format: "int64",
									Enum:   []v1beta1.JSON{v1beta1.JSON{Raw: []byte{49}}, v1beta1.JSON{Raw: []byte{50}}, v1beta1.JSON{Raw: []byte{51}}},
								},
								"replicas": v1beta1.JSONSchemaProps{
									Type:   "integer",
									Format: "int32",
								},
								"template": v1beta1.JSONSchemaProps{
									Type:       "object",
									Properties: map[string]v1beta1.JSONSchemaProps{},
								},
								"winner": v1beta1.JSONSchemaProps{
									Type: "boolean",
								},
							},
							Required: []string{
								"rank",
								"template",
								"replicas",
							}},
						"status": v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"replicas": v1beta1.JSONSchemaProps{
									Type:   "integer",
									Format: "int32",
								},
							},
							Required: []string{
								"replicas",
							}},
					},
				},
			},
			Subresources: &v1beta1.CustomResourceSubresources{
				Status: &v1beta1.CustomResourceSubresourceStatus{},
				Scale: &v1beta1.CustomResourceSubresourceScale{
					SpecReplicasPath:   ".spec.replicas",
					StatusReplicasPath: ".status.replicas",
				},
			},
		},
	}
)
