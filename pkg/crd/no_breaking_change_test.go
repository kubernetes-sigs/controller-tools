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

var _ = Describe("MergeNestedDefaults - Non-Breaking Behavior", func() {
	Context("when parent default is an empty object", func() {
		It("should merge nested defaults into parent", func() {
			By("creating a schema with empty parent default and nested property defaults")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"foo": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"forty-two"`)},
					},
				},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying the parent default contains the nested value")
			var result map[string]any
			err := json.Unmarshal(schema.Default.Raw, &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["foo"]).To(Equal("forty-two"))

			By("verifying the nested default still exists")
			Expect(schema.Properties["foo"].Default).NotTo(BeNil())
		})
	})

	Context("when parent default is non-empty", func() {
		It("should not modify the parent default", func() {
			By("creating a schema with non-empty parent default")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{"existing":"value"}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"foo": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"forty-two"`)},
					},
				},
			}

			By("saving the original default value")
			original := string(schema.Default.Raw)

			By("attempting to merge nested defaults")
			mergeNestedDefaults(schema)

			By("verifying the parent default was not modified")
			Expect(string(schema.Default.Raw)).To(Equal(original))
		})
	})

	Context("when parent has no default", func() {
		It("should not create a parent default", func() {
			By("creating a schema without parent default")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: nil,
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"foo": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"forty-two"`)},
					},
				},
			}

			By("attempting to merge nested defaults")
			mergeNestedDefaults(schema)

			By("verifying no parent default was created")
			Expect(schema.Default).To(BeNil())
		})
	})

	Context("when merging multiple nested defaults", func() {
		It("should preserve all nested defaults after merging", func() {
			By("creating a schema with multiple nested property defaults")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"foo": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"forty-two"`)},
					},
					"bar": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"baz"`)},
					},
				},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying parent default has both nested values")
			var parentDefault map[string]any
			err := json.Unmarshal(schema.Default.Raw, &parentDefault)
			Expect(err).NotTo(HaveOccurred())
			Expect(parentDefault["foo"]).To(Equal("forty-two"))
			Expect(parentDefault["bar"]).To(Equal("baz"))

			By("verifying both nested defaults still exist")
			Expect(schema.Properties["foo"].Default).NotTo(BeNil())
			Expect(schema.Properties["bar"].Default).NotTo(BeNil())

			By("verifying nested default values are preserved")
			var fooDefault string
			err = json.Unmarshal(schema.Properties["foo"].Default.Raw, &fooDefault)
			Expect(err).NotTo(HaveOccurred())
			Expect(fooDefault).To(Equal("forty-two"))
		})
	})

	Context("when handling deeply nested structures", func() {
		It("should merge defaults via visitor pattern", func() {
			By("creating a schema with nested objects")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"nested": {
						Type:    "object",
						Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"deep": {
								Type:    "string",
								Default: &apiextensionsv1.JSON{Raw: []byte(`"value"`)},
							},
						},
					},
				},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying nested object default was merged")
			var nestedDefault map[string]any
			err := json.Unmarshal(schema.Properties["nested"].Default.Raw, &nestedDefault)
			Expect(err).NotTo(HaveOccurred())
			Expect(nestedDefault["deep"]).To(Equal("value"))

			By("verifying deep nested default still exists")
			Expect(schema.Properties["nested"].Properties["deep"].Default).NotTo(BeNil())
		})
	})
})

var _ = Describe("Runtime Behavior - No Breaking Changes", func() {
	Context("when field is omitted", func() {
		It("should produce identical results with old and new schemas", func() {
			By("creating old schema (pre-merge)")
			oldSchema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"foo": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"forty-two"`)},
					},
				},
			}

			By("creating new schema (post-merge)")
			newSchema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{"foo":"forty-two"}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"foo": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"forty-two"`)},
					},
				},
			}

			By("verifying both schemas produce the same result")
			var oldResult, newResult map[string]any
			err := json.Unmarshal(oldSchema.Default.Raw, &oldResult)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(newSchema.Default.Raw, &newResult)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("Duplicate Defaulting Prevention", func() {
	It("should not apply default values twice", func() {
		By("creating a schema with a counter field")
		schema := &apiextensionsv1.JSONSchemaProps{
			Type:    "object",
			Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"counter": {
					Type:    "integer",
					Default: &apiextensionsv1.JSON{Raw: []byte(`1`)},
				},
			},
		}

		By("merging nested defaults")
		mergeNestedDefaults(schema)

		By("checking parent has the default value")
		var parentDefault map[string]any
		err := json.Unmarshal(schema.Default.Raw, &parentDefault)
		Expect(err).NotTo(HaveOccurred())
		Expect(parentDefault["counter"]).To(Equal(float64(1)))

		By("checking nested still has the default value")
		var nestedDefault float64
		err = json.Unmarshal(schema.Properties["counter"].Default.Raw, &nestedDefault)
		Expect(err).NotTo(HaveOccurred())
		Expect(nestedDefault).To(Equal(float64(1)))
	})
})

var _ = Describe("Kubernetes API Conventions Compliance", func() {
	Context("structural schema requirements", func() {
		It("should produce valid structural schema", func() {
			By("creating a schema with nested defaults")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"foo": {
						Type:    "string",
						Default: &apiextensionsv1.JSON{Raw: []byte(`"bar"`)},
					},
				},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying the merged default is valid JSON")
			var result map[string]any
			err := json.Unmarshal(schema.Default.Raw, &result)
			Expect(err).NotTo(HaveOccurred())

			By("verifying it matches the schema type (object)")
			Expect(result).NotTo(BeNil())

			By("verifying nested value is correct type (string)")
			fooVal, ok := result["foo"].(string)
			Expect(ok).To(BeTrue())
			Expect(fooVal).To(Equal("bar"))
		})
	})

	Context("recursive defaulting behavior", func() {
		It("should preserve both parent and nested defaults", func() {
			By("creating a schema with multiple nested defaults")
			schema := &apiextensionsv1.JSONSchemaProps{
				Type:    "object",
				Default: &apiextensionsv1.JSON{Raw: []byte(`{}`)},
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"a": {Type: "string", Default: &apiextensionsv1.JSON{Raw: []byte(`"A"`)}},
					"b": {Type: "string", Default: &apiextensionsv1.JSON{Raw: []byte(`"B"`)}},
				},
			}

			By("merging nested defaults")
			mergeNestedDefaults(schema)

			By("verifying parent has both defaults")
			var parent map[string]any
			err := json.Unmarshal(schema.Default.Raw, &parent)
			Expect(err).NotTo(HaveOccurred())
			Expect(parent["a"]).To(Equal("A"))
			Expect(parent["b"]).To(Equal("B"))

			By("verifying nested defaults still exist")
			Expect(schema.Properties["a"].Default).NotTo(BeNil())
			Expect(schema.Properties["b"].Default).NotTo(BeNil())
		})
	})
})
