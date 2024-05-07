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

var _ = Describe("CRD Generation golang files", func() {
	var (
		ctx, ctx2 *genall.GenerationContext
		out       *outputRule

		genDir = filepath.Join("testdata", "multiple_files")
	)

	BeforeEach(func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(genDir)).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots("file_one.go")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		Expect(pkgs[0].GoFiles).To(HaveLen(1))
		pkgs2, err := loader.LoadRoots("file_two.go", "file_two_reference.go")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs2).To(HaveLen(1))
		Expect(pkgs2[0].GoFiles).To(HaveLen(2))

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

	It("should have deterministic output for single golang file", func() {
		By("calling Generate on single golang file")
		gen := &crd.Generator{
			CRDVersions: []string{"v1"},
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("loading the desired YAML")
		expectedFileOne, err := os.ReadFile(filepath.Join(genDir, "one.example.com.yaml"))
		Expect(err).NotTo(HaveOccurred())
		expectedFileOne = fixAnnotations(expectedFileOne)
		expectedOut := string(expectedFileOne)
		Expect(out.buf.String()).To(Equal(expectedOut), cmp.Diff(out.buf.String(), expectedOut))
	})
	It("should have deterministic output for multiple golang files referencing other types", func() {
		By("calling Generate on two golang files")
		gen := &crd.Generator{
			CRDVersions: []string{"v1"},
		}
		Expect(gen.Generate(ctx2)).NotTo(HaveOccurred())

		By("loading the desired YAML file")
		expectedFileTwo, err := os.ReadFile(filepath.Join(genDir, "two.example.com.yaml"))
		Expect(err).NotTo(HaveOccurred())
		expectedFileTwo = fixAnnotations(expectedFileTwo)
		expectedOut := string(expectedFileTwo)
		Expect(out.buf.String()).To(Equal(expectedOut), cmp.Diff(out.buf.String(), expectedOut))
	})
})
