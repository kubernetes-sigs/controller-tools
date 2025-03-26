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

package applyconfiguration

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"

	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

const (
	cronjobDir            = "./testdata/cronjob"
	applyConfigurationDir = "applyconfiguration"
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

var _ = Describe("ApplyConfiguration generation from API types", func() {
	var originalCWD string

	BeforeEach(func() {
		var tmpDir string

		By("Setting up a temporary directory", func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "applyconfiguration-integration-test")
			Expect(err).NotTo(HaveOccurred(), "Should be able to create a temporary directory")

			// Copy the testdata directory, but removed the generated files.
			Expect(os.CopyFS(tmpDir, os.DirFS(cronjobDir))).To(Succeed(), "Should be able to copy source files")
			Expect(os.RemoveAll(filepath.Join(tmpDir, applyConfigurationDir))).To(Succeed(), "Should be able to remove generated file from temp directory")
		})

		By("Switching into testdata to appease go modules", func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			originalCWD = cwd

			Expect(os.Chdir(tmpDir)).To(Succeed()) // go modules are directory-sensitive
		})

		By(fmt.Sprintf("Completed set up in %s", tmpDir))
	})

	AfterEach(func() {
		// Reset the working directory
		Expect(os.Chdir(originalCWD)).To(Succeed())
	})

	DescribeTable("should be able to verify generated ApplyConfiguration types for the CronJob schema", func(outputPackage string) {
		Expect(replaceOutputPkgMarker(".", outputPackage)).To(Succeed())

		// The output is used to capture the generated CRD file.
		// The output of the applyconfiguration cannot be generated to memory, gengo handles all of the writing to disk directly.
		output := make(outputToMap)

		By("Initializing the runtime")
		optionsRegistry := &markers.Registry{}
		Expect(optionsRegistry.Register(markers.Must(markers.MakeDefinition("crd", markers.DescribesPackage, crd.Generator{})))).To(Succeed())
		Expect(optionsRegistry.Register(markers.Must(markers.MakeDefinition("applyconfiguration", markers.DescribesPackage, Generator{})))).To(Succeed())

		rt, err := genall.FromOptions(optionsRegistry, []string{
			"crd", // Run another generator first to make sure they don't interfere; see also: the comment on cronjob_types.go:UntypedBlob
			"applyconfiguration",
		})
		Expect(err).NotTo(HaveOccurred())

		rt.OutputRules = genall.OutputRules{Default: output}

		originalFS := os.DirFS(filepath.Join(originalCWD, cronjobDir))
		tmpFS := os.DirFS(".")

		By("Running the generator")
		hadErrs := rt.Run()

		By("Checking for generation errors")
		Expect(hadErrs).To(BeFalse(), "Generator should run without errors")

		filesInOriginal := make(map[string][]byte)
		originalFileNames := sets.New[string]()
		Expect(fs.WalkDir(originalFS, applyConfigurationDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			data, err := os.ReadFile(filepath.Join(originalCWD, cronjobDir, path))
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", path, err)
			}

			// Record the path without the path prefix for comparison later.
			path = strings.TrimPrefix(path, applyConfigurationDir+"/")
			originalFileNames.Insert(path)
			filesInOriginal[path] = data
			return nil
		})).To(Succeed())

		filesInOutput := make(map[string][]byte)
		outputFileNames := sets.New[string]()
		Expect(fs.WalkDir(tmpFS, outputPackage, func(path string, d fs.DirEntry, err error) error {
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
			path = strings.TrimPrefix(path, outputPackage+"/")
			outputFileNames.Insert(path)
			filesInOutput[path] = data
			return nil
		})).To(Succeed())

		// // Every file should be in both sets, check for files not in both sets.
		Expect(outputFileNames.UnsortedList()).To(ConsistOf(originalFileNames.UnsortedList()), "Generated files should match the checked in files")

		for name, content := range filesInOriginal {
			// Make sure the package string is correct for the newly generated content.
			content = []byte(strings.Replace(string(content), "package applyconfiguration", fmt.Sprintf("package %s", outputPackage), 1))

			// Make sure the import paths are correct for the newly generated content.
			content = []byte(strings.ReplaceAll(string(content), "testdata/cronjob/applyconfiguration", fmt.Sprintf("testdata/cronjob/%s", outputPackage)))

			Expect(string(filesInOutput[name])).To(BeComparableTo(string(content)), "Generated files should match the checked in files, diff found in %s", name)
		}
	},
		Entry("with the default applyconfiguration output package", "applyconfiguration"),
		Entry("with the an alternative output package", "other"),
	)
})

func replaceOutputPkgMarker(dir string, newOutputPackage string) error {
	f, err := os.Open(filepath.Join(dir, "groupversion_info.go"))
	if err != nil {
		return fmt.Errorf("error opening groupversion_info.go: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading groupversion_info.go: %w", err)
	}

	newData := strings.Replace(string(data), "// +kubebuilder:ac:output:package=\"applyconfiguration\"", fmt.Sprintf("// +kubebuilder:ac:output:package=\"%s\"", newOutputPackage), 1)

	if err := os.WriteFile(filepath.Join(dir, "groupversion_info.go"), []byte(newData), 0644); err != nil {
		return fmt.Errorf("error writing groupversion_info.go: %w", err)
	}

	return nil
}
