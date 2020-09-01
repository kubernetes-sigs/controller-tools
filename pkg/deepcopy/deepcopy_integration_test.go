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
	"io"
	"io/ioutil"
	"os"
	"sort"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/deepcopy"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

type outputToMap map[string]*outputFile

// Open implements genall.OutputRule.
func (m outputToMap) Open(_ *loader.Package, path string) (io.WriteCloser, error) {
	if _, ok := m[path]; !ok {
		m[path] = &outputFile{}
	}
	return m[path], nil
}

func (m outputToMap) fileList() []string {
	ret := make([]string, 0, len(m))
	for path := range m {
		ret = append(ret, path)
	}
	sort.Strings(ret)
	return ret
}

type outputFile struct {
	contents []byte
}

func (o *outputFile) Write(p []byte) (int, error) {
	o.contents = append(o.contents, p...)
	return len(p), nil
}

func (o *outputFile) Close() error {
	return nil
}

var _ = Describe("CRD Generation From Parsing to CustomResourceDefinition", func() {
	It("should properly generate and flatten the rewritten CronJob schema", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		output := make(outputToMap)

		By("initializing the runtime")
		optionsRegistry := &markers.Registry{}
		Expect(optionsRegistry.Register(markers.Must(markers.MakeDefinition("crd", markers.DescribesPackage, crd.Generator{})))).To(Succeed())
		Expect(optionsRegistry.Register(markers.Must(markers.MakeDefinition("object", markers.DescribesPackage, deepcopy.Generator{})))).To(Succeed())
		rt, err := genall.FromOptions(optionsRegistry, []string{
			"crd", // Run another generator first to make sure they don't interfere; see also: the comment on cronjob_types.go:UntypedBlob
			"object",
		})
		Expect(err).NotTo(HaveOccurred())
		rt.OutputRules = genall.OutputRules{Default: output}

		By("running the generator and checking for errors")
		hadErrs := rt.Run()

		By("checking that we got output contents")
		Expect(output.fileList()).To(ContainElement("zz_generated.deepcopy.go")) // Don't use HaveKey--the output is too verbose to be usable
		outFile := output["zz_generated.deepcopy.go"]
		Expect(outFile).NotTo(BeNil())
		outContents := outFile.contents
		Expect(outContents).NotTo(BeNil())

		By("loading the desired code")
		expectedFile, err := ioutil.ReadFile("zz_generated.deepcopy.go")
		Expect(err).NotTo(HaveOccurred())

		By("comparing the two")
		Expect(string(outContents)).To(Equal(string(expectedFile)), "generated code not as expected, check pkg/deepcopy/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(outContents, expectedFile))

		By("checking for errors")
		Expect(hadErrs).To(BeFalse())
	})
})
