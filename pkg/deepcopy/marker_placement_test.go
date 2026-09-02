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

package deepcopy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pkgstest "golang.org/x/tools/go/packages/packagestest"

	"sigs.k8s.io/controller-tools/pkg/deepcopy"
	"sigs.k8s.io/controller-tools/pkg/loader"
	testloader "sigs.k8s.io/controller-tools/pkg/loader/testutils"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("Deepcopy markers on var/const/func declarations", func() {
	var col *markers.Collector
	var pkg *loader.Package
	var exported *pkgstest.Exported

	BeforeEach(func() {
		By("setting up a fake package with deepcopy markers on var/const/func")
		modules := []pkgstest.Module{
			{
				Name: "sigs.k8s.io/controller-tools/pkg/deepcopy/testdata/markerplacement",
				Files: map[string]any{
					"markerplacement.go": `package markerplacement

// Deepcopy marker directly above var (no blank line)
// +kubebuilder:object:generate=true
var DirectVar = "test"

// Deepcopy marker with blank line above const
// +kubebuilder:object:generate=true

const BlankLineConst = "test"

// Legacy deepcopy marker directly above func
// +k8s:deepcopy-gen=package
func DirectFunc() {}
`,
				},
			},
		}

		var pkgs []*loader.Package
		var err error
		pkgs, exported, err = testloader.LoadFakeRoots(pkgstest.Modules, modules, "sigs.k8s.io/controller-tools/pkg/deepcopy/testdata/markerplacement")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		pkg = pkgs[0]

		By("setting up the registry and collector with deepcopy markers")
		reg := &markers.Registry{}
		gen := deepcopy.Generator{}
		Expect(gen.RegisterMarkers(reg)).To(Succeed())
		col = &markers.Collector{Registry: reg}
	})

	AfterEach(func() {
		if exported != nil {
			exported.Cleanup()
		}
	})

	It("should recognise deepcopy markers on var/const/func declarations", func() {
		By("collecting package markers")
		pkgMarkers, err := markers.PackageMarkers(col, pkg)
		Expect(err).NotTo(HaveOccurred())
		Expect(pkg.Errors).To(HaveLen(0))

		By("checking that kubebuilder:object:generate markers were collected")
		objectGenMarkers := pkgMarkers["kubebuilder:object:generate"]
		Expect(objectGenMarkers).NotTo(BeNil(), "no kubebuilder:object:generate markers found")
		Expect(objectGenMarkers).To(HaveLen(2), "expected 2 kubebuilder:object:generate markers")

		By("verifying marker values")
		for _, marker := range objectGenMarkers {
			enabled := marker.(bool)
			Expect(enabled).To(BeTrue())
		}

		By("checking that legacy k8s:deepcopy-gen marker was collected")
		legacyMarkers := pkgMarkers["k8s:deepcopy-gen"]
		Expect(legacyMarkers).NotTo(BeNil(), "no k8s:deepcopy-gen markers found")
		Expect(legacyMarkers).To(HaveLen(1), "expected 1 k8s:deepcopy-gen marker")
	})
})
