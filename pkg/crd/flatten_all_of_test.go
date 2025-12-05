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
	. "github.com/onsi/gomega/gstruct"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/crd"
)

type fakeErrRecorder struct {
	Errors []error
}

func (f *fakeErrRecorder) AddError(err error) {
	f.Errors = append(f.Errors, err)
}
func (f *fakeErrRecorder) FirstError() error {
	if len(f.Errors) == 0 {
		return nil
	}
	return f.Errors[0]
}

var _ = Describe("AllOf Flattening", func() {
	var errRec *fakeErrRecorder

	BeforeEach(func() { errRec = &fakeErrRecorder{} })

	Context("for special types that make AllOf non-structural", func() {
		It("should consider the whole field to be Nullable if at least one AllOf clause is Nullable", func() {
			By("flattening a schema with at one branch set as Nullable")
			original := &apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"multiNullable": {
						AllOf: []apiextensionsv1.JSONSchemaProps{
							{Nullable: true}, {Nullable: false}, {Nullable: false},
						},
					},
				},
			}
			flattened := crd.FlattenEmbedded(original, errRec)
			Expect(errRec.FirstError()).NotTo(HaveOccurred())

			By("ensuring that the result has no branches and is nullable")
			Expect(flattened).To(Equal(&apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"multiNullable": {Nullable: true},
				},
			}))
		})

		It("should consider the field not to be Nullable if no AllOf clauses are Nullable", func() {
			By("flattening a schema with at no branches set as Nullable")
			original := &apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"multiNullable": {
						AllOf: []apiextensionsv1.JSONSchemaProps{
							{Nullable: false}, {Nullable: false}, {Nullable: false},
						},
					},
				},
			}
			flattened := crd.FlattenEmbedded(original, errRec)
			Expect(errRec.FirstError()).NotTo(HaveOccurred())

			By("ensuring that the result has no branches and is not nullable")
			Expect(flattened).To(Equal(&apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"multiNullable": {Nullable: false},
				},
			}))
		})

		It("should ignore AdditionalProperties with no schema", func() {
			By("flattening a schema with one branch having non-schema AdditionalProperties")
			original := apiextensionsv1.JSONSchemaProps{
				AllOf: []apiextensionsv1.JSONSchemaProps{
					{AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{ /* make sure we set a nil schema */ }},
					{AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}}},
					{AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Allows: true}},
				},
			}
			flattened := crd.FlattenEmbedded(&original, errRec)
			Expect(errRec.FirstError()).NotTo(HaveOccurred())

			By("checking that the flattened version contains just the schema")
			Expect(flattened).To(Equal(&apiextensionsv1.JSONSchemaProps{
				AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
			}))
		})

		It("should attempt to collapse AdditionalProperties to non-AllOf per the normal rules when possible", func() {
			By("flattening a schema with some conflicting and some non-conflicting AdditionalProperties branches")
			defSeven := int64(7)
			defOne := int64(1)
			original := &apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"multiAdditionalProps": {
						AllOf: []apiextensionsv1.JSONSchemaProps{
							{
								AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{
									Nullable:  true,
									MaxLength: &defSeven,
								}},
							},
							{
								AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{
									Nullable: false,
									Type:     "string",
									Pattern:  "^[abc]$",
								}},
							},
							{
								AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{
									Type:      "string",
									Pattern:   "^[abcdef]$",
									MinLength: &defOne,
								}},
							},
						},
					},
				},
			}
			flattened := crd.FlattenEmbedded(original, errRec)
			Expect(errRec.FirstError()).NotTo(HaveOccurred())

			By("ensuring that the result has the minimal set of AllOf branches required, pushed inside AdditionalProperites")
			Expect(flattened).To(Equal(&apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"multiAdditionalProps": {
						AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{Schema: &apiextensionsv1.JSONSchemaProps{
							Nullable:  true,
							MaxLength: &defSeven,
							MinLength: &defOne,
							Type:      "string",
							AllOf: []apiextensionsv1.JSONSchemaProps{
								{Pattern: "^[abc]$"}, {Pattern: "^[abcdef]$"},
							},
						}},
					},
				},
			}))
		})

		It("should error out if Type values conflict", func() {
			By("flattening a schema with a single property with two different types")
			crd.FlattenEmbedded(&apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"multiType": {AllOf: []apiextensionsv1.JSONSchemaProps{{Type: "string"}, {Type: "int"}}},
				},
			}, errRec)

			By("ensuring that an error was recorded")
			Expect(errRec.FirstError()).To(HaveOccurred())
		})

		It("should merge Required fields, deduplicating", func() {
			By("flattening a schema with multiple required fields, some duplicate across branches")
			original := &apiextensionsv1.JSONSchemaProps{
				AllOf: []apiextensionsv1.JSONSchemaProps{
					{Required: []string{"foo", "bar"}},
					{Required: []string{"quux", "cheddar"}},
					{Required: []string{"bar", "baz"}},
					{Required: []string{"cheddar"}},
				},
			}
			flattened := crd.FlattenEmbedded(original, errRec)
			Expect(errRec.FirstError()).NotTo(HaveOccurred())

			By("ensuring that the result lists all required fields once, with no branches")
			Expect(flattened).To(PointTo(MatchFields(IgnoreExtras, Fields{
				// use gstruct to avoid relying on map ordering
				"Required": ConsistOf("foo", "bar", "quux", "cheddar", "baz"),
				"AllOf":    BeNil(),
			})))
		})

		It("should merge Properties when possible, pushing AllOf inside Properties when not possible", func() {
			By("flattening a schema with some conflicting and some non-conflicting Properties branches")
			defSeven := float64(7)
			defEight := float64(8)
			defOne := int64(1)
			original := &apiextensionsv1.JSONSchemaProps{
				AllOf: []apiextensionsv1.JSONSchemaProps{
					{
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"nonConflicting":    {Type: "string"},
							"conflicting1":      {Type: "string", Format: "date-time"},
							"nonConflictingDup": {Type: "bool"},
						},
					},
					{
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"conflicting1": {Type: "string", MinLength: &defOne},
							"conflicting2": {Type: "int", MultipleOf: &defSeven},
						},
					},
					{
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"conflicting2":      {Type: "int", MultipleOf: &defEight},
							"nonConflictingDup": {Type: "bool"},
						},
					},
				},
			}
			flattened := crd.FlattenEmbedded(original, errRec)
			Expect(errRec.FirstError()).NotTo(HaveOccurred())

			By("ensuring that the result has the minimal set of AllOf branches required, pushed inside Properties")
			Expect(flattened).To(Equal(&apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"nonConflicting":    {Type: "string"},
					"nonConflictingDup": {Type: "bool"},
					"conflicting1": {
						Type:      "string",
						Format:    "date-time",
						MinLength: &defOne,
					},
					"conflicting2": {
						Type:  "int",
						AllOf: []apiextensionsv1.JSONSchemaProps{{MultipleOf: &defSeven}, {MultipleOf: &defEight}},
					},
				},
			}))
		})

		It("should merge XValidation fields", func() {
			By("flattening a schema with multiple validation fields")
			original := &apiextensionsv1.JSONSchemaProps{
				AllOf: []apiextensionsv1.JSONSchemaProps{
					{XValidations: apiextensionsv1.ValidationRules{{Rule: "rule2"}, {Rule: "rule3"}}},
					{XValidations: apiextensionsv1.ValidationRules{{Rule: "rule1"}}},
				},
			}
			flattened := crd.FlattenEmbedded(original, errRec)
			Expect(errRec.FirstError()).NotTo(HaveOccurred())

			By("ensuring that the result lists all validation rules")
			Expect(flattened).To(Equal(&apiextensionsv1.JSONSchemaProps{
				XValidations: []apiextensionsv1.ValidationRule{
					{Rule: "rule1"}, {Rule: "rule2"}, {Rule: "rule3"},
				},
			}))
		})
	})

	It("should skip Title, Description, Example, and ExternalDocs, assuming they've been merged pre-AllOf flattening", func() {
		By("flattening a schema with documentation in and out of an AllOf branch")
		original := apiextensionsv1.JSONSchemaProps{
			AllOf: []apiextensionsv1.JSONSchemaProps{
				{Title: "a title"},
				{Description: "a desc"},
				{Example: &apiextensionsv1.JSON{Raw: []byte("an ex")}},
				{ExternalDocs: &apiextensionsv1.ExternalDocumentation{Description: "some exdocs", URL: "https://other.example.com"}},
			},
			Title:        "title",
			Description:  "desc",
			Example:      &apiextensionsv1.JSON{Raw: []byte("ex")},
			ExternalDocs: &apiextensionsv1.ExternalDocumentation{Description: "exdocs", URL: "https://example.com"},
		}
		flattened := crd.FlattenEmbedded(&original, errRec)
		Expect(errRec.FirstError()).NotTo(HaveOccurred())

		By("ensuring the flattened schema only has documentation outside the AllOf branch")
		Expect(flattened).To(Equal(&apiextensionsv1.JSONSchemaProps{
			Title:        "title",
			Description:  "desc",
			Example:      &apiextensionsv1.JSON{Raw: []byte("ex")},
			ExternalDocs: &apiextensionsv1.ExternalDocumentation{Description: "exdocs", URL: "https://example.com"},
		}))
	})

	It("should just use the value when only one AllOf branch specifies a value", func() {
		By("flattening a schema with non-conflicting branches")
		defTwo := int64(2)
		original := apiextensionsv1.JSONSchemaProps{
			AllOf: []apiextensionsv1.JSONSchemaProps{
				{Type: "string"},
				{MinLength: &defTwo},
				{Enum: []apiextensionsv1.JSON{{Raw: []byte("ab")}, {Raw: []byte("ac")}}},
			},
		}
		flattened := crd.FlattenEmbedded(&original, errRec)
		Expect(errRec.FirstError()).NotTo(HaveOccurred())

		By("checking that the result doesn't have any branches")
		Expect(flattened).To(Equal(&apiextensionsv1.JSONSchemaProps{
			Type:      "string",
			MinLength: &defTwo,
			Enum:      []apiextensionsv1.JSON{{Raw: []byte("ab")}, {Raw: []byte("ac")}},
		}))
	})

	Context("for all other types", func() {
		It("should push the AllOf as far down the stack as possible, eliminating it if possible", func() {
			By("flattening a high-up AllOf with a low-down difference")
			original := apiextensionsv1.JSONSchemaProps{
				AllOf: []apiextensionsv1.JSONSchemaProps{
					{
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"prop1": {
								Properties: map[string]apiextensionsv1.JSONSchemaProps{
									"prop2": {
										Type:    "string",
										Pattern: "^[abc]+$",
									},
								},
							},
						},
					},
					{
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"prop1": {
								Properties: map[string]apiextensionsv1.JSONSchemaProps{
									"prop2": {
										Pattern: "^(bc)+$",
									},
								},
							},
						},
					},
				},
			}
			flattened := crd.FlattenEmbedded(&original, errRec)
			Expect(errRec.FirstError()).NotTo(HaveOccurred())

			By("ensuring that the result has the minimal AllOf branches possible")
			Expect(flattened).To(Equal(&apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"prop1": {
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"prop2": {
								Type:  "string",
								AllOf: []apiextensionsv1.JSONSchemaProps{{Pattern: "^[abc]+$"}, {Pattern: "^(bc)+$"}},
							},
						},
					},
				},
			}))
		})
	})

	It("should leave properties not in an AllOf branch (and minimal AllOf branches) alone", func() {
		By("flattening an irreducible schema")
		original := &apiextensionsv1.JSONSchemaProps{
			Type:  "string",
			AllOf: []apiextensionsv1.JSONSchemaProps{{Pattern: "^[abc]+$"}, {Pattern: "^(bc)+$"}},
		}
		flattened := crd.FlattenEmbedded(original.DeepCopy() /* DeepCopy so we can compare later */, errRec)
		Expect(errRec.FirstError()).NotTo(HaveOccurred())

		By("checking that the flattened version is unmodified")
		Expect(flattened).To(Equal(original))
	})

	It("should flattened nested AllOfs as normal", func() {
		By("flattening a schema with nested AllOf branches")
		defOne := int64(1)
		original := apiextensionsv1.JSONSchemaProps{
			AllOf: []apiextensionsv1.JSONSchemaProps{
				{
					AllOf: []apiextensionsv1.JSONSchemaProps{
						{Pattern: "^[abc]$"},
						{Pattern: "^[abcdef]$", MinLength: &defOne},
					},
				},
				{
					Type: "string",
				},
			},
		}
		flattened := crd.FlattenEmbedded(original.DeepCopy() /* DeepCopy so we can compare later */, errRec)
		Expect(errRec.FirstError()).NotTo(HaveOccurred())

		By("ensuring that the flattened version is contains the minimal branches")
		Expect(flattened).To(Equal(&apiextensionsv1.JSONSchemaProps{
			Type:      "string",
			MinLength: &defOne,
			AllOf:     []apiextensionsv1.JSONSchemaProps{{Pattern: "^[abc]$"}, {Pattern: "^[abcdef]$"}},
		}))
	})
})
