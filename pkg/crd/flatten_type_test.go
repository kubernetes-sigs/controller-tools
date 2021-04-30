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
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"golang.org/x/tools/go/packages"
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
				Schemata: map[crd.TypeIdent]apiext.JSONSchemaProps{},
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
			fl.Parser.Schemata = map[crd.TypeIdent]apiext.JSONSchemaProps{
				rootType: {
					Properties: map[string]apiext.JSONSchemaProps{
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
			Expect(outSchema).To(Equal(&apiext.JSONSchemaProps{
				Properties: map[string]apiext.JSONSchemaProps{
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
			fl.Parser.Schemata = map[crd.TypeIdent]apiext.JSONSchemaProps{
				rootType: {
					Properties: map[string]apiext.JSONSchemaProps{
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
			Expect(outSchema).To(Equal(&apiext.JSONSchemaProps{
				Properties: map[string]apiext.JSONSchemaProps{
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
		fl.Parser.Schemata = map[crd.TypeIdent]apiext.JSONSchemaProps{
			rootType: {
				Properties: map[string]apiext.JSONSchemaProps{
					"refProp": {Ref: &toSubtype},
				},
			},
			subtypeWithRefs: {
				AdditionalProperties: &apiext.JSONSchemaPropsOrBool{
					Schema: &apiext.JSONSchemaProps{
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
		Expect(outSchema).To(Equal(&apiext.JSONSchemaProps{
			Properties: map[string]apiext.JSONSchemaProps{
				"refProp": {
					AllOf: []apiext.JSONSchemaProps{
						{
							AdditionalProperties: &apiext.JSONSchemaPropsOrBool{
								Schema: &apiext.JSONSchemaProps{
									AllOf: []apiext.JSONSchemaProps{
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
		fl.Parser.Schemata = map[crd.TypeIdent]apiext.JSONSchemaProps{
			rootType: {
				Properties: map[string]apiext.JSONSchemaProps{
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
		Expect(outSchema).To(Equal(&apiext.JSONSchemaProps{
			Properties: map[string]apiext.JSONSchemaProps{
				"useWithOtherPattern": {
					AllOf: []apiext.JSONSchemaProps{
						{Type: "string", Pattern: "^[abc]$"},
						{Pattern: "^[cde]$"},
					},
					Description: "has other pattern",
				},
				"useWithMinLen": {
					AllOf: []apiext.JSONSchemaProps{
						{Type: "string", Pattern: "^[abc]$"},
						{MinLength: &defOne},
					},
					Description: "has min len",
				},
				"useWithMaxLen": {
					AllOf: []apiext.JSONSchemaProps{
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
		fl.Parser.Schemata = map[crd.TypeIdent]apiext.JSONSchemaProps{
			rootType: {
				Properties: map[string]apiext.JSONSchemaProps{
					"hasTitle": {
						Ref:         &toLeaf,
						Description: "has title",
						Title:       "some title",
					},
					"hasExample": {
						Ref:         &toLeaf,
						Description: "has example",
						Example:     &apiext.JSON{Raw: []byte("[42]")},
					},
					"hasExternalDocs": {
						Ref:         &toLeaf,
						Description: "has external docs",
						ExternalDocs: &apiext.ExternalDocumentation{
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
		Expect(outSchema).To(Equal(&apiext.JSONSchemaProps{
			Properties: map[string]apiext.JSONSchemaProps{
				"hasTitle": {
					AllOf:       []apiext.JSONSchemaProps{{Type: "string", Pattern: "^[abc]$"}, {}},
					Description: "has title",
					Title:       "some title",
				},
				"hasExample": {
					AllOf:       []apiext.JSONSchemaProps{{Type: "string", Pattern: "^[abc]$"}, {}},
					Description: "has example",
					Example:     &apiext.JSON{Raw: []byte("[42]")},
				},
				"hasExternalDocs": {
					AllOf:       []apiext.JSONSchemaProps{{Type: "string", Pattern: "^[abc]$"}, {}},
					Description: "has external docs",
					ExternalDocs: &apiext.ExternalDocumentation{
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
		fl.Parser.Schemata = map[crd.TypeIdent]apiext.JSONSchemaProps{
			rootType: {
				Properties: map[string]apiext.JSONSchemaProps{
					"isRef": {
						Ref: &toSubtype,
					},
					"notRef": {
						Type: "int",
					},
				},
			},
			subtypeWithRefs: {
				Properties: map[string]apiext.JSONSchemaProps{
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
		Expect(outSchema).To(Equal(&apiext.JSONSchemaProps{
			Properties: map[string]apiext.JSONSchemaProps{
				"isRef": {
					AllOf: []apiext.JSONSchemaProps{
						{
							Properties: map[string]apiext.JSONSchemaProps{
								"leafRef": {
									AllOf: []apiext.JSONSchemaProps{
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
