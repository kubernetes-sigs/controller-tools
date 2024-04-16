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

var _ = Describe("CRD Generation proper defaulting", func() {
	var (
		ctx, ctx2 *genall.GenerationContext
		out       *outputRule

		genDir = filepath.Join("testdata", "gen")
	)

	BeforeEach(func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(genDir)).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		pkgs2, err := loader.LoadRoots("./...")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs2).To(HaveLen(2))

		By("setup up the context")
		reg := &markers.Registry{}
		Expect(crdmarkers.Register(reg)).To(Succeed())
		out = &outputRule{
			buf: &bytes.Buffer{},
		}
		ctx = &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			Checker:    &loader.TypeChecker{},
			OutputRule: out,
		}
		ctx2 = &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs2,
			Checker:    &loader.TypeChecker{},
			OutputRule: out,
		}
	})

	It("should fail to generate v1beta1 CRDs", func() {
		By("calling Generate")
		gen := &crd.Generator{
			CRDVersions: []string{"v1beta1"},
		}
		Expect(gen.Generate(ctx)).To(MatchError(`apiVersion "apiextensions.k8s.io/v1beta1" is not supported`))
	})

	It("should not strip v1 CRDs of default fields and metadata description", func() {
		By("calling Generate")
		gen := &crd.Generator{
			CRDVersions: []string{"v1"},
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("loading the desired YAML")
		expectedFile, err := os.ReadFile(filepath.Join(genDir, "bar.example.com_foos.yaml"))
		Expect(err).NotTo(HaveOccurred())
		expectedFile = fixAnnotations(expectedFile)

		By("comparing the two")
		Expect(out.buf.String()).To(Equal(string(expectedFile)), cmp.Diff(out.buf.String(), string(expectedFile)))
	})

	It("should have deterministic output", func() {
		By("calling Generate on multple packages")
		gen := &crd.Generator{
			CRDVersions: []string{"v1"},
		}
		Expect(gen.Generate(ctx2)).NotTo(HaveOccurred())

		By("loading the desired YAMLs")
		expectedFileFoos, err := os.ReadFile(filepath.Join(genDir, "bar.example.com_foos.yaml"))
		Expect(err).NotTo(HaveOccurred())
		expectedFileFoos = fixAnnotations(expectedFileFoos)
		expectedFileZoos, err := os.ReadFile(filepath.Join(genDir, "zoo", "bar.example.com_zoos.yaml"))
		Expect(err).NotTo(HaveOccurred())
		expectedFileZoos = fixAnnotations(expectedFileZoos)

		By("comparing the two, output must be deterministic because groupKinds are sorted")
		expectedOut := string(expectedFileFoos) + string(expectedFileZoos)
		Expect(out.buf.String()).To(Equal(expectedOut), cmp.Diff(out.buf.String(), expectedOut))
	})

	It("should add preserveUnknownFields=false when specified", func() {
		By("calling Generate")
		no := false
		gen := &crd.Generator{
			CRDVersions: []string{"v1"},
			DeprecatedV1beta1CompatibilityPreserveUnknownFields: &no,
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("searching preserveUnknownFields")
		Expect(out.buf.String()).To(ContainSubstring("preserveUnknownFields: false"))
	})

	It("should add preserveUnknownFields=true when specified", func() {
		By("calling Generate")
		yes := true
		gen := &crd.Generator{
			CRDVersions: []string{"v1"},
			DeprecatedV1beta1CompatibilityPreserveUnknownFields: &yes,
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("searching preserveUnknownFields")
		Expect(out.buf.String()).To(ContainSubstring("preserveUnknownFields: true"))
	})

	It("should not add preserveUnknownFields when not specified", func() {
		By("calling Generate")
		gen := &crd.Generator{
			CRDVersions: []string{"v1"},
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("searching preserveUnknownFields")
		Expect(out.buf.String()).NotTo(ContainSubstring("preserveUnknownFields"))
	})
})

// fixAnnotations fixes the attribution annotation for tests.
func fixAnnotations(crdBytes []byte) []byte {
	return bytes.Replace(crdBytes, []byte("(devel)"), []byte("(unknown)"), 1)
}

type outputRule struct {
	buf *bytes.Buffer
}

func (o *outputRule) Open(_ *loader.Package, _ string) (io.WriteCloser, error) {
	return nopCloser{o.buf}, nil
}

type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error {
	return nil
}
