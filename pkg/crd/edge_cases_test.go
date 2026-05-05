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

package crd

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ = Describe("MergeNestedDefaults - Edge Cases", func() {
	Context("when handling arrays", func() {
		It("should not merge into array defaults", func() {
			By("creating a schema with array type")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "array",
				Default: &apiextensionsv1.JSON{Raw: []byte(`[]`)},
				Items: &apiextensionsv1.JSONSchemaPropsOrArray{
					Schema: &apiextensionsv1.JSONSchemaProps{
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"item"`)},
					},
				},
			}

			By("attempting to merge nested defaults")
			mergeNestedDefaults(schema)

			By("verifying array default remains empty")
			var result []any
			err := json.Unmarshal(schema.Default.Raw, &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(0))
		})
	})

	Context("when handling mixed property types", func() {
		It("should merge all types correctly", func() {
			By("creating a schema with various property types")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"stringField": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"hello"`)},
					},
					"intField": {
						Type:    "integer",
						Default: &apiextensionsv1.JSON{Raw: []byte(`42`)},
					},
					"boolField": {
						Type:    "boolean",
						Default: &apiextensionsv1.JSON{Raw: []byte(`true`)},
					},
					"arrayField": {
						Type:    "array",
						Default: &apiextensionsv1.JSON{Raw: []byte(`["a","b"]`)},
					},
					"objectField": {
						Type:    "object",
						Default: &apiextensionsv1.JSON{Raw: []byte(`{"nested":"value"}`)},
					},
				},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying all types merged correctly")
			var result map[string]any
			err := json.Unmarshal(schema.Default.Raw, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result["stringField"]).To(Equal("hello"))
			Expect(result["intField"]).To(Equal(float64(42)))
			Expect(result["boolField"]).To(Equal(true))

			arr, ok := result["arrayField"].([]any)
			Expect(ok).To(BeTrue())
			Expect(arr).To(HaveLen(2))

			obj, ok := result["objectField"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(obj["nested"]).To(Equal("value"))
		})
	})

	Context("when properties have mixed default configurations", func() {
		It("should only merge properties with defaults", func() {
			By("creating a schema with some properties having defaults and some without")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"withDefault": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"value"`)},
					},
					"withoutDefault": {
						Type: "string",
					},
					"required": {
						Type: "string",
					},
				},
				Required: []string{"required"},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying only properties with defaults are merged")
			var result map[string]any
			err := json.Unmarshal(schema.Default.Raw, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result["withDefault"]).To(Equal("value"))
			Expect(result).NotTo(HaveKey("withoutDefault"))
			Expect(result).NotTo(HaveKey("required"))
		})
	})

	Context("when handling deeply nested objects", func() {
		It("should merge all levels correctly", func() {
			By("creating a schema with multiple nesting levels")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"level1": {
						Type:    "object",
						Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"level2": {
								Type:    "object",
								Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
								Properties: map[string]apiextensionsv1.JSONSchemaProps{
									"level3": {
										Type:    "string",
										Default: &apiextensionsv1.JSON{Raw: []byte(`"deep"`)},
									},
								},
							},
						},
					},
				},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying deep nesting was merged correctly")
			var level2 map[string]any
			err := json.Unmarshal(schema.Properties["level1"].Properties["level2"].Default.Raw, &level2)
			Expect(err).NotTo(HaveOccurred())
			Expect(level2["level3"]).To(Equal("deep"))
		})
	})

	Context("when parent has existing non-empty defaults", func() {
		It("should preserve them unchanged", func() {
			By("creating a schema with non-empty parent default")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{"existing":"value","another":123}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"foo": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"bar"`)},
					},
				},
			}

			By("saving the original default")
			original := string(schema.Default.Raw)

			By("attempting to merge nested defaults")
			mergeNestedDefaults(schema)

			By("verifying parent default was not modified")
			Expect(string(schema.Default.Raw)).To(Equal(original))
		})
	})

	Context("when handling special default values", func() {
		It("should correctly handle empty strings and null", func() {
			By("creating a schema with empty string and null defaults")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"emptyString": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`""`)},
					},
					"nullString": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`null`)},
					},
				},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying empty string and null are merged correctly")
			var result map[string]any
			err := json.Unmarshal(schema.Default.Raw, &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["emptyString"]).To(Equal(""))
			Expect(result["nullString"]).To(BeNil())
		})
	})

	Context("when object has no property defaults", func() {
		It("should remain as empty object", func() {
			By("creating a schema with properties but no defaults")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"noDefault1": {Type: "string"},
					"noDefault2": {Type: "integer"},
				},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying parent default remains empty")
			var result map[string]any
			err := json.Unmarshal(schema.Default.Raw, &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(0))
		})
	})

	Context("when handling zero values", func() {
		It("should treat zero values as valid defaults", func() {
			By("creating a schema with zero value defaults")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"zeroInt": {
						Type:    "integer",
						Default: &apiextensionsv1.JSON{Raw: []byte(`0`)},
					},
					"falseBool": {
						Type:    "boolean",
						Default: &apiextensionsv1.JSON{Raw: []byte(`false`)},
					},
					"emptyArray": {
						Type:    "array",
						Default: &apiextensionsv1.JSON{Raw: []byte(`[]`)},
					},
				},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying zero values are merged correctly")
			var result map[string]any
			err := json.Unmarshal(schema.Default.Raw, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result["zeroInt"]).To(Equal(float64(0)))
			Expect(result["falseBool"]).To(Equal(false))

			arr, ok := result["emptyArray"].([]any)
			Expect(ok).To(BeTrue())
			Expect(arr).To(HaveLen(0))
		})
	})
})
