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

package applyconfiguration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pkgstest "golang.org/x/tools/go/packages/packagestest"

	"sigs.k8s.io/controller-tools/pkg/applyconfiguration"
	"sigs.k8s.io/controller-tools/pkg/loader"
	testloader "sigs.k8s.io/controller-tools/pkg/loader/testutils"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("ApplyConfiguration markers on var/const/func declarations", func() {
	var col *markers.Collector
	var pkg *loader.Package
	var exported *pkgstest.Exported

	BeforeEach(func() {
		By("setting up a fake package with applyconfiguration markers on var/const/func")
		modules := []pkgstest.Module{
			{
				Name: "sigs.k8s.io/controller-tools/pkg/applyconfiguration/testdata/markerplacement",
				Files: map[string]any{
					"markerplacement.go": `package markerplacement

// ApplyConfiguration generate marker directly above var (no blank line)
// +kubebuilder:ac:generate=true
var DirectVar = "test"

// ApplyConfiguration output package marker with blank line above const
// +kubebuilder:ac:output:package=myapplyconfig

const BlankLineConst = "test"

// ApplyConfiguration generate marker directly above func
// +kubebuilder:ac:generate=true
func DirectFunc() {}
`,
				},
			},
		}

		var pkgs []*loader.Package
		var err error
		pkgs, exported, err = testloader.LoadFakeRoots(pkgstest.Modules, modules, "sigs.k8s.io/controller-tools/pkg/applyconfiguration/testdata/markerplacement")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		pkg = pkgs[0]

		By("setting up the registry and collector with applyconfiguration markers")
		reg := &markers.Registry{}
		gen := applyconfiguration.Generator{}
		Expect(gen.RegisterMarkers(reg)).To(Succeed())
		col = &markers.Collector{Registry: reg}
	})

	AfterEach(func() {
		if exported != nil {
			exported.Cleanup()
		}
	})

	It("should recognise applyconfiguration markers on var/const/func declarations", func() {
		By("collecting package markers")
		pkgMarkers, err := markers.PackageMarkers(col, pkg)
		Expect(err).NotTo(HaveOccurred())
		Expect(pkg.Errors).To(HaveLen(0))

		By("checking that kubebuilder:ac:generate markers were collected")
		acGenMarkers := pkgMarkers["kubebuilder:ac:generate"]
		Expect(acGenMarkers).NotTo(BeNil(), "no kubebuilder:ac:generate markers found")
		Expect(acGenMarkers).To(HaveLen(2), "expected 2 kubebuilder:ac:generate markers")

		By("verifying marker values")
		for _, marker := range acGenMarkers {
			enabled := marker.(bool)
			Expect(enabled).To(BeTrue())
		}

		By("checking that kubebuilder:ac:output:package marker was collected")
		outputPkgMarkers := pkgMarkers["kubebuilder:ac:output:package"]
		Expect(outputPkgMarkers).NotTo(BeNil(), "no kubebuilder:ac:output:package markers found")
		Expect(outputPkgMarkers).To(HaveLen(1), "expected 1 kubebuilder:ac:output:package marker")

		outputPkg := outputPkgMarkers[0].(string)
		Expect(outputPkg).To(Equal("myapplyconfig"))
	})
})
