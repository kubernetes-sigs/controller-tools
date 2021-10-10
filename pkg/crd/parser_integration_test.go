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
	"io/ioutil"
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
	It("should properly generate and flatten the rewritten CronJob schema", func() {
		// TODO(directxman12): test generation across multiple versions (right
		// now, we're trusting k/k's conversion code, though, which is probably
		// fine for the time being)
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots("./", "./unserved", "./deprecated")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(3))
		cronJobPkg := pkgs[0]

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(crdmarkers.Register(reg)).To(Succeed())
		parser := &crd.Parser{
			Collector: &markers.Collector{Registry: reg},
			Checker:   &loader.TypeChecker{},
		}
		crd.AddKnownTypes(parser)

		By("requesting that the packages be parsed")
		for _, pkg := range pkgs {
			parser.NeedPackage(pkg)
		}

		By("requesting that the CRD be generated")
		groupKind := schema.GroupKind{Kind: "CronJob", Group: "testdata.kubebuilder.io"}
		parser.NeedCRDFor(groupKind, nil)

		By("fixing top level ObjectMeta on the CRD")
		crd.FixTopLevelMetadata(parser.CustomResourceDefinitions[groupKind])

		By("checking that no errors occurred along the way (expect for type errors)")
		Expect(packageErrors(cronJobPkg, packages.TypeError)).NotTo(HaveOccurred())

		By("checking that the CRD is present")
		Expect(parser.CustomResourceDefinitions).To(HaveKey(groupKind))

		By("loading the desired YAML")
		expectedFile, err := ioutil.ReadFile("testdata.kubebuilder.io_cronjobs.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("parsing the desired YAML")
		var crd apiext.CustomResourceDefinition
		Expect(yaml.Unmarshal(expectedFile, &crd)).To(Succeed())
		// clear the annotations -- we don't care about the attribution annotation
		crd.Annotations = nil

		By("comparing the two")
		Expect(parser.CustomResourceDefinitions[groupKind]).To(Equal(crd), "type not as expected, check pkg/crd/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(parser.CustomResourceDefinitions[groupKind], crd))
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
		expectedFile, err := ioutil.ReadFile("plural.example.com_testquotas.yaml")
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
})
