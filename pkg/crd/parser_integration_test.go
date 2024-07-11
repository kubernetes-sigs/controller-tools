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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
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
			var crd apiext.CustomResourceDefinition
			ExpectWithOffset(1, yaml.Unmarshal(expectedFile, &crd)).To(Succeed())
			// clear the version annotation -- we don't care about the attribution annotation
			delete(crd.Annotations, "controller-gen.kubebuilder.io/version")
			if len(crd.Annotations) == 0 {
				crd.Annotations = nil
			}

			By(fmt.Sprintf("comparing the two %s CRDs", kind))
			ExpectWithOffset(1, parser.CustomResourceDefinitions[groupKind]).To(Equal(crd), "type not as expected, check pkg/crd/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(parser.CustomResourceDefinitions[groupKind], crd))
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

		Context("CronJob API with Wrong Annotation Format", func() {
			BeforeEach(func() {
				pkgPaths = []string{"./wrong_annotation_format"}
				expPkgLen = 1
			})
			It("can not successfully generate the CronJob CRD", func() {
				assertError(pkgs[0], "CronJob", "is not in 'xxx=xxx' format")
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
		var crd apiext.CustomResourceDefinition
		Expect(yaml.Unmarshal(expectedFile, &crd)).To(Succeed())
		// clear the annotations -- we don't care about the attribution annotation
		crd.Annotations = nil

		By("comparing the two")
		Expect(parser.CustomResourceDefinitions[groupKind]).To(Equal(crd), "type not as expected, check pkg/crd/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(parser.CustomResourceDefinitions[groupKind], crd))
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
		var crd apiext.CustomResourceDefinition
		Expect(yaml.Unmarshal(expectedFile, &crd)).To(Succeed())
		// clear the annotations -- we don't care about the attribution annotation
		crd.Annotations = nil

		By("comparing the two")
		Expect(parser.CustomResourceDefinitions[groupKind]).To(Equal(crd), "type not as expected, check pkg/crd/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(parser.CustomResourceDefinitions[groupKind], crd))
	})

})
