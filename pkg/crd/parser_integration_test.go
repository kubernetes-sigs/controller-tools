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
	"os"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/yaml"
)

func packageErrors(pkg *loader.Package, filterKinds ...packages.ErrorKind) error {
	toSkip := make(map[packages.ErrorKind]struct{})
	for _, errKind := range filterKinds {
		toSkip[errKind] = struct{}{}
	}
	var outErr error
	packages.Visit([]*packages.Package{pkg.Package}, nil, func(pkgRaw *packages.Package) {
		for _, err := range pkgRaw.Errors {
			if _, skip := toSkip[err.Kind]; skip {
				continue
			}
			outErr = err
		}
	})
	return outErr
}

var _ = Describe("CRD Generation From Parsing to CustomResourceDefinition", func() {
	Context("should properly generate and flatten the rewritten schemas", func() {
		var (
			prevCwd   string
			pkgPaths  []string
			pkgs      []*loader.Package
			reg       *markers.Registry
			parser    *crd.Parser
			expPkgLen int
		)

		BeforeEach(func() {
			var err error
			By("switching into testdata to appease go modules")
			prevCwd, err = os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			Expect(prevCwd).To(BeADirectory())
			Expect(os.Chdir("./testdata")).To(Succeed())

			By("setting up the parser")
			reg = &markers.Registry{}
			Expect(crdmarkers.Register(reg)).To(Succeed())
			parser = &crd.Parser{
				Collector:              &markers.Collector{Registry: reg},
				Checker:                &loader.TypeChecker{},
				IgnoreUnexportedFields: true,
				AllowDangerousTypes:    true, // need to allow “dangerous types” in this file for testing
			}
			crd.AddKnownTypes(parser)
		})

		AfterEach(func() {
			Expect(os.Chdir(prevCwd)).To(Succeed())
		})

		JustBeforeEach(func() {
			var err error
			By("loading the roots")
			pkgs, err = loader.LoadRoots(pkgPaths...)
			Expect(err).NotTo(HaveOccurred())
			Expect(pkgs).To(HaveLen(expPkgLen))

			By("requesting that the packages be parsed")
			for _, p := range pkgs {
				parser.NeedPackage(p)
			}
		})

		assertCRDForGroupKind := func(pkg *loader.Package, groupKind schema.GroupKind, fileName string) {
			kind := groupKind.Kind
			By(fmt.Sprintf("requesting that the %s CRD be generated in group %s", kind, groupKind.Group))
			parser.NeedCRDFor(groupKind, nil)

			By(fmt.Sprintf("fixing top level ObjectMeta on the %s CRD", kind))
			crd.FixTopLevelMetadata(parser.CustomResourceDefinitions[groupKind])

			By("checking that no errors occurred along the way (expect for type errors)")
			ExpectWithOffset(1, packageErrors(pkg, packages.TypeError)).NotTo(HaveOccurred())

			By(fmt.Sprintf("checking that the %s CRD is present", kind))
			ExpectWithOffset(1, parser.CustomResourceDefinitions).To(HaveKey(groupKind))

			By(fmt.Sprintf("loading the desired %s YAML", kind))
			expectedFile, err := os.ReadFile(fileName)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By(fmt.Sprintf("parsing the desired %s YAML", kind))
			var crd apiextensionsv1.CustomResourceDefinition
			ExpectWithOffset(1, yaml.Unmarshal(expectedFile, &crd)).To(Succeed())
			// clear the version annotation -- we don't care about the attribution annotation
			delete(crd.Annotations, "controller-gen.kubebuilder.io/version")
			if len(crd.Annotations) == 0 {
				crd.Annotations = nil
			}

			By(fmt.Sprintf("comparing the two %s CRDs", kind))
			// Use cmpopts.EquateEmpty() to treat nil and empty maps/slices as equal.
			// k8s.io/api v0.37.0+ produces nil Properties for object types without
			// explicit properties, while YAML deserialization produces empty maps.
			cmpOpts := cmpopts.EquateEmpty()
			ExpectWithOffset(1, cmp.Equal(parser.CustomResourceDefinitions[groupKind], crd, cmpOpts)).To(BeTrue(), "type not as expected, check pkg/crd/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(parser.CustomResourceDefinitions[groupKind], crd, cmpOpts))
		}

		assertCRD := func(pkg *loader.Package, kind, fileName string) {
			assertCRDForGroupKind(pkg, schema.GroupKind{Group: "testdata.kubebuilder.io", Kind: kind}, fileName)
		}

		assertError := func(pkg *loader.Package, kind, errorMsg string) {
			By(fmt.Sprintf("requesting that the %s CRD be generated", kind))
			groupKind := schema.GroupKind{Kind: kind, Group: "testdata.kubebuilder.io"}
			parser.NeedCRDFor(groupKind, nil)

			By(fmt.Sprintf("fixing top level ObjectMeta on the %s CRD", kind))
			crd.FixTopLevelMetadata(parser.CustomResourceDefinitions[groupKind])

			By("checking that specific errors occurred along the way")
			Expect(packageErrors(pkg)).To(MatchError(ContainSubstring(errorMsg)))
		}

		Context("CronJob API", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./", "./unserved", "./deprecated"}
				expPkgLen = 3
			})
			It("should successfully generate the CronJob CRD", func() {
				assertCRD(pkgs[0], "CronJob", "testdata.kubebuilder.io_cronjobs.yaml")
			})
		})

		Context("Job API", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./job/..."}
				expPkgLen = 1
			})
			It("should successfully generate the Job CRD", func() {
				assertCRD(pkgs[0], "Job", "testdata.kubebuilder.io_jobs.yaml")
			})
		})

		Context("CronJob and Job API", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./", "./unserved", "./deprecated", "./job/..."}
				expPkgLen = 4
			})
			It("should successfully generate the CronJob and Job CRDs", func() {
				assertCRD(pkgs[0], "CronJob", "testdata.kubebuilder.io_cronjobs.yaml")
				assertCRD(pkgs[3], "Job", "testdata.kubebuilder.io_jobs.yaml")
			})
		})

		Context("OneOf API", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./oneof/..."}
				expPkgLen = 1
			})
			It("should successfully generate the CRD with OneOf validation constraints", func() {
				assertCRD(pkgs[0], "Oneof", "testdata.kubebuilder.io_oneofs.yaml")
			})
		})

		Context("Enum API", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./enum/..."}
				expPkgLen = 1
			})
			It("should successfully generate the CRD with enum validation constraints", func() {
				assertCRD(pkgs[0], "Enum", "testdata.kubebuilder.io_enums.yaml")
			})
		})

		Context("EmptyObjectDefault API", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./emptyobjectdefault/..."}
				expPkgLen = 1
			})
			It("should generate default: {} for pointer fields defaulting to an empty object", func() {
				assertCRD(pkgs[0], "EmptyObjectDefault", "testdata.kubebuilder.io_emptyobjectdefaults.yaml")
				assertCRD(pkgs[0], "RequiredChildDefault", "testdata.kubebuilder.io_requiredchilddefaults.yaml")
			})
		})

		Context("Enum API with marker on field", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./enum_error/..."}
				expPkgLen = 1
			})
			It("should generate an error when k8s:enum is set on a field", func() {
				groupKind := schema.GroupKind{Kind: "EnumError", Group: "testdata.kubebuilder.io"}
				parser.NeedCRDFor(groupKind, nil)

				expectedErr := "k8s:enum must be set on a type, not a field"
				Expect(packageErrors(pkgs[0])).To(MatchError(ContainSubstring(expectedErr)))
			})
		})

		Context("Inline struct error", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./inline_struct_error/..."}
				expPkgLen = 1
			})
			It("should reject inline struct literals with a clear error message", func() {
				groupKind := schema.GroupKind{Kind: "InlineStruct", Group: "testdata.kubebuilder.io"}
				parser.NeedCRDFor(groupKind, nil)

				expectedErr := "inline struct types are not supported, use a named type instead"
				Expect(packageErrors(pkgs[0])).To(MatchError(ContainSubstring(expectedErr)))
			})
		})

		Context("OneOf API with invalid marker", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./oneof_error/..."}
				expPkgLen = 1
			})
			It("should generate an error with nested field in marker", func() {
				kind := "Oneof"
				groupKind := schema.GroupKind{Kind: kind, Group: "testdata.kubebuilder.io"}
				parser.NeedCRDFor(groupKind, nil)

				expectedErr := "kubebuilder:validation:AtMostOneOf: cannot reference nested fields: field.foo,field.bar"
				Expect(packageErrors(pkgs[0])).To(MatchError(ContainSubstring(expectedErr)))
			})
		})

		Context("OneOf API with missing omitempty/omitzero tag", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./oneof_missing_tag_error/..."}
				expPkgLen = 1
			})
			It("should generate an error when OneOf field lacks omitempty or omitzero tag", func() {
				kind := "Oneof"
				groupKind := schema.GroupKind{Kind: kind, Group: "testdata.kubebuilder.io"}
				parser.NeedCRDFor(groupKind, nil)

				expectedErr := "field foo is part of OneOf constraint and must have omitempty or omitzero tag"
				Expect(packageErrors(pkgs[0])).To(MatchError(ContainSubstring(expectedErr)))
			})
		})

		Context("CronJob API with Wrong Annotation Format", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./wrong_annotation_format"}
				expPkgLen = 1
			})
			It("can not successfully generate the CronJob CRD", func() {
				assertError(pkgs[0], "CronJob", "is not in 'xxx=xxx' format")
			})
		})

		Context("Field with unvalid title format", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./wrong_title_format"}
				expPkgLen = 1
			})
			It("cannot generate title field from integer", func() {
				assertError(pkgs[0], "JobSpec", "expected string, got int")
			})
			It("cannot generate title field from map", func() {
				assertError(pkgs[0], "TestType", "expected string, got map[string]interface {}")
			})
		})

		Context("ExternalDoc API", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./external_docs/..."}
				expPkgLen = 1
			})
			It("should successfully generate the CRD with external documentation", func() {
				assertCRD(pkgs[0], "ExternalDoc", "testdata.kubebuilder.io_externaldocs.yaml")
			})
		})

		Context("InterfaceField API with any/interface{} types", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./iface/..."}
				expPkgLen = 1
			})
			It("should error on any type", func() {
				assertError(pkgs[0], "InterfaceField", "cannot generate schema for any")
			})
		})

		Context("CronJob API without group", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./nogroup"}
				expPkgLen = 1
			})
			It("should successfully generate the CronJob CRD", func() {
				assertCRDForGroupKind(pkgs[0], schema.GroupKind{Kind: "CronJob"}, "testdata._cronjobs.yaml")
			})
		})

		Context("Plugin API with fields of named types", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./rawmessage"}
				expPkgLen = 1
			})
			It("should generate json.RawMessage and named byte slices as base64-encoded strings", func() {
				assertCRD(pkgs[0], "Plugin", "testdata.kubebuilder.io_plugins.yaml")
			})
			It("should type fields of named types like the schema of the named type", func() {
				By("requesting the schema of the spec before it is flattened")
				typeIdent := crd.TypeIdent{Package: pkgs[0], Name: "PluginSpec"}
				parser.NeedSchemaFor(typeIdent)
				Expect(packageErrors(pkgs[0], packages.TypeError)).NotTo(HaveOccurred())
				spec, found := parser.Schemata[typeIdent]
				Expect(found).To(BeTrue(), "schema for PluginSpec should be generated")

				expectField := func(name, typ, format string) {
					field, found := spec.Properties[name]
					ExpectWithOffset(1, found).To(BeTrue(), "field %s should be generated", name)
					ExpectWithOffset(1, field.Ref).NotTo(BeNil(), "field %s should reference its named type", name)
					ExpectWithOffset(1, field.Type).To(Equal(typ), "unexpected type for field %s", name)
					ExpectWithOffset(1, field.Format).To(Equal(format), "unexpected format for field %s", name)
				}

				By("checking that byte slices are base64-encoded strings, not arrays")
				for _, name := range []string{"config", "caBundle", "clientCertificate", "authToken"} {
					expectField(name, "string", "byte")
				}

				By("checking that type markers on the named type win over its Go type")
				Expect(spec.Properties["values"].Type).To(Equal("object"))
				expectField("allowedHosts", "string", "")
				expectField("deniedHosts", "string", "")

				By("checking that custom text marshalling is a string")
				expectField("supportedVersions", "string", "")

				By("checking that other named slices and maps keep their type so field-level markers apply")
				for _, name := range []string{"tags", "keywords", "topics", "scopes"} {
					expectField(name, "array", "")
					Expect(spec.Properties[name].XListType).To(HaveValue(Equal("set")), "field %s should keep its list marker", name)
				}
				for _, name := range []string{"labels", "headers"} {
					expectField(name, "object", "")
					Expect(spec.Properties[name].XMapType).To(HaveValue(Equal("granular")), "field %s should keep its map marker", name)
				}

				By("checking that every field of a named type has the type and format of that type's schema")
				for name, field := range spec.Properties {
					if field.Ref == nil {
						continue
					}
					typeName, pkgPath, err := crd.RefParts(*field.Ref)
					Expect(err).NotTo(HaveOccurred())
					pkg := pkgs[0]
					if pkgPath != "" {
						pkg = pkgs[0].Imports()[pkgPath]
						Expect(pkg).NotTo(BeNil(), "package %s of field %s should be loaded", pkgPath, name)
					}
					named, found := parser.Schemata[crd.TypeIdent{Package: pkg, Name: typeName}]
					Expect(found).To(BeTrue(), "schema of %s for field %s should be generated", typeName, name)
					Expect(field.Type).To(Equal(named.Type), "field %s should have the type of %s", name, typeName)
					Expect(field.Format).To(Equal(named.Format), "field %s should have the format of %s", name, typeName)
				}
			})
		})

		Context("Extension API with a list marker on a json.RawMessage field", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./rawmessage_error"}
				expPkgLen = 1
			})
			It("should reject the list marker because a byte slice is not an array", func() {
				assertError(pkgs[0], "Extension", "must apply listType to an array, found string")
			})
		})

		Context("Type Alias API", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./typealias_indirect"}
				expPkgLen = 1
			})
			It("should handle types from indirectly imported packages via type aliases", func() {
				By("verifying no errors occurred during package loading and parsing")
				Expect(packageErrors(pkgs[0])).NotTo(HaveOccurred())

				By("requesting schema for type with indirect alias")
				typeIdent := crd.TypeIdent{
					Package: pkgs[0],
					Name:    "IndirectAliasSpec",
				}
				parser.NeedSchemaFor(typeIdent)

				By("verifying the schema includes properties from the base type")
				schema, found := parser.Schemata[typeIdent]
				Expect(found).To(BeTrue(), "schema for IndirectAliasSpec should be generated")
				Expect(schema.Properties).To(HaveKey("field"), "should have field property")
			})
		})
	})

	It("should generate plural words for Kind correctly", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/plural")).To(Succeed())
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		pkg := pkgs[0]

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(crdmarkers.Register(reg)).To(Succeed())
		parser := &crd.Parser{
			Collector: &markers.Collector{Registry: reg},
			Checker:   &loader.TypeChecker{},
		}
		crd.AddKnownTypes(parser)

		By("requesting that the package be parsed")
		parser.NeedPackage(pkg)

		By("requesting that the CRD be generated")
		groupKind := schema.GroupKind{Kind: "TestQuota", Group: "plural.example.com"}
		parser.NeedCRDFor(groupKind, nil)

		By("fixing top level ObjectMeta on the CRD")
		crd.FixTopLevelMetadata(parser.CustomResourceDefinitions[groupKind])

		By("loading the desired YAML")
		expectedFile, err := os.ReadFile("plural.example.com_testquotas.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("parsing the desired YAML")
		var crd apiextensionsv1.CustomResourceDefinition
		Expect(yaml.Unmarshal(expectedFile, &crd)).To(Succeed())
		// clear the annotations -- we don't care about the attribution annotation
		crd.Annotations = nil

		By("comparing the two")
		cmpOpts := cmpopts.EquateEmpty()
		Expect(cmp.Equal(parser.CustomResourceDefinitions[groupKind], crd, cmpOpts)).To(BeTrue(), "type not as expected, check pkg/crd/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(parser.CustomResourceDefinitions[groupKind], crd, cmpOpts))
	})

	It("should skip api internal package", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/internal_version")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		cronJobPkg := pkgs[0]

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(crdmarkers.Register(reg)).To(Succeed())
		parser := &crd.Parser{
			Collector: &markers.Collector{Registry: reg},
			Checker:   &loader.TypeChecker{},
		}
		crd.AddKnownTypes(parser)

		By("requesting that the package be parsed")
		parser.NeedPackage(cronJobPkg)

		By("checking that there is no GroupVersion")
		Expect(parser.GroupVersions).To(BeEmpty())

		By("checking that there are no Types")
		Expect(parser.Types).To(BeEmpty())

		By("checking that no errors occurred along the way (expect for type errors)")
		Expect(packageErrors(cronJobPkg, packages.TypeError)).NotTo(HaveOccurred())
	})

	It("should generate markers properly among several package versions", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/multiple_versions")).To(Succeed())
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots("./v1beta1", "./v1beta2")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(2))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(crdmarkers.Register(reg)).To(Succeed())
		parser := &crd.Parser{
			Collector: &markers.Collector{Registry: reg},
			Checker:   &loader.TypeChecker{},
		}
		crd.AddKnownTypes(parser)

		By("requesting that the package be parsed")
		for _, pkg := range pkgs {
			parser.NeedPackage(pkg)
		}

		By("requesting that the CRD be generated")
		groupKind := schema.GroupKind{Kind: "VersionedResource", Group: "testdata.kubebuilder.io"}
		parser.NeedCRDFor(groupKind, nil)

		By("fixing top level ObjectMeta on the CRD")
		crd.FixTopLevelMetadata(parser.CustomResourceDefinitions[groupKind])

		By("loading the desired YAML")
		expectedFile, err := os.ReadFile("testdata.kubebuilder.io_versionedresources.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("parsing the desired YAML")
		var crd apiextensionsv1.CustomResourceDefinition
		Expect(yaml.Unmarshal(expectedFile, &crd)).To(Succeed())
		// clear the annotations -- we don't care about the attribution annotation
		crd.Annotations = nil

		By("comparing the two")
		cmpOpts := cmpopts.EquateEmpty()
		Expect(cmp.Equal(parser.CustomResourceDefinitions[groupKind], crd, cmpOpts)).To(BeTrue(), "type not as expected, check pkg/crd/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(parser.CustomResourceDefinitions[groupKind], crd, cmpOpts))
	})
})
