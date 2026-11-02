/*
Copyright 2026 The Kubernetes Authors.

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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("ApplyConfiguration header year substitution", func() {
	generate := func(headerOptions string) string {
		optionsRegistry := &markers.Registry{}
		Expect(genall.RegisterOptionsMarkers(optionsRegistry)).To(Succeed())
		Expect(optionsRegistry.Register(markers.Must(markers.MakeDefinition("applyconfiguration", markers.DescribesPackage, Generator{})))).To(Succeed())

		genOpt := "applyconfiguration"
		if headerOptions != "" {
			genOpt += ":" + headerOptions
		}
		rt, err := genall.FromOptions(optionsRegistry, []string{
			genOpt,
			"paths=./api/v1",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(rt.Run()).To(BeFalse(), "generator should run without errors")

		data, err := os.ReadFile(filepath.Join("api/v1", applyConfigurationDir, "utils.go"))
		Expect(err).NotTo(HaveOccurred())
		return string(data)
	}

	DescribeTable("writes the resolved header into generated files",
		func(headerContent, headerOptions, expectPresent string, expectAbsent ...string) {
			tmpDir, err := os.MkdirTemp("", "applyconfiguration-year-test")
			Expect(err).NotTo(HaveOccurred())
			defer func() { Expect(os.RemoveAll(tmpDir)).To(Succeed()) }()

			Expect(os.CopyFS(tmpDir, os.DirFS(cronjobDir))).To(Succeed())
			Expect(os.RemoveAll(filepath.Join(tmpDir, "api/v1", applyConfigurationDir))).To(Succeed())

			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chdir(tmpDir)).To(Succeed())
			defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

			if headerContent != "" {
				Expect(os.WriteFile("boilerplate.txt", []byte(headerContent), 0o644)).To(Succeed())
			}

			generated := generate(headerOptions)

			Expect(generated).To(ContainSubstring(expectPresent))
			for _, absent := range expectAbsent {
				Expect(generated).NotTo(ContainSubstring(absent))
			}
		},
		// Pin to a historical year: the underlying generator substitutes the
		// current year for any leftover YEAR, so this value can never match it.
		Entry("substitutes YEAR with the given year",
			"/*\nCopyright YEAR The Kubernetes Authors.\n*/\n",
			"headerFile=boilerplate.txt,year=2000",
			"Copyright 2000 The Kubernetes Authors.",
			"YEAR"),
		Entry("drops the year when none is given",
			"/*\nCopyright YEAR The Kubernetes Authors.\n*/\n",
			"headerFile=boilerplate.txt",
			"Copyright  The Kubernetes Authors.",
			"YEAR"),
		Entry("leaves a header without YEAR untouched",
			"/*\nCopyright 2023 The Kubernetes Authors.\n*/\n",
			"headerFile=boilerplate.txt,year=2000",
			"Copyright 2023 The Kubernetes Authors.",
			"YEAR"),
		Entry("omits the license header when no header file is given",
			"",
			"",
			"DO NOT EDIT.",
			"Copyright"),
	)
})
