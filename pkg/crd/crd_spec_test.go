/*
Copyright 2019 The Kubernetes Authors.

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

package crd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"sigs.k8s.io/controller-tools/pkg/crd"
)

var _ = Describe("CRD Generation", func() {
	Describe("Utilities", func() {
		Describe("MergeIdenticalVersionInfo", func() {
			It("should replace per-version schemata with a top-level schema if only one version", func() {
				spec := &apiext.CustomResourceDefinition{
					Spec: apiext.CustomResourceDefinitionSpec{
						Versions: []apiext.CustomResourceDefinitionVersion{
							{
								Name:    "v1",
								Storage: true,
								Schema: &apiext.CustomResourceValidation{
									OpenAPIV3Schema: &apiext.JSONSchemaProps{
										Required:   []string{"foo"},
										Type:       "object",
										Properties: map[string]apiext.JSONSchemaProps{"foo": apiext.JSONSchemaProps{Type: "string"}},
									},
								},
							},
						},
					},
				}
				crd.MergeIdenticalVersionInfo(spec)
				Expect(spec.Spec.Validation).To(Equal(&apiext.CustomResourceValidation{
					OpenAPIV3Schema: &apiext.JSONSchemaProps{
						Required:   []string{"foo"},
						Type:       "object",
						Properties: map[string]apiext.JSONSchemaProps{"foo": apiext.JSONSchemaProps{Type: "string"}},
					},
				}))
				Expect(spec.Spec.Versions).To(Equal([]apiext.CustomResourceDefinitionVersion{
					{Name: "v1", Storage: true},
				}))
			})
			It("should replace per-version schemata with a top-level schema if all are identical", func() {
				spec := &apiext.CustomResourceDefinition{
					Spec: apiext.CustomResourceDefinitionSpec{
						Versions: []apiext.CustomResourceDefinitionVersion{
							{
								Name: "v1",
								Schema: &apiext.CustomResourceValidation{
									OpenAPIV3Schema: &apiext.JSONSchemaProps{
										Required:   []string{"foo"},
										Type:       "object",
										Properties: map[string]apiext.JSONSchemaProps{"foo": apiext.JSONSchemaProps{Type: "string"}},
									},
								},
							},
							{
								Name:    "v2",
								Storage: true,
								Schema: &apiext.CustomResourceValidation{
									OpenAPIV3Schema: &apiext.JSONSchemaProps{
										Required:   []string{"foo"},
										Type:       "object",
										Properties: map[string]apiext.JSONSchemaProps{"foo": apiext.JSONSchemaProps{Type: "string"}},
									},
								},
							},
						},
					},
				}
				crd.MergeIdenticalVersionInfo(spec)
				Expect(spec.Spec.Validation).To(Equal(&apiext.CustomResourceValidation{
					OpenAPIV3Schema: &apiext.JSONSchemaProps{
						Required:   []string{"foo"},
						Type:       "object",
						Properties: map[string]apiext.JSONSchemaProps{"foo": apiext.JSONSchemaProps{Type: "string"}},
					},
				}))
				Expect(spec.Spec.Versions).To(Equal([]apiext.CustomResourceDefinitionVersion{
					{Name: "v1"}, {Name: "v2", Storage: true},
				}))
			})

			It("shouldn't merge different schemata", func() {
				spec := &apiext.CustomResourceDefinition{
					Spec: apiext.CustomResourceDefinitionSpec{
						Versions: []apiext.CustomResourceDefinitionVersion{
							{
								Name: "v1",
								Schema: &apiext.CustomResourceValidation{
									OpenAPIV3Schema: &apiext.JSONSchemaProps{
										Type:       "object",
										Properties: map[string]apiext.JSONSchemaProps{"foo": apiext.JSONSchemaProps{Type: "string"}},
									},
								},
							},
							{
								Name:    "v2",
								Storage: true,
								Schema: &apiext.CustomResourceValidation{
									OpenAPIV3Schema: &apiext.JSONSchemaProps{
										Required:   []string{"foo"},
										Type:       "object",
										Properties: map[string]apiext.JSONSchemaProps{"foo": apiext.JSONSchemaProps{Type: "string"}},
									},
								},
							},
						},
					},
				}
				orig := spec.DeepCopy()
				crd.MergeIdenticalVersionInfo(spec)
				Expect(spec).To(Equal(orig))
			})

			It("should replace per-version subresources with top-level subresources if only one version", func() {
				spec := &apiext.CustomResourceDefinition{
					Spec: apiext.CustomResourceDefinitionSpec{
						Versions: []apiext.CustomResourceDefinitionVersion{
							{
								Name:    "v1",
								Storage: true,
								Subresources: &apiext.CustomResourceSubresources{
									Status: &apiext.CustomResourceSubresourceStatus{},
								},
							},
						},
					},
				}

				crd.MergeIdenticalVersionInfo(spec)
				Expect(spec.Spec.Subresources).To(Equal(&apiext.CustomResourceSubresources{
					Status: &apiext.CustomResourceSubresourceStatus{},
				}))
				Expect(spec.Spec.Versions).To(Equal([]apiext.CustomResourceDefinitionVersion{
					{Name: "v1", Storage: true},
				}))
			})

			It("should replace per-version subresources with top-level subresources if all are identical", func() {
				spec := &apiext.CustomResourceDefinition{
					Spec: apiext.CustomResourceDefinitionSpec{
						Versions: []apiext.CustomResourceDefinitionVersion{
							{
								Name: "v1",
								Subresources: &apiext.CustomResourceSubresources{
									Status: &apiext.CustomResourceSubresourceStatus{},
								},
							},
							{
								Name:    "v2",
								Storage: true,
								Subresources: &apiext.CustomResourceSubresources{
									Status: &apiext.CustomResourceSubresourceStatus{},
								},
							},
						},
					},
				}

				crd.MergeIdenticalVersionInfo(spec)
				Expect(spec.Spec.Subresources).To(Equal(&apiext.CustomResourceSubresources{
					Status: &apiext.CustomResourceSubresourceStatus{},
				}))
				Expect(spec.Spec.Versions).To(Equal([]apiext.CustomResourceDefinitionVersion{
					{Name: "v1"}, {Name: "v2", Storage: true},
				}))
			})

			It("shouldn't merge different subresources", func() {
				spec := &apiext.CustomResourceDefinition{
					Spec: apiext.CustomResourceDefinitionSpec{
						Versions: []apiext.CustomResourceDefinitionVersion{
							{
								Name: "v1",
								Subresources: &apiext.CustomResourceSubresources{
									Status: &apiext.CustomResourceSubresourceStatus{},
								},
							},
							{
								Name:    "v2",
								Storage: true,
							},
						},
					},
				}
				orig := spec.DeepCopy()
				crd.MergeIdenticalVersionInfo(spec)
				Expect(spec).To(Equal(orig))
			})

			It("should replace per-version printer columns with top-level printer columns if only one version", func() {
				spec := &apiext.CustomResourceDefinition{
					Spec: apiext.CustomResourceDefinitionSpec{
						Versions: []apiext.CustomResourceDefinitionVersion{
							{
								Name:    "v1",
								Storage: true,
								AdditionalPrinterColumns: []apiext.CustomResourceColumnDefinition{
									{Name: "Cheddar", JSONPath: ".spec.cheddar"},
									{Name: "Parmesan", JSONPath: ".status.parmesan"},
								},
							},
						},
					},
				}

				crd.MergeIdenticalVersionInfo(spec)
				Expect(spec.Spec.AdditionalPrinterColumns).To(Equal([]apiext.CustomResourceColumnDefinition{
					{Name: "Cheddar", JSONPath: ".spec.cheddar"},
					{Name: "Parmesan", JSONPath: ".status.parmesan"},
				}))
				Expect(spec.Spec.Versions).To(Equal([]apiext.CustomResourceDefinitionVersion{
					{Name: "v1", Storage: true},
				}))
			})

			It("should replace per-version printer columns with top-level printer columns if all are identical", func() {
				spec := &apiext.CustomResourceDefinition{
					Spec: apiext.CustomResourceDefinitionSpec{
						Versions: []apiext.CustomResourceDefinitionVersion{
							{
								Name: "v1",
								AdditionalPrinterColumns: []apiext.CustomResourceColumnDefinition{
									{Name: "Cheddar", JSONPath: ".spec.cheddar"},
									{Name: "Parmesan", JSONPath: ".status.parmesan"},
								},
							},
							{
								Name:    "v2",
								Storage: true,
								AdditionalPrinterColumns: []apiext.CustomResourceColumnDefinition{
									{Name: "Cheddar", JSONPath: ".spec.cheddar"},
									{Name: "Parmesan", JSONPath: ".status.parmesan"},
								},
							},
						},
					},
				}

				crd.MergeIdenticalVersionInfo(spec)
				Expect(spec.Spec.AdditionalPrinterColumns).To(Equal([]apiext.CustomResourceColumnDefinition{
					{Name: "Cheddar", JSONPath: ".spec.cheddar"},
					{Name: "Parmesan", JSONPath: ".status.parmesan"},
				}))
				Expect(spec.Spec.Versions).To(Equal([]apiext.CustomResourceDefinitionVersion{
					{Name: "v1"}, {Name: "v2", Storage: true},
				}))
			})

			It("shouldn't merge different printer columns", func() {
				spec := &apiext.CustomResourceDefinition{
					Spec: apiext.CustomResourceDefinitionSpec{
						Versions: []apiext.CustomResourceDefinitionVersion{
							{
								Name: "v1",
								AdditionalPrinterColumns: []apiext.CustomResourceColumnDefinition{
									{Name: "Cheddar", JSONPath: ".spec.cheddar"},
									{Name: "Parmesan", JSONPath: ".status.parmesan"},
								},
							},
							{
								Name:    "v2",
								Storage: true,
							},
						},
					},
				}
				orig := spec.DeepCopy()
				crd.MergeIdenticalVersionInfo(spec)
				Expect(spec).To(Equal(orig))
			})
		})
	})
})
