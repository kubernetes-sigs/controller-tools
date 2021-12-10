/*
Copyright 2021 The Kubernetes Authors.

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
	"path/filepath"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = DescribeTable("CRD Generation marker handling", testMarkers,
	TableEntry{
		Description: "skipDescriptions on field",
		Parameters: []interface{}{
			filepath.Join("skipdescriptions", "onfield"),
			"example.com_testresources.yaml",
		},
	},
	TableEntry{
		Description: "skipDescriptions on package",
		Parameters: []interface{}{
			filepath.Join("skipdescriptions", "onpackage"),
			"example.com_testresources.yaml",
		},
	},
)

// testMarkers is a Ginkgo It() body that runs the CRD generator in a directory and expects
// the output to match the contents of a file.
func testMarkers(testdataDirname, expectedFilename string) {
	workDir := filepath.Join("testdata", "markers", testdataDirname)

	By("switching into testdata to appease go modules")
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	Expect(os.Chdir(workDir)).To(Succeed()) // go modules are directory-sensitive
	defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

	By("loading the roots")
	pkgs, err := loader.LoadRoots(".")
	Expect(err).NotTo(HaveOccurred())
	Expect(pkgs).To(HaveLen(1))

	By("setup up the context")
	reg := &markers.Registry{}
	Expect(crdmarkers.Register(reg)).To(Succeed())
	out := &inMemoryOutput{}
	ctx := &genall.GenerationContext{
		Collector:  &markers.Collector{Registry: reg},
		Roots:      pkgs,
		Checker:    &loader.TypeChecker{},
		OutputRule: out,
	}

	By("calling Generate")
	gen := &crd.Generator{
		CRDVersions: []string{"v1"},
	}
	Expect(gen.Generate(ctx)).To(Succeed())

	By("loading the desired YAML")
	expectedFile, err := ioutil.ReadFile(expectedFilename)
	Expect(err).NotTo(HaveOccurred())
	expectedFile = bytes.Replace(expectedFile, []byte("(devel)"), []byte("(unknown)"), 1)

	By("comparing the output to the desired YAML")
	Expect(out.body).To(Equal(string(expectedFile)), cmp.Diff(out.body, string(expectedFile)))
}

var _ genall.OutputRule = &inMemoryOutput{}

type inMemoryOutput struct {
	body string
}

func (o *inMemoryOutput) Open(*loader.Package, string) (io.WriteCloser, error) {
	return &bufferedInMemoryWriter{inMemoryOutput: o}, nil
}

type bufferedInMemoryWriter struct {
	bytes.Buffer
	inMemoryOutput *inMemoryOutput
}

func (w *bufferedInMemoryWriter) Close() error {
	w.inMemoryOutput.body = w.Buffer.String()
	return nil
}
