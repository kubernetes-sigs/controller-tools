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

package markers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pkgstest "golang.org/x/tools/go/packages/packagestest"
	testloader "sigs.k8s.io/controller-tools/pkg/loader/testutils"
	. "sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("Package-level markers on var/const/func", func() {
	It("should re-classify all package-level marker types", func() {
		By("setting up a package with various package-level markers on var/const/func")
		modules := []pkgstest.Module{
			{
				Name: "sigs.k8s.io/controller-tools/pkg/markers/testdata/pkgmarkers",
				Files: map[string]any{
					"pkgmarkers.go": `package pkgmarkers

// CRD markers on var/const/func (no blank line)
// +groupName=test.example.com
var varWithGroupName = "test"

// +versionName=v1beta1
const constWithVersionName = "test"

// +kubebuilder:skip
func funcWithSkip() {}

// +kubebuilder:validation:Optional
var varWithOptional = "test"

// +kubebuilder:validation:Required
const constWithRequired = "test"

// Deepcopy markers
// +kubebuilder:object:generate=true
var varWithObjectGenerate = "test"

// +k8s:deepcopy-gen=package
const constWithDeepCopyGen = "test"

// ApplyConfiguration markers
// +kubebuilder:ac:generate=true
func funcWithACGenerate() {}

// +kubebuilder:ac:output:package=myapplyconfig
var varWithACOutputPackage = "test"

// Webhook markers
// +kubebuilder:webhook:path=/validate,mutating=false,failurePolicy=fail,sideEffects=None,groups=test,resources=foos,verbs=create,versions=v1,name=test.example.com,admissionReviewVersions=v1
var varWithWebhook = "test"

// +kubebuilder:webhookconfiguration:name=my-webhook-config,mutating=false
const constWithWebhookConfig = "test"

// Multiple package markers on var (no blank line)
// +groupName=multi.example.com
// +versionName=v1alpha1
var varWithMultipleMarkers = "test"
`,
				},
			},
		}

		pkgs, exported, err := testloader.LoadFakeRoots(pkgstest.Modules, modules, "sigs.k8s.io/controller-tools/pkg/markers/testdata/pkgmarkers")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		defer exported.Cleanup()

		pkg := pkgs[0]

		By("setting up registry with all package-level markers")
		reg := &Registry{}
		Expect(reg.Define("groupName", DescribesPackage, "")).To(Succeed())
		Expect(reg.Define("versionName", DescribesPackage, "")).To(Succeed())
		Expect(reg.Define("kubebuilder:skip", DescribesPackage, struct{}{})).To(Succeed())
		Expect(reg.Define("kubebuilder:validation:Optional", DescribesPackage, struct{}{})).To(Succeed())
		Expect(reg.Define("kubebuilder:validation:Required", DescribesPackage, struct{}{})).To(Succeed())
		Expect(reg.Define("kubebuilder:object:generate", DescribesPackage, false)).To(Succeed())
		Expect(reg.Define("k8s:deepcopy-gen", DescribesPackage, "")).To(Succeed())
		Expect(reg.Define("kubebuilder:ac:generate", DescribesPackage, false)).To(Succeed())
		Expect(reg.Define("kubebuilder:ac:output:package", DescribesPackage, "")).To(Succeed())
		Expect(reg.Define("kubebuilder:webhook", DescribesPackage, struct {
			Path                    string
			Mutating                bool
			FailurePolicy           string
			SideEffects             string
			Groups                  []string
			Resources               []string
			Verbs                   []string
			Versions                []string
			Name                    string
			AdmissionReviewVersions []string
		}{})).To(Succeed())
		Expect(reg.Define("kubebuilder:webhookconfiguration", DescribesPackage, struct {
			Name     string
			Mutating bool
		}{})).To(Succeed())

		col := &Collector{Registry: reg}

		By("collecting package markers")
		pkgMarkers, err := PackageMarkers(col, pkg)
		Expect(err).NotTo(HaveOccurred())
		Expect(pkg.Errors).To(HaveLen(0))

		By("verifying all CRD package-level markers were collected")
		Expect(pkgMarkers).To(HaveKey("groupName"))
		Expect(pkgMarkers).To(HaveKey("versionName"))
		Expect(pkgMarkers).To(HaveKey("kubebuilder:skip"))
		Expect(pkgMarkers).To(HaveKey("kubebuilder:validation:Optional"))
		Expect(pkgMarkers).To(HaveKey("kubebuilder:validation:Required"))

		By("verifying deepcopy markers were collected")
		Expect(pkgMarkers).To(HaveKey("kubebuilder:object:generate"))
		Expect(pkgMarkers).To(HaveKey("k8s:deepcopy-gen"))

		By("verifying applyconfiguration markers were collected")
		Expect(pkgMarkers).To(HaveKey("kubebuilder:ac:generate"))
		Expect(pkgMarkers).To(HaveKey("kubebuilder:ac:output:package"))

		By("verifying webhook markers were collected")
		Expect(pkgMarkers).To(HaveKey("kubebuilder:webhook"))
		Expect(pkgMarkers).To(HaveKey("kubebuilder:webhookconfiguration"))

		By("verifying marker values are correct")
		Expect(pkgMarkers["groupName"]).To(ContainElement("test.example.com"))
		Expect(pkgMarkers["versionName"]).To(ContainElement("v1beta1"))
		Expect(pkgMarkers["kubebuilder:ac:output:package"]).To(ContainElement("myapplyconfig"))

		By("verifying multiple markers on same var")
		Expect(pkgMarkers["groupName"]).To(ContainElement("multi.example.com"))
		Expect(pkgMarkers["versionName"]).To(ContainElement("v1alpha1"))
	})

	It("should work with blank lines (backward compatibility)", func() {
		By("setting up a package with blank lines before declarations")
		modules := []pkgstest.Module{
			{
				Name: "sigs.k8s.io/controller-tools/pkg/markers/testdata/pkgmarkersblank",
				Files: map[string]any{
					"pkgmarkers.go": `package pkgmarkersblank

// groupName marker with blank line
// +groupName=blank.example.com

var varWithBlankLine = "test"
`,
				},
			},
		}

		pkgs, exported, err := testloader.LoadFakeRoots(pkgstest.Modules, modules, "sigs.k8s.io/controller-tools/pkg/markers/testdata/pkgmarkersblank")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		defer exported.Cleanup()

		pkg := pkgs[0]

		reg := &Registry{}
		Expect(reg.Define("groupName", DescribesPackage, "")).To(Succeed())
		col := &Collector{Registry: reg}

		By("collecting package markers")
		pkgMarkers, err := PackageMarkers(col, pkg)
		Expect(err).NotTo(HaveOccurred())

		By("verifying blank line style still works")
		Expect(pkgMarkers["groupName"]).To(ContainElement("blank.example.com"))
	})
})
