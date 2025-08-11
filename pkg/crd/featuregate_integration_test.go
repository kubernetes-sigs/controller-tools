/*
Copyright 2025.

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
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("CRD Feature Gate Generation", func() {
	var (
		ctx                *genall.GenerationContext
		out                *featureGateOutputRule
		featureGateDir     string
		originalWorkingDir string
	)

	BeforeEach(func() {
		var err error
		originalWorkingDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		featureGateDir = filepath.Join(originalWorkingDir, "testdata", "featuregates")

		By("switching into featuregates testdata")
		err = os.Chdir(featureGateDir)
		Expect(err).NotTo(HaveOccurred())

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		out = &featureGateOutputRule{buf: &bytes.Buffer{}}
		ctx = &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: &markers.Registry{}},
			Roots:      pkgs,
			Checker:    &loader.TypeChecker{},
			OutputRule: out,
		}
		Expect(crdmarkers.Register(ctx.Collector.Registry)).To(Succeed())
	})

	AfterEach(func() {
		By("restoring original working directory")
		err := os.Chdir(originalWorkingDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should not include feature-gated fields when no gates are enabled", func() {
		By("calling the generator")
		gen := &crd.Generator{
			CRDVersions: []string{"v1"},
			// No FeatureGates specified
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("loading the expected YAML")
		expectedFile, err := os.ReadFile("output_none/_featuregatetests.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("comparing the two")
		Expect(out.buf.String()).To(Equal(string(expectedFile)), cmp.Diff(out.buf.String(), string(expectedFile)))
	})

	It("should include only alpha-gated fields when alpha gate is enabled", func() {
		By("calling the generator")
		gen := &crd.Generator{
			CRDVersions:  []string{"v1"},
			FeatureGates: "alpha=true",
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("loading the expected YAML")
		expectedFile, err := os.ReadFile("output_alpha/_featuregatetests.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("comparing the two")
		Expect(out.buf.String()).To(Equal(string(expectedFile)), cmp.Diff(out.buf.String(), string(expectedFile)))
	})

	It("should include only beta-gated fields when beta gate is enabled", func() {
		By("calling the generator")
		gen := &crd.Generator{
			CRDVersions:  []string{"v1"},
			FeatureGates: "beta=true",
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("loading the expected YAML")
		expectedFile, err := os.ReadFile("output_beta/_featuregatetests.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("comparing the two")
		Expect(out.buf.String()).To(Equal(string(expectedFile)), cmp.Diff(out.buf.String(), string(expectedFile)))
	})

	It("should include both feature-gated fields when both gates are enabled", func() {
		By("calling the generator")
		gen := &crd.Generator{
			CRDVersions:  []string{"v1"},
			FeatureGates: "alpha=true,beta=true",
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("loading the expected YAML")
		expectedFile, err := os.ReadFile("output_both/_featuregatetests.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("comparing the two")
		Expect(out.buf.String()).To(Equal(string(expectedFile)), cmp.Diff(out.buf.String(), string(expectedFile)))
	})

	It("should handle complex precedence: (alpha&beta)|gamma", func() {
		By("calling the generator with only gamma enabled")
		gen := &crd.Generator{
			CRDVersions:  []string{"v1"},
			FeatureGates: "gamma=true",
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("loading the expected YAML")
		expectedFile, err := os.ReadFile("output_gamma/_featuregatetests.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("comparing the two")
		Expect(out.buf.String()).To(Equal(string(expectedFile)), cmp.Diff(out.buf.String(), string(expectedFile)))
	})

	It("should include all fields when all gates are enabled", func() {
		By("calling the generator with all gates enabled")
		gen := &crd.Generator{
			CRDVersions:  []string{"v1"},
			FeatureGates: "alpha=true,beta=true,gamma=true",
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("loading the expected YAML")
		expectedFile, err := os.ReadFile("output_all/_featuregatetests.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("comparing the two")
		Expect(out.buf.String()).To(Equal(string(expectedFile)), cmp.Diff(out.buf.String(), string(expectedFile)))
	})
})

// Helper types for testing
type featureGateOutputRule struct {
	buf *bytes.Buffer
}

func (o *featureGateOutputRule) Open(_ *loader.Package, _ string) (io.WriteCloser, error) {
	return featureGateNopCloser{o.buf}, nil
}

type featureGateNopCloser struct {
	io.Writer
}

func (n featureGateNopCloser) Close() error {
	return nil
}
