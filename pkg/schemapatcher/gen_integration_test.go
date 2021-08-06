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

package schemapatcher_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-tools/pkg/genall"
	. "sigs.k8s.io/controller-tools/pkg/schemapatcher"
)

var _ = Describe("CRD Patching From Parsing to Editing", func() {
	It("should fail to load legacy CRDs", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the generation runtime")
		var crdSchemaGen genall.Generator = &Generator{
			ManifestsPath: "./invalid",
		}
		rt, err := genall.Generators{&crdSchemaGen}.ForRoots("./...")
		Expect(err).NotTo(HaveOccurred())

		outputDir, err := ioutil.TempDir("", "controller-tools-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		rt.OutputRules.Default = genall.OutputToDirectory(outputDir)

		By("running the generator")
		Expect(rt.Run()).To(BeTrue(), "unexpectedly succeeded")
	})

	It("should properly generate and patch the test CRDs", func() {
		// TODO(directxman12): I've ported these over from @sttts's tests,
		// but they should really probably not be writing to an actual filesystem

		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the generation runtime")
		var crdSchemaGen genall.Generator = &Generator{
			ManifestsPath: "./valid",
		}
		rt, err := genall.Generators{&crdSchemaGen}.ForRoots("./...")
		Expect(err).NotTo(HaveOccurred())

		outputDir, err := ioutil.TempDir("", "controller-tools-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		rt.OutputRules.Default = genall.OutputToDirectory(outputDir)
		rt.ErrorWriter = GinkgoWriter

		By("running the generator")
		Expect(rt.Run()).To(BeFalse(), "unexpectedly had errors")

		By("loading the output files")
		expectedFiles, err := ioutil.ReadDir("expected")
		Expect(err).NotTo(HaveOccurred())

		for _, expectedFile := range expectedFiles {
			By("reading the expected and actual files for " + expectedFile.Name())
			actualContents, err := ioutil.ReadFile(filepath.Join(outputDir, expectedFile.Name()))
			Expect(err).NotTo(HaveOccurred())

			expectedContents, err := ioutil.ReadFile(filepath.Join("expected", expectedFile.Name()))
			Expect(err).NotTo(HaveOccurred())

			By("checking that the expected and actual files for " + expectedFile.Name() + " are identical")
			Expect(actualContents).To(Equal(expectedContents), "contents not as expected, check pkg/schemapatcher/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(string(actualContents), string(expectedContents)))
		}
	})
})
