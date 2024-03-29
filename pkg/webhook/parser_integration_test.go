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

package webhook_test

import (
	"bytes"
	"os"
	"path"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/webhook"
)

// TODO(directxman12): test generation across multiple versions (right
// now, we're trusting k/k's conversion code, though, which is probably
// fine for the time being)
var _ = Describe("Webhook Generation From Parsing to CustomResourceDefinition", func() {
	assertSame := func(actual, expected interface{}) {
		ExpectWithOffset(1, actual).To(Equal(expected), "type not as expected, check pkg/webhook/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(actual, expected))
	}

	It("should fail generating a v1beta1 webhook", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/invalid-v1beta1NotSupported")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		genCtx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		err = webhook.Generator{}.Generate(genCtx)
		Expect(err).To(MatchError("unsupported webhook version: v1beta1"))
	})

	It("should fail without admissionReviewVersions specified", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/invalid-admissionReviewVersionsRequired")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		genCtx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		Expect(webhook.Generator{}.Generate(genCtx)).To(Succeed())
		Expect(genCtx.Roots).To(HaveLen(1))
		Expect(genCtx.Roots[0].Errors).To(HaveLen(1))
		Expect(genCtx.Roots[0].Errors[0].Error()).To(ContainSubstring(`missing argument "admissionReviewVersions"`))
	})

	It("should fail with invalid side effects", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/invalid-sideEffects")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		genCtx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		err = webhook.Generator{}.Generate(genCtx)
		Expect(err).To(MatchError("SideEffects should not be set to `Some` or `Unknown` for v1 {Mutating,Validating}WebhookConfiguration"))
	})

	It("should fail with invalid timeout seconds", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/invalid-timeoutSeconds")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		genCtx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		err = webhook.Generator{}.Generate(genCtx)
		Expect(err).To(MatchError("TimeoutSeconds must be between 1 and 30 seconds"))
	})

	It("should properly generate the webhook definition", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/valid")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		genCtx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		Expect(webhook.Generator{}.Generate(genCtx)).To(Succeed())
		for _, r := range genCtx.Roots {
			Expect(r.Errors).To(HaveLen(0))
		}

		By("loading the generated v1 YAML")
		actualFile, err := os.ReadFile(path.Join(outputDir, "manifests.yaml"))
		Expect(err).NotTo(HaveOccurred())
		actualMutating, actualValidating := unmarshalBothV1(actualFile)

		By("loading the desired v1 YAML")
		expectedFile, err := os.ReadFile("manifests.yaml")
		Expect(err).NotTo(HaveOccurred())
		expectedMutating, expectedValidating := unmarshalBothV1(expectedFile)

		By("comparing the two")
		assertSame(actualMutating, expectedMutating)
		assertSame(actualValidating, expectedValidating)
	})

	It("should generate the ordered webhook definitions", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/valid-ordered")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)

		for i := 0; i < 10; i++ {
			genCtx := &genall.GenerationContext{
				Collector:  &markers.Collector{Registry: reg},
				Roots:      pkgs,
				OutputRule: genall.OutputToDirectory(outputDir),
			}
			Expect(webhook.Generator{}.Generate(genCtx)).To(Succeed())
			for _, r := range genCtx.Roots {
				Expect(r.Errors).To(HaveLen(0))
			}

			By("loading the generated v1 YAML")
			actualFile, err := os.ReadFile(path.Join(outputDir, "manifests.yaml"))
			Expect(err).NotTo(HaveOccurred())
			actualManifest := &admissionregv1.ValidatingWebhookConfiguration{}
			Expect(yaml.UnmarshalStrict(actualFile, actualManifest)).To(Succeed())

			By("loading the desired v1 YAML")
			expectedFile, err := os.ReadFile("manifests.yaml")
			Expect(err).NotTo(HaveOccurred())
			expectedManifest := &admissionregv1.ValidatingWebhookConfiguration{}
			Expect(yaml.UnmarshalStrict(expectedFile, expectedManifest)).To(Succeed())

			By("comparing the manifest")
			assertSame(actualManifest, expectedManifest)
		}
	})

	It("should properly generate the webhook definition with url instead of service", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/valid-url")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		genCtx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		Expect(webhook.Generator{}.Generate(genCtx)).To(Succeed())
		for _, r := range genCtx.Roots {
			Expect(r.Errors).To(HaveLen(0))
		}

		By("loading the generated v1 YAML")
		actualFile, err := os.ReadFile(path.Join(outputDir, "manifests.yaml"))
		Expect(err).NotTo(HaveOccurred())
		actualMutating, actualValidating := unmarshalBothV1(actualFile)

		By("loading the desired v1 YAML")
		expectedFile, err := os.ReadFile("manifests.yaml")
		Expect(err).NotTo(HaveOccurred())
		expectedMutating, expectedValidating := unmarshalBothV1(expectedFile)

		By("comparing the two")
		assertSame(actualMutating, expectedMutating)
		assertSame(actualValidating, expectedValidating)
	})

	It("should fail to generate when both path and url are set", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/invalid-path-and-url")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		genCtx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		err = webhook.Generator{}.Generate(genCtx)
		Expect(err).To(HaveOccurred())
	})

	It("should properly generate the webhook definition for single file one", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/valid-single")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the golang file")
		pkgs, err := loader.LoadRoots("webhook_one.go")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		genCtx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		Expect(webhook.Generator{}.Generate(genCtx)).To(Succeed())
		for _, r := range genCtx.Roots {
			Expect(r.Errors).To(HaveLen(0))
		}

		By("loading the generated v1 YAML")
		actualFile, err := os.ReadFile(path.Join(outputDir, "manifests.yaml"))
		Expect(err).NotTo(HaveOccurred())
		actualMutating, actualValidating := unmarshalBothV1(actualFile)

		By("loading the desired v1 YAML")
		expectedFile, err := os.ReadFile("manifests_one.yaml")
		Expect(err).NotTo(HaveOccurred())
		expectedMutating, expectedValidating := unmarshalBothV1(expectedFile)

		By("comparing the two")
		assertSame(actualMutating, expectedMutating)
		assertSame(actualValidating, expectedValidating)
	})

	It("should properly generate the webhook definition for single file two", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/valid-single")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the golang file")
		pkgs, err := loader.LoadRoots("webhook_two.go")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		genCtx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		Expect(webhook.Generator{}.Generate(genCtx)).To(Succeed())
		for _, r := range genCtx.Roots {
			Expect(r.Errors).To(HaveLen(0))
		}

		By("loading the generated v1 YAML")
		actualFile, err := os.ReadFile(path.Join(outputDir, "manifests.yaml"))
		Expect(err).NotTo(HaveOccurred())
		actualMutating, actualValidating := unmarshalBothV1(actualFile)

		By("loading the desired v1 YAML")
		expectedFile, err := os.ReadFile("manifests_two.yaml")
		Expect(err).NotTo(HaveOccurred())
		expectedMutating, expectedValidating := unmarshalBothV1(expectedFile)

		By("comparing the two")
		assertSame(actualMutating, expectedMutating)
		assertSame(actualValidating, expectedValidating)
	})

	It("should properly generate the webhook definition for multiple files", func() {
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata/valid-single")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := os.MkdirTemp("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		genCtx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		Expect(webhook.Generator{}.Generate(genCtx)).To(Succeed())
		for _, r := range genCtx.Roots {
			Expect(r.Errors).To(HaveLen(0))
		}

		By("loading the generated v1 YAML")
		actualFile, err := os.ReadFile(path.Join(outputDir, "manifests.yaml"))
		Expect(err).NotTo(HaveOccurred())
		actualMutating, actualValidating := unmarshalBothV1(actualFile)

		By("loading the desired v1 YAML")
		expectedFile, err := os.ReadFile("manifests_all.yaml")
		Expect(err).NotTo(HaveOccurred())
		expectedMutating, expectedValidating := unmarshalBothV1(expectedFile)

		By("comparing the two")
		assertSame(actualMutating, expectedMutating)
		assertSame(actualValidating, expectedValidating)
	})
})

func unmarshalBothV1(in []byte) (mutating admissionregv1.MutatingWebhookConfiguration, validating admissionregv1.ValidatingWebhookConfiguration) {
	documents := bytes.Split(in, []byte("\n---\n"))
	ExpectWithOffset(1, documents).To(HaveLen(2), "expected two documents in file, found %d", len(documents))

	ExpectWithOffset(1, yaml.UnmarshalStrict(documents[0], &mutating)).To(Succeed(), "expected the first document in the file to be a mutating webhook configuration")
	ExpectWithOffset(1, yaml.UnmarshalStrict(documents[1], &validating)).To(Succeed(), "expected the second document in the file to be a validating webhook configuration")
	return
}
