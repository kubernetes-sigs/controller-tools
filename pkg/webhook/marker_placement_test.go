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

package webhook_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pkgstest "golang.org/x/tools/go/packages/packagestest"

	"sigs.k8s.io/controller-tools/pkg/loader"
	testloader "sigs.k8s.io/controller-tools/pkg/loader/testutils"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/webhook"
)

var _ = Describe("Webhook markers on var/const/func declarations", func() {
	var col *markers.Collector
	var pkg *loader.Package
	var exported *pkgstest.Exported

	BeforeEach(func() {
		By("setting up a fake package with webhook markers on var/const/func")
		modules := []pkgstest.Module{
			{
				Name: "sigs.k8s.io/controller-tools/pkg/webhook/testdata/markerplacement",
				Files: map[string]any{
					"markerplacement.go": `package markerplacement

// Webhook marker directly above var (no blank line)
// +kubebuilder:webhook:path=/validate-v1,mutating=false,failurePolicy=fail,sideEffects=None,groups=test.example.com,resources=foos,verbs=create;update,versions=v1,name=validate.foo.test.example.com,admissionReviewVersions=v1
var DirectVar = "test"

// Webhook marker with blank line above const
// +kubebuilder:webhook:path=/mutate-v1,mutating=true,failurePolicy=fail,sideEffects=None,groups=test.example.com,resources=bars,verbs=create,versions=v1,name=mutate.bar.test.example.com,admissionReviewVersions=v1

const BlankLineConst = "test"

// WebhookConfiguration marker directly above func
// +kubebuilder:webhookconfiguration:mutating=true,name=my-mutating-webhook-configuration
func DirectFunc() {}
`,
				},
			},
		}

		var pkgs []*loader.Package
		var err error
		pkgs, exported, err = testloader.LoadFakeRoots(pkgstest.Modules, modules, "sigs.k8s.io/controller-tools/pkg/webhook/testdata/markerplacement")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		pkg = pkgs[0]

		By("setting up the registry and collector with webhook markers")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())
		Expect(reg.Register(webhook.WebhookConfigDefinition)).To(Succeed())
		col = &markers.Collector{Registry: reg}
	})

	AfterEach(func() {
		if exported != nil {
			exported.Cleanup()
		}
	})

	It("should recognise webhook markers on var/const/func declarations", func() {
		By("collecting package markers")
		pkgMarkers, err := markers.PackageMarkers(col, pkg)
		Expect(err).NotTo(HaveOccurred())
		Expect(pkg.Errors).To(HaveLen(0))

		By("checking that webhook markers were collected at package level")
		webhookMarkers := pkgMarkers["kubebuilder:webhook"]
		Expect(webhookMarkers).NotTo(BeNil(), "no webhook markers found")
		Expect(webhookMarkers).To(HaveLen(2), "expected 2 webhook markers")

		By("verifying webhook paths are correct")
		paths := make([]string, 0, len(webhookMarkers))
		for _, marker := range webhookMarkers {
			config := marker.(webhook.Config)
			paths = append(paths, config.Path)
		}
		Expect(paths).To(ContainElement("/validate-v1"), "validate webhook marker should be collected")
		Expect(paths).To(ContainElement("/mutate-v1"), "mutate webhook marker should be collected")

		By("checking that webhookconfiguration marker was collected")
		webhookConfigMarkers := pkgMarkers["kubebuilder:webhookconfiguration"]
		Expect(webhookConfigMarkers).NotTo(BeNil(), "no webhookconfiguration markers found")
		Expect(webhookConfigMarkers).To(HaveLen(1), "expected 1 webhookconfiguration marker")

		webhookConfig := webhookConfigMarkers[0].(webhook.WebhookConfig)
		Expect(webhookConfig.Name).To(Equal("my-mutating-webhook-configuration"))
		Expect(webhookConfig.Mutating).To(BeTrue())
	})
})
