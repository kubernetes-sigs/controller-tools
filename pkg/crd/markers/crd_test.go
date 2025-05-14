package markers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/controller-tools/pkg/crd/markers"
)

var _ = Describe("SchemaModifierMarker", func() {
	baseCRD := func() *apiext.CustomResourceDefinitionSpec {
		return &apiext.CustomResourceDefinitionSpec{
			Versions: []apiext.CustomResourceDefinitionVersion{
				{
					Name: "v1",
					Schema: &apiext.CustomResourceValidation{
						OpenAPIV3Schema: &apiext.JSONSchemaProps{
							Properties: map[string]apiext.JSONSchemaProps{
								"spec": {
									Properties: map[string]apiext.JSONSchemaProps{},
								},
							},
						},
					},
				},
			},
		}
	}
	addToSpec := func(crd *apiext.CustomResourceDefinitionSpec, key string, props apiext.JSONSchemaProps) {
		crd.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties[key] = props
	}

	Context("Pattern matching", func() {
		const (
			descOrig     = "original"
			descExpected = "modified"
		)
		barProps := func(desc string) map[string]apiext.JSONSchemaProps {
			return map[string]apiext.JSONSchemaProps{
				"foo": {
					Description: desc,
					Properties: map[string]apiext.JSONSchemaProps{
						"bar": {
							Description: desc,
						},
					},
				},
				"baz": {
					Description: desc,
				},
			}
		}

		It("should match only direct children with pattern /*", func() {
			crdOrig := baseCRD()
			crdExpected := baseCRD()

			addToSpec(crdOrig, "foo", apiext.JSONSchemaProps{Description: descOrig})
			addToSpec(crdOrig, "bar", apiext.JSONSchemaProps{
				Description: descOrig,
				Properties:  barProps(descOrig),
			})

			addToSpec(crdExpected, "foo", apiext.JSONSchemaProps{Description: descExpected})
			addToSpec(crdExpected, "bar", apiext.JSONSchemaProps{
				Description: descExpected,
				Properties:  barProps(descOrig),
			})

			marker := &markers.SchemaModifier{
				PathPattern: "/spec/*",
				Description: ptr.To(descExpected),
			}

			Expect(marker.ApplyToCRD(crdOrig, "v1")).To(Succeed())
			Expect(crdOrig).To(Equal(crdExpected))
		})

		It("should match deep nested fields with /**", func() {
			crdOrig := baseCRD()
			crdExpected := baseCRD()

			addToSpec(crdOrig, "foo", apiext.JSONSchemaProps{Description: descOrig})
			addToSpec(crdOrig, "bar", apiext.JSONSchemaProps{
				Description: descOrig,
				Properties:  barProps(descOrig),
			})

			addToSpec(crdExpected, "foo", apiext.JSONSchemaProps{Description: descExpected})
			addToSpec(crdExpected, "bar", apiext.JSONSchemaProps{
				Description: descExpected,
				Properties:  barProps(descExpected),
			})

			marker := &markers.SchemaModifier{
				PathPattern: "/spec/**",
				Description: ptr.To(descExpected),
			}

			Expect(marker.ApplyToCRD(crdOrig, "v1")).To(Succeed())
			Expect(crdOrig).To(Equal(crdExpected))
		})

		It("should return error on invalid path pattern", func() {
			crd := baseCRD()
			addToSpec(crd, "foo", apiext.JSONSchemaProps{Description: descOrig})

			marker := &markers.SchemaModifier{
				PathPattern: "[invalid-regex",
			}

			err := marker.ApplyToCRD(crd, "v1")
			Expect(err).To(HaveOccurred())
		})
	})

	DescribeTable("Should modify crd /spec/foo", func(origFooProps, expectedFooProps apiext.JSONSchemaProps, marker *markers.SchemaModifier) {
		crdOrig := baseCRD()
		addToSpec(crdOrig, "foo", origFooProps)

		crdExpected := baseCRD()
		addToSpec(crdExpected, "foo", expectedFooProps)

		marker.PathPattern = "/spec/foo"
		Expect(marker.ApplyToCRD(crdOrig, "v1")).To(Succeed())
		Expect(crdOrig).To(Equal(crdExpected))
	},
		Entry("should trim description",
			apiext.JSONSchemaProps{Description: "foo"},
			apiext.JSONSchemaProps{Description: ""},
			&markers.SchemaModifier{Description: ptr.To("")},
		),
		Entry("should replace format",
			apiext.JSONSchemaProps{Format: "foo"},
			apiext.JSONSchemaProps{Format: "bar"},
			&markers.SchemaModifier{Format: ptr.To("bar")},
		),
		Entry("should replace maximum",
			apiext.JSONSchemaProps{Maximum: ptr.To(1.0)},
			apiext.JSONSchemaProps{Maximum: ptr.To(2.0)},
			&markers.SchemaModifier{Maximum: ptr.To(2.0)},
		),
		Entry("should replace exclusiveMaximum",
			apiext.JSONSchemaProps{ExclusiveMaximum: true},
			apiext.JSONSchemaProps{ExclusiveMaximum: false},
			&markers.SchemaModifier{ExclusiveMaximum: ptr.To(false)},
		),
		Entry("should replace minimum",
			apiext.JSONSchemaProps{Minimum: ptr.To(1.0)},
			apiext.JSONSchemaProps{Minimum: ptr.To(2.0)},
			&markers.SchemaModifier{Minimum: ptr.To(2.0)},
		),
		Entry("should replace exclusiveMinimum",
			apiext.JSONSchemaProps{ExclusiveMinimum: true},
			apiext.JSONSchemaProps{ExclusiveMinimum: false},
			&markers.SchemaModifier{ExclusiveMinimum: ptr.To(false)},
		),
		Entry("should replace maxLength",
			apiext.JSONSchemaProps{MaxLength: ptr.To[int64](1)},
			apiext.JSONSchemaProps{MaxLength: ptr.To[int64](2)},
			&markers.SchemaModifier{MaxLength: ptr.To(2)},
		),
		Entry("should replace minLength",
			apiext.JSONSchemaProps{MinLength: ptr.To[int64](1)},
			apiext.JSONSchemaProps{MinLength: ptr.To[int64](2)},
			&markers.SchemaModifier{MinLength: ptr.To(2)},
		),
		Entry("should replace pattern",
			apiext.JSONSchemaProps{Pattern: "foo"},
			apiext.JSONSchemaProps{Pattern: "bar"},
			&markers.SchemaModifier{Pattern: ptr.To("bar")},
		),
		Entry("should replace maxItems",
			apiext.JSONSchemaProps{MaxItems: ptr.To[int64](1)},
			apiext.JSONSchemaProps{MaxItems: ptr.To[int64](2)},
			&markers.SchemaModifier{MaxItems: ptr.To(2)},
		),
		Entry("should replace minItems",
			apiext.JSONSchemaProps{MinItems: ptr.To[int64](1)},
			apiext.JSONSchemaProps{MinItems: ptr.To[int64](2)},
			&markers.SchemaModifier{MinItems: ptr.To(2)},
		),
		Entry("should replace uniqueItems",
			apiext.JSONSchemaProps{UniqueItems: true},
			apiext.JSONSchemaProps{UniqueItems: false},
			&markers.SchemaModifier{UniqueItems: ptr.To(false)},
		),
		Entry("should replace multipleOf",
			apiext.JSONSchemaProps{MultipleOf: ptr.To(1.0)},
			apiext.JSONSchemaProps{MultipleOf: ptr.To(2.0)},
			&markers.SchemaModifier{MultipleOf: ptr.To(2.0)},
		),
		Entry("should replace maxProperties",
			apiext.JSONSchemaProps{MaxProperties: ptr.To[int64](1)},
			apiext.JSONSchemaProps{MaxProperties: ptr.To[int64](2)},
			&markers.SchemaModifier{MaxProperties: ptr.To(2)},
		),
		Entry("should replace minProperties",
			apiext.JSONSchemaProps{MinProperties: ptr.To[int64](1)},
			apiext.JSONSchemaProps{MinProperties: ptr.To[int64](2)},
			&markers.SchemaModifier{MinProperties: ptr.To(2)},
		),
		Entry("should replace required",
			apiext.JSONSchemaProps{Required: []string{"foo"}},
			apiext.JSONSchemaProps{Required: []string{"bar"}},
			&markers.SchemaModifier{Required: ptr.To([]string{"bar"})},
		),
		Entry("should replace nullable",
			apiext.JSONSchemaProps{Nullable: true},
			apiext.JSONSchemaProps{Nullable: false},
			&markers.SchemaModifier{Nullable: ptr.To(false)},
		),
	)

	Context("Parse pattern", func() {
		It("should convert * to [^/]+", func() {
			sm := markers.SchemaModifier{PathPattern: "/spec/*"}
			re, err := sm.ParsePattern()
			Expect(err).NotTo(HaveOccurred())
			Expect(re.String()).To(Equal("^/spec/[^/]+$"))
		})

		It("should convert ** to .*", func() {
			sm := markers.SchemaModifier{PathPattern: "/spec/**/field"}
			re, err := sm.ParsePattern()
			Expect(err).NotTo(HaveOccurred())
			Expect(re.String()).To(Equal("^/spec/.*/field$"))
		})

		It("should fail with invalid rule", func() {
			sm := markers.SchemaModifier{PathPattern: "[invalid-regex"}
			_, err := sm.ParsePattern()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("error parsing regexp")))
		})
	})
})
