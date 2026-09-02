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

package rbac_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pkgstest "golang.org/x/tools/go/packages/packagestest"

	"sigs.k8s.io/controller-tools/pkg/loader"
	testloader "sigs.k8s.io/controller-tools/pkg/loader/testutils"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/rbac"
)

var _ = Describe("RBAC markers on var/const/func declarations", func() {
	var col *markers.Collector
	var pkg *loader.Package
	var exported *pkgstest.Exported

	BeforeEach(func() {
		By("setting up a fake package with RBAC markers on var/const/func")
		modules := []pkgstest.Module{
			{
				Name: "sigs.k8s.io/controller-tools/pkg/rbac/testdata/markerplacement",
				Files: map[string]any{
					"markerplacement.go": `package markerplacement

// RBAC marker directly above var (no blank line)
// +kubebuilder:rbac:groups=test,resources=directvar,verbs=get
var DirectVar = "test"

// RBAC marker with blank line above var
// +kubebuilder:rbac:groups=test,resources=blanklinevar,verbs=get

var BlankLineVar = "test"

// RBAC marker directly above const
// +kubebuilder:rbac:groups=test,resources=directconst,verbs=get
const DirectConst = "test"

// RBAC marker directly above func
// +kubebuilder:rbac:groups=test,resources=directfunc,verbs=get
func DirectFunc() {}
`,
				},
			},
		}

		var pkgs []*loader.Package
		var err error
		pkgs, exported, err = testloader.LoadFakeRoots(pkgstest.Modules, modules, "sigs.k8s.io/controller-tools/pkg/rbac/testdata/markerplacement")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		pkg = pkgs[0]

		By("setting up the registry and collector with RBAC marker")
		reg := &markers.Registry{}
		Expect(reg.Register(rbac.RuleDefinition)).To(Succeed())
		col = &markers.Collector{Registry: reg}
	})

	AfterEach(func() {
		if exported != nil {
			exported.Cleanup()
		}
	})

	It("should recognise RBAC marker directly above var declaration", func() {
		By("collecting package markers")
		pkgMarkers, err := markers.PackageMarkers(col, pkg)
		Expect(err).NotTo(HaveOccurred())
		Expect(pkg.Errors).To(HaveLen(0))

		By("checking that all RBAC markers were collected at package level")
		rbacMarkers := pkgMarkers["kubebuilder:rbac"]
		Expect(rbacMarkers).NotTo(BeNil(), "no RBAC markers found")
		Expect(rbacMarkers).To(HaveLen(4), "expected 4 RBAC markers")

		By("verifying each marker was collected")
		resources := make([]string, 0)
		for _, marker := range rbacMarkers {
			rule := marker.(rbac.Rule)
			if len(rule.Resources) > 0 {
				resources = append(resources, rule.Resources[0])
			}
		}

		Expect(resources).To(ContainElement("directvar"), "directvar marker should be collected")
		Expect(resources).To(ContainElement("blanklinevar"), "blanklinevar marker should be collected")
		Expect(resources).To(ContainElement("directconst"), "directconst marker should be collected")
		Expect(resources).To(ContainElement("directfunc"), "directfunc marker should be collected")
	})
})
