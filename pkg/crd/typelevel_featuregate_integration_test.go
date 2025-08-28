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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("CRD Type-Level Feature Gates", func() {
	var (
		ctx                *genall.GenerationContext
		out                *typeLevelFeatureGateOutputRule
		typeLevelDir       string
		originalWorkingDir string
	)

	BeforeEach(func() {
		var err error
		originalWorkingDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		typeLevelDir = filepath.Join(originalWorkingDir, "testdata", "typelevelfeaturegates")

		By("switching into typelevelfeaturegates testdata")
		err = os.Chdir(typeLevelDir)
		Expect(err).NotTo(HaveOccurred())

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		out = &typeLevelFeatureGateOutputRule{buf: &bytes.Buffer{}}
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

	It("should only generate the always-on CRD when no feature gates are enabled", func() {
		By("calling the generator")
		gen := &crd.Generator{
			CRDVersions: []string{"v1"},
			// No FeatureGates specified
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("checking that only AlwaysOn CRD was generated")
		output := out.buf.String()
		Expect(output).To(ContainSubstring("kind: AlwaysOn"))
		Expect(output).NotTo(ContainSubstring("kind: AlphaGated"))
		Expect(output).NotTo(ContainSubstring("kind: BetaGated"))
		Expect(output).NotTo(ContainSubstring("kind: OrGated"))
		Expect(output).NotTo(ContainSubstring("kind: AndGated"))
		Expect(output).NotTo(ContainSubstring("kind: ComplexGated"))
	})

	It("should generate alpha-gated and OR-gated CRDs when alpha gate is enabled", func() {
		By("calling the generator")
		gen := &crd.Generator{
			CRDVersions:  []string{"v1"},
			FeatureGates: "alpha=true",
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("checking the generated CRDs")
		output := out.buf.String()
		Expect(output).To(ContainSubstring("kind: AlwaysOn"))
		Expect(output).To(ContainSubstring("kind: AlphaGated"))
		Expect(output).To(ContainSubstring("kind: OrGated")) // alpha|beta satisfied
		Expect(output).NotTo(ContainSubstring("kind: BetaGated"))
		Expect(output).NotTo(ContainSubstring("kind: AndGated"))     // alpha&beta not satisfied
		Expect(output).NotTo(ContainSubstring("kind: ComplexGated")) // (alpha&beta)|gamma not satisfied
	})

	It("should generate beta-gated and OR-gated CRDs when beta gate is enabled", func() {
		By("calling the generator")
		gen := &crd.Generator{
			CRDVersions:  []string{"v1"},
			FeatureGates: "beta=true",
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("checking the generated CRDs")
		output := out.buf.String()
		Expect(output).To(ContainSubstring("kind: AlwaysOn"))
		Expect(output).To(ContainSubstring("kind: BetaGated"))
		Expect(output).To(ContainSubstring("kind: OrGated")) // alpha|beta satisfied
		Expect(output).NotTo(ContainSubstring("kind: AlphaGated"))
		Expect(output).NotTo(ContainSubstring("kind: AndGated"))     // alpha&beta not satisfied
		Expect(output).NotTo(ContainSubstring("kind: ComplexGated")) // (alpha&beta)|gamma not satisfied
	})

	It("should generate all applicable CRDs when both alpha and beta gates are enabled", func() {
		By("calling the generator")
		gen := &crd.Generator{
			CRDVersions:  []string{"v1"},
			FeatureGates: "alpha=true,beta=true",
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("checking the generated CRDs")
		output := out.buf.String()
		Expect(output).To(ContainSubstring("kind: AlwaysOn"))
		Expect(output).To(ContainSubstring("kind: AlphaGated"))
		Expect(output).To(ContainSubstring("kind: BetaGated"))
		Expect(output).To(ContainSubstring("kind: OrGated"))      // alpha|beta satisfied
		Expect(output).To(ContainSubstring("kind: AndGated"))     // alpha&beta satisfied
		Expect(output).To(ContainSubstring("kind: ComplexGated")) // (alpha&beta)|gamma satisfied
	})

	It("should generate complex-gated CRD when gamma gate is enabled", func() {
		By("calling the generator")
		gen := &crd.Generator{
			CRDVersions:  []string{"v1"},
			FeatureGates: "gamma=true",
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("checking the generated CRDs")
		output := out.buf.String()
		Expect(output).To(ContainSubstring("kind: AlwaysOn"))
		Expect(output).To(ContainSubstring("kind: ComplexGated")) // (alpha&beta)|gamma satisfied by gamma
		Expect(output).NotTo(ContainSubstring("kind: AlphaGated"))
		Expect(output).NotTo(ContainSubstring("kind: BetaGated"))
		Expect(output).NotTo(ContainSubstring("kind: OrGated"))
		Expect(output).NotTo(ContainSubstring("kind: AndGated"))
	})

	It("should handle invalid feature gate expressions gracefully", func() {
		By("calling the generator with invalid expression")
		gen := &crd.Generator{
			CRDVersions:  []string{"v1"},
			FeatureGates: "invalid-syntax===true",
		}
		err := gen.Generate(ctx)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid feature gates"))
	})
})

// typeLevelFeatureGateOutputRule implements genall.OutputRule for capturing generated YAML
type typeLevelFeatureGateOutputRule struct {
	buf *bytes.Buffer
}

func (o *typeLevelFeatureGateOutputRule) Open(_ *loader.Package, _ string) (io.WriteCloser, error) {
	return typeLevelNopCloser{o.buf}, nil
}

type typeLevelNopCloser struct {
	io.Writer
}

func (n typeLevelNopCloser) Close() error {
	return nil
}

var _ genall.OutputRule = &typeLevelFeatureGateOutputRule{}
