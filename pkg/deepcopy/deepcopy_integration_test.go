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

package deepcopy_test

import (
	"io/ioutil"
	"os"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"

	"sigs.k8s.io/controller-tools/pkg/deepcopy"
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
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		cronJobPkg := pkgs[0]

		By("setting up the generation context")
		collector := &markers.Collector{Registry: &markers.Registry{}}
		Expect(deepcopy.Generator{}.RegisterMarkers(collector.Registry)).To(Succeed())
		checker := &loader.TypeChecker{}
		ctx := &deepcopy.ObjectGenCtx{
			Collector: collector,
			Checker:   checker,
		}

		By("requesting that types be generated")
		outContents := ctx.GenerateForPackage(cronJobPkg)

		By("checking that no errors occurred along the way (expect for type errors)")
		Expect(packageErrors(cronJobPkg, packages.TypeError)).NotTo(HaveOccurred())

		By("checking that we got output contents")
		Expect(outContents).NotTo(BeNil())

		By("loading the desired code")
		expectedFile, err := ioutil.ReadFile("zz_generated.deepcopy.go")
		Expect(err).NotTo(HaveOccurred())

		By("comparing the two")
		Expect(outContents).To(Equal(expectedFile), "generated code not as expected, check pkg/deepcopy/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(outContents, expectedFile))
	})
})
