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

package applyconfigurations

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"

	"sigs.k8s.io/controller-tools/pkg/crd"
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
	It("should be able to verify generated ApplyConfiguration types for the CronJob schema", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/cronjob")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		output := make(outputToMap)

		By("initializing the runtime")
		optionsRegistry := &markers.Registry{}
		Expect(optionsRegistry.Register(markers.Must(markers.MakeDefinition("crd", markers.DescribesPackage, crd.Generator{})))).To(Succeed())
		// Add the applyconfigurations generator but set it to verify only.
		// This allows us to check if there's a diff between the checked in generated data and what it would generate now.
		Expect(optionsRegistry.Register(markers.Must(markers.MakeDefinition("apply", markers.DescribesPackage, Generator{})))).To(Succeed())
		rt, err := genall.FromOptions(optionsRegistry, []string{
			"crd", // Run another generator first to make sure they don't interfere; see also: the comment on cronjob_types.go:UntypedBlob
			"apply",
		})
		Expect(err).NotTo(HaveOccurred())
		rt.OutputRules = genall.OutputRules{Default: output}

		By("running the generator and checking for errors")
		hadErrs := rt.Run()

		By("checking for errors")
		Expect(hadErrs).To(BeFalse(), "Generator should run without errors")

		filesInMaster := make(map[string][]byte)
		masterFileNames := sets.New[string]()
		cronJobFS := os.DirFS(".")
		masterPath := "applyconfiguration-master"
		Expect(fs.WalkDir(cronJobFS, masterPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", path, err)
			}

			// Record the path without the path prefix for comparison later.
			path = strings.TrimPrefix(path, masterPath+"/")
			masterFileNames.Insert(path)
			filesInMaster[path] = data
			return nil
		})).To(Succeed())

		filesInOutput := make(map[string][]byte)
		outputFileNames := sets.New[string]()
		outputPath := "applyconfiguration"
		Expect(fs.WalkDir(cronJobFS, outputPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", path, err)
			}

			// Record the path without the path prefix for comparison later.
			path = strings.TrimPrefix(path, outputPath+"/")
			outputFileNames.Insert(path)
			filesInOutput[path] = data
			return nil
		})).To(Succeed())

		// Every file should be in both sets, check for files not in both sets.
		Expect(masterFileNames.SymmetricDifference(outputFileNames).UnsortedList()).To(BeEmpty(), "Generated files should match the checked in files")

		for name, content := range filesInMaster {
			Expect(string(filesInOutput[name])).To(Equal(string(content)), "Generated files should match the checked in files, diff found in %s", name)
		}
	})
})
