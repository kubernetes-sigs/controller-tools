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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/loader"
)

var _ = Describe("General Schema Flattening", func() {
	var fl *crd.Flattener

	var (
		// just enough so we don't panic
		rootPkg  = &loader.Package{Package: &packages.Package{PkgPath: "root"}}
		otherPkg = &loader.Package{Package: &packages.Package{PkgPath: "other"}}

		rootType        = crd.TypeIdent{Name: "RootType", Package: rootPkg}
		subtypeWithRefs = crd.TypeIdent{Name: "SubtypeWithRefs", Package: rootPkg}
		leafAliasType   = crd.TypeIdent{Name: "LeafAlias", Package: rootPkg}
		leafType        = crd.TypeIdent{Name: "LeafType", Package: otherPkg}
		inPkgLeafType   = crd.TypeIdent{Name: "InPkgLeafType", Package: rootPkg}
	)

	BeforeEach(func() {
		fl = &crd.Flattener{
			Parser: &crd.Parser{
				Schemata: map[crd.TypeIdent]apiextensionsv1.JSONSchemaProps{},
				PackageOverrides: map[string]crd.PackageOverride{
					"root":  func(_ *crd.Parser, _ *loader.Package) {},
					"other": func(_ *crd.Parser, _ *loader.Package) {},
				},
			},
			LookupReference: func(ref string, contextPkg *loader.Package) (crd.TypeIdent, error) {
				typ, pkgName, err := crd.RefParts(ref)
				if err != nil {
					return crd.TypeIdent{}, err
				}

				// cheat and just treat these as global
				switch pkgName {
				case "":
					return crd.TypeIdent{Name: typ, Package: contextPkg}, nil
				case "root":
					return crd.TypeIdent{Name: typ, Package: rootPkg}, nil
				case "other":
					return crd.TypeIdent{Name: typ, Package: otherPkg}, nil
				default:
					return crd.TypeIdent{}, fmt.Errorf("unknown package %q", pkgName)
				}

			},
		}
	})

	Context("when dealing with reference chains", func() {
		It("should flatten them", func() {
			By("setting up a RootType, LeafAlias --> Alias --> Int")
			toLeafAlias := crd.TypeRefLink("", leafAliasType.Name)
			toLeaf := crd.TypeRefLink("other", leafType.Name)
			fl.Parser.Schemata = map[crd.TypeIdent]apiextensionsv1.JSONSchemaProps{
				rootType: {
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"refProp": {Ref: &toLeafAlias},
					},
				},
				leafAliasType: {Ref: &toLeaf},
				leafType: {
					Type:    "string",
					Pattern: "^[abc]$",
				},
			}

			By("flattening the type hierarchy")
			// flattenAllOf to avoid the normalize the all-of forms to what we
			// really want (instead of caring about nested all-ofs)
			outSchema := crd.FlattenEmbedded(fl.FlattenType(rootType), rootPkg)
			Expect(rootPkg.Errors).To(HaveLen(0))
			Expect(otherPkg.Errors).To(HaveLen(0))

			By("verifying that it was flattened to have no references")
			Expect(outSchema).To(Equal(&apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"refProp": {
						Type: "string", Pattern: "^[abc]$",
					},
				},
			}))
		})

		It("should not infinite-loop on circular references", func() {
			By("setting up a RootType, LeafAlias --> Alias --> LeafAlias")
			toLeafAlias := crd.TypeRefLink("", leafAliasType.Name)
			toLeaf := crd.TypeRefLink("", inPkgLeafType.Name)
			fl.Parser.Schemata = map[crd.TypeIdent]apiextensionsv1.JSONSchemaProps{
				rootType: {
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"refProp": {Ref: &toLeafAlias},
					},
				},
				leafAliasType: {Ref: &toLeaf},
				inPkgLeafType: {Ref: &toLeafAlias},
			}

			By("flattening the type hierarchy")
			// flattenAllOf to avoid the normalize the all-of forms to what we
			// really want (instead of caring about nested all-ofs)
			outSchema := crd.FlattenEmbedded(fl.FlattenType(rootType), rootPkg)

			// This should *finish* to some degree, leaving the circular reference in
			// place.  It should be fine to error on circular references in the future, though.
			Expect(rootPkg.Errors).To(HaveLen(0))
			Expect(otherPkg.Errors).To(HaveLen(0))

			By("verifying that it was flattened to *something*")
			Expect(outSchema).To(Equal(&apiextensionsv1.JSONSchemaProps{
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"refProp": {
						Ref: &toLeafAlias,
					},
				},
			}))
		})
	})

	It("should flatten a hierarchy of references", func() {
		By("setting up a series of types RootType --> SubtypeWithRef --> LeafType")
		toSubtype := crd.TypeRefLink("", subtypeWithRefs.Name)
		toLeaf := crd.TypeRefLink("other", leafType.Name)
		fl.Parser.Schemata = map[crd.TypeIdent]apiextensionsv1.JSONSchemaProps{
			rootType: {
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"refProp": {Ref: &toSubtype},
				},
			},
			subtypeWithRefs: {
				AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
					Schema: &apiextensionsv1.JSONSchemaProps{
						Ref: &toLeaf,
					},
				},
			},
			leafType: {
				Type:    "string",
				Pattern: "^[abc]$",
			},
		}

		By("flattening the type hierarchy")
		outSchema := fl.FlattenType(rootType)
		Expect(rootPkg.Errors).To(HaveLen(0))
		Expect(otherPkg.Errors).To(HaveLen(0))

		By("verifying that it was flattened to have no references")
		Expect(outSchema).To(Equal(&apiextensionsv1.JSONSchemaProps{
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"refProp": {
					AllOf: []apiextensionsv1.JSONSchemaProps{
						{
							AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
								Schema: &apiextensionsv1.JSONSchemaProps{
									AllOf: []apiextensionsv1.JSONSchemaProps{
										{Type: "string", Pattern: "^[abc]$"},
										{},
									},
								},
							},
						},
						{},
					},
				},
			},
		}))
	})

	It("should preserve the properties of each separate use of a type without modifying the cache", func() {
		By("setting up a series of types RootType --> LeafType with 3 uses")
		defOne := int64(1)
		defThree := int64(3)
		toLeaf := crd.TypeRefLink("other", leafType.Name)
		fl.Parser.Schemata = map[crd.TypeIdent]apiextensionsv1.JSONSchemaProps{
			rootType: {
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"useWithOtherPattern": {
						Ref:         &toLeaf,
						Pattern:     "^[cde]$",
						Description: "has other pattern",
					},
					"useWithMinLen": {
						Ref:         &toLeaf,
						MinLength:   &defOne,
						Description: "has min len",
					},
					"useWithMaxLen": {
						Ref:         &toLeaf,
						MaxLength:   &defThree,
						Description: "has max len",
					},
				},
			},
			leafType: {
				Type:    "string",
				Pattern: "^[abc]$",
			},
		}

		By("flattening the type hierarchy")
		outSchema := fl.FlattenType(rootType)
		Expect(rootPkg.Errors).To(HaveLen(0))
		Expect(otherPkg.Errors).To(HaveLen(0))

		By("verifying that each use has its own properties set in allof branches")
		Expect(outSchema).To(Equal(&apiextensionsv1.JSONSchemaProps{
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"useWithOtherPattern": {
					AllOf: []apiextensionsv1.JSONSchemaProps{
						{Type: "string", Pattern: "^[abc]$"},
						{Pattern: "^[cde]$"},
					},
					Description: "has other pattern",
				},
				"useWithMinLen": {
					AllOf: []apiextensionsv1.JSONSchemaProps{
						{Type: "string", Pattern: "^[abc]$"},
						{MinLength: &defOne},
					},
					Description: "has min len",
				},
				"useWithMaxLen": {
					AllOf: []apiextensionsv1.JSONSchemaProps{
						{Type: "string", Pattern: "^[abc]$"},
						{MaxLength: &defThree},
					},
					Description: "has max len",
				},
			},
		}))
	})

	It("should copy over documentation for each use of a type", func() {
		By("setting up a series of types RootType --> LeafType with 3 doc-only uses")
		toLeaf := crd.TypeRefLink("other", leafType.Name)
		fl.Parser.Schemata = map[crd.TypeIdent]apiextensionsv1.JSONSchemaProps{
			rootType: {
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"hasTitle": {
						Ref:         &toLeaf,
						Description: "has title",
						Title:       "some title",
					},
					"hasExample": {
						Ref:         &toLeaf,
						Description: "has example",
						Example:     &apiextensionsv1.JSON{Raw: []byte("[42]")},
					},
					"hasExternalDocs": {
						Ref:         &toLeaf,
						Description: "has external docs",
						ExternalDocs: &apiextensionsv1.ExternalDocumentation{
							Description: "somewhere else",
							URL:         "https://example.com", // RFC 2606
						},
					},
				},
			},
			leafType: {
				Type:    "string",
				Pattern: "^[abc]$",
			},
		}

		By("flattening the type hierarchy")
		outSchema := fl.FlattenType(rootType)
		Expect(rootPkg.Errors).To(HaveLen(0))
		Expect(otherPkg.Errors).To(HaveLen(0))

		By("verifying that each use has its own properties set in allof branches")
		Expect(outSchema).To(Equal(&apiextensionsv1.JSONSchemaProps{
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"hasTitle": {
					AllOf:       []apiextensionsv1.JSONSchemaProps{{Type: "string", Pattern: "^[abc]$"}, {}},
					Description: "has title",
					Title:       "some title",
				},
				"hasExample": {
					AllOf:       []apiextensionsv1.JSONSchemaProps{{Type: "string", Pattern: "^[abc]$"}, {}},
					Description: "has example",
					Example:     &apiextensionsv1.JSON{Raw: []byte("[42]")},
				},
				"hasExternalDocs": {
					AllOf:       []apiextensionsv1.JSONSchemaProps{{Type: "string", Pattern: "^[abc]$"}, {}},
					Description: "has external docs",
					ExternalDocs: &apiextensionsv1.ExternalDocumentation{
						Description: "somewhere else",
						URL:         "https://example.com", // RFC 2606
					},
				},
			},
		}))
	})

	It("should ignore schemata that aren't references, but continue flattening", func() {
		By("setting up a series of types RootType --> LeafType with non-ref properties")
		toLeaf := crd.TypeRefLink("other", leafType.Name)
		toSubtype := crd.TypeRefLink("", subtypeWithRefs.Name)
		fl.Parser.Schemata = map[crd.TypeIdent]apiextensionsv1.JSONSchemaProps{
			rootType: {
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"isRef": {
						Ref: &toSubtype,
					},
					"notRef": {
						Type: "int",
					},
				},
			},
			subtypeWithRefs: {
				Properties: map[string]apiextensionsv1.JSONSchemaProps{
					"leafRef": {
						Ref: &toLeaf,
					},
					"alsoNotRef": {
						Type: "bool",
					},
				},
			},
			leafType: {
				Type:    "string",
				Pattern: "^[abc]$",
			},
		}

		By("flattening the type hierarchy")
		outSchema := fl.FlattenType(rootType)
		Expect(rootPkg.Errors).To(HaveLen(0))
		Expect(otherPkg.Errors).To(HaveLen(0))

		By("verifying that each use has its own properties set in allof branches")
		Expect(outSchema).To(Equal(&apiextensionsv1.JSONSchemaProps{
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"isRef": {
					AllOf: []apiextensionsv1.JSONSchemaProps{
						{
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"leafRef": {
									AllOf: []apiextensionsv1.JSONSchemaProps{
										{Type: "string", Pattern: "^[abc]$"}, {},
									},
								},
								"alsoNotRef": {
									Type: "bool",
								},
							},
						},
						{},
					},
				},
				"notRef": {
					Type: "int",
				},
			},
		}))

	})
})
