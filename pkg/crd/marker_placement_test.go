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

package crd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pkgstest "golang.org/x/tools/go/packages/packagestest"

	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	testloader "sigs.k8s.io/controller-tools/pkg/loader/testutils"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("CRD package markers on var/const/func declarations", func() {
	var col *markers.Collector
	var pkg *loader.Package
	var exported *pkgstest.Exported

	BeforeEach(func() {
		By("setting up a fake package with CRD package markers on var/const/func")
		modules := []pkgstest.Module{
			{
				Name: "sigs.k8s.io/controller-tools/pkg/crd/testdata/markerplacement",
				Files: map[string]any{
					"markerplacement.go": `package markerplacement

// CRD groupName marker directly above var (no blank line)
// +groupName=test.example.com
var DirectVar = "test"

// CRD versionName marker with blank line above const
// +versionName=v1beta1

const BlankLineConst = "test"

// CRD skip marker directly above func
// +kubebuilder:skip
func DirectFunc() {}

// CRD validation marker on var
// +kubebuilder:validation:Optional
var OptionalVar = "test"

// CRD validation marker on const
// +kubebuilder:validation:Required
const RequiredConst = "test"
`,
				},
			},
		}

		var pkgs []*loader.Package
		var err error
		pkgs, exported, err = testloader.LoadFakeRoots(pkgstest.Modules, modules, "sigs.k8s.io/controller-tools/pkg/crd/testdata/markerplacement")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		pkg = pkgs[0]

		By("setting up the registry and collector with CRD markers")
		reg := &markers.Registry{}
		Expect(crdmarkers.Register(reg)).To(Succeed())
		col = &markers.Collector{Registry: reg}
	})

	AfterEach(func() {
		if exported != nil {
			exported.Cleanup()
		}
	})

	It("should recognise CRD package markers on var/const/func declarations", func() {
		By("collecting package markers")
		pkgMarkers, err := markers.PackageMarkers(col, pkg)
		Expect(err).NotTo(HaveOccurred())
		Expect(pkg.Errors).To(HaveLen(0))

		By("checking that groupName marker was collected")
		groupNameMarkers := pkgMarkers["groupName"]
		Expect(groupNameMarkers).NotTo(BeNil(), "no groupName markers found")
		Expect(groupNameMarkers).To(HaveLen(1), "expected 1 groupName marker")
		Expect(groupNameMarkers[0].(string)).To(Equal("test.example.com"))

		By("checking that versionName marker was collected")
		versionNameMarkers := pkgMarkers["versionName"]
		Expect(versionNameMarkers).NotTo(BeNil(), "no versionName markers found")
		Expect(versionNameMarkers).To(HaveLen(1), "expected 1 versionName marker")
		Expect(versionNameMarkers[0].(string)).To(Equal("v1beta1"))

		By("checking that kubebuilder:skip marker was collected")
		skipMarkers := pkgMarkers["kubebuilder:skip"]
		Expect(skipMarkers).NotTo(BeNil(), "no kubebuilder:skip markers found")
		Expect(skipMarkers).To(HaveLen(1), "expected 1 kubebuilder:skip marker")

		By("checking that validation markers were collected")
		optionalMarkers := pkgMarkers["kubebuilder:validation:Optional"]
		Expect(optionalMarkers).NotTo(BeNil(), "no validation:Optional markers found")
		Expect(optionalMarkers).To(HaveLen(1), "expected 1 validation:Optional marker")

		requiredMarkers := pkgMarkers["kubebuilder:validation:Required"]
		Expect(requiredMarkers).NotTo(BeNil(), "no validation:Required markers found")
		Expect(requiredMarkers).To(HaveLen(1), "expected 1 validation:Required marker")
	})
})
