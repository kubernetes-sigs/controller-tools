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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/genall"
)

type crdGenParams struct {
	// rootPaths is the relative path to the CRD go models
	rootPaths string
	// expectedFile is the relative path to the expected CRD yaml file
	expectedFile string
	// generator to use
	generator crd.Generator
}

var _ = Describe("CRD generation", func() {

	Context("when there is a single api version", func() {
		Context("generate cronjob v1", func() {
			Context("using crd v1beta1", func() {
				testCrdGeneration(crdGenParams{
					rootPaths:    "./cronjob/v1/...",
					expectedFile: "./cronjob/v1/cronjob_crdv1beta1.yaml",
					generator: crd.Generator{
						CRDVersions: []string{"v1beta1"},
					},
				})
			})
			Context("using crd v1", func() {
				testCrdGeneration(crdGenParams{
					rootPaths:    "./cronjob/v1/...",
					expectedFile: "./cronjob/v1/cronjob_crdv1.yaml",
					generator: crd.Generator{
						CRDVersions: []string{"v1"},
					},
				})
			})
		})
		Context("generate cronjob v2", func() {
			Context("using crd v1beta1", func() {
				testCrdGeneration(crdGenParams{
					rootPaths:    "./cronjob/v2/...",
					expectedFile: "./cronjob/v2/cronjob_crdv1beta1.yaml",
					generator: crd.Generator{
						CRDVersions: []string{"v1beta1"},
					},
				})
			})
			Context("using crd v1", func() {
				testCrdGeneration(crdGenParams{
					rootPaths:    "./cronjob/v2/...",
					expectedFile: "./cronjob/v2/cronjob_crdv1.yaml",
					generator: crd.Generator{
						CRDVersions: []string{"v1"},
					},
				})
			})
		})
	})

	Context("when there are multiple api versions (cronjob v1 + v2)", func() {
		rootPaths := "./cronjob/..."
		expectedYamlPath := func(crdVersion string, feature string) string {
			if feature != "" {
				feature = fmt.Sprintf("_%s", feature)
			}
			return fmt.Sprintf("./cronjob/cronjob_crd%s%s.yaml", crdVersion, feature)
		}

		Context("using crd v1", func() {
			crdVersion := "v1"
			Context("with default options", func() {
				testCrdGeneration(crdGenParams{
					rootPaths:    rootPaths,
					expectedFile: expectedYamlPath(crdVersion, ""),
					generator: crd.Generator{
						CRDVersions: []string{crdVersion},
					},
				})
			})
			Context("with maxDescLen=10", func() {
				maxDescLen := 10
				testCrdGeneration(crdGenParams{
					rootPaths:    rootPaths,
					expectedFile: expectedYamlPath(crdVersion, "maxdesclen"),
					generator: crd.Generator{
						CRDVersions: []string{crdVersion},
						MaxDescLen:  &maxDescLen,
					},
				})
			})
		})

		Context("using crd v1 beta1", func() {
			crdVersion := "v1beta1"
			Context("with default options", func() {
				testCrdGeneration(crdGenParams{
					rootPaths:    rootPaths,
					expectedFile: expectedYamlPath(crdVersion, ""),
					generator: crd.Generator{
						CRDVersions: []string{crdVersion},
					},
				})
			})
			Context("with maxDescLen=10", func() {
				maxDescLen := 10
				testCrdGeneration(crdGenParams{
					rootPaths:    rootPaths,
					expectedFile: expectedYamlPath(crdVersion, "maxdesclen"),
					generator: crd.Generator{
						CRDVersions: []string{crdVersion},
						MaxDescLen:  &maxDescLen,
					},
				})
			})
			Context("with trivialVersions=true", func() {
				testCrdGeneration(crdGenParams{
					rootPaths:    rootPaths,
					expectedFile: expectedYamlPath(crdVersion, "trivial"),
					generator: crd.Generator{
						CRDVersions:     []string{crdVersion},
						TrivialVersions: true,
					},
				})
			})
			Context("with preserveUnknownFields=false", func() {
				preserveUnknownFields := false
				testCrdGeneration(crdGenParams{
					rootPaths:    rootPaths,
					expectedFile: expectedYamlPath(crdVersion, "preserve_false"),
					generator: crd.Generator{
						CRDVersions:           []string{crdVersion},
						PreserveUnknownFields: &preserveUnknownFields,
					},
				})
			})
		})

		Context("when crdVersions=v1;v1beta1", func() {
			testCrdGeneration(crdGenParams{
				rootPaths:    rootPaths,
				expectedFile: expectedYamlPath("v1v1beta1", ""),
				generator: crd.Generator{
					CRDVersions: []string{"v1", "v1beta1"},
				},
			})
		})

		Context("when crdVersions is empty (default to v1beta1)", func() {
			testCrdGeneration(crdGenParams{
				rootPaths:    rootPaths,
				expectedFile: expectedYamlPath("v1beta1", ""),
				generator: crd.Generator{
					CRDVersions: []string{},
				},
			})
		})
	})
})

func testCrdGeneration(params crdGenParams) {
	It("should generate the expected CRD", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("initializing the generator")
		var gen genall.Generator = params.generator
		rt, err := genall.Generators{&gen}.ForRoots(params.rootPaths)
		Expect(err).NotTo(HaveOccurred())

		outputDir, err := ioutil.TempDir("", "controller-tools-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		var buf bytes.Buffer
		rt.OutputRules.Default = genall.OutputToBuffer{Buffer: &buf}

		By("running the generator")
		Expect(rt.Run()).To(BeFalse(), "unexpectedly had errors")

		By("loading the actual and expected outputs")
		actualStr := buf.String()
		expected, err := ioutil.ReadFile(params.expectedFile)
		Expect(err).NotTo(HaveOccurred())
		expectedStr := string(expected)

		By("patching version annotation in the actual content")
		// patch the version annotation that appears as "(unknown)" in the test-generated content
		Expect(actualStr).To(ContainSubstring("(unknown)"))
		actualStr = strings.ReplaceAll(actualStr, "(unknown)", "(devel)")

		By("comparing actual and expected CRD output")
		Expect(actualStr).To(Equal(expectedStr), "contents not as expected\n\nDiff:\n\n%s", cmp.Diff(actualStr, expectedStr))
	})
}
