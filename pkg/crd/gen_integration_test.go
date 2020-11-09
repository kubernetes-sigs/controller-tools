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
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("CRD Generation proper defaulting", func() {
	var ctx *genall.GenerationContext
	var out *outputRule
	BeforeEach(func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/gen")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("settup up the context")
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
	})

	It("should strip v1beta1 CRDs of default fields", func() {
		By("calling Generate")
		gen := &crd.Generator{
			CRDVersions: []string{"v1beta1"},
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("loading the desired YAML")
		expectedFile, err := ioutil.ReadFile("./testdata/gen/foo_crd_v1beta1.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("comparing the two")
		Expect(out.buf.Bytes()).To(Equal(expectedFile))

	})

	It("should not strip v1 CRDs of default fields", func() {
		By("calling Generate")
		gen := &crd.Generator{
			CRDVersions: []string{"v1"},
		}
		Expect(gen.Generate(ctx)).NotTo(HaveOccurred())

		By("loading the desired YAML")
		expectedFile, err := ioutil.ReadFile("./testdata/gen/foo_crd_v1.yaml")
		Expect(err).NotTo(HaveOccurred())

		By("comparing the two")
		Expect(out.buf.Bytes()).To(Equal(expectedFile))

	})
})

type outputRule struct {
	buf *bytes.Buffer
}

func (o *outputRule) Open(_ *loader.Package, itemPath string) (io.WriteCloser, error) {
	return nopCloser{o.buf}, nil
}

type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error {
	return nil
}
