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
	"io/ioutil"
	"os"
	"path"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionreg "k8s.io/api/admissionregistration/v1beta1"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/webhook"
)

var _ = Describe("Webhook Generation From Parsing to CustomResourceDefinition", func() {
	It("should properly generate the webhook definition", func() {
		// TODO(directxman12): test generation across multiple versions (right
		// now, we're trusting k/k's conversion code, though, which is probably
		// fine for the time being)
		By("switching into testdata to appease go modules")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata")).To(Succeed()) // go modules are directory-sensitive
		defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

		By("loading the roots")
		pkgs, err := loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg := &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err := ioutil.TempDir("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)
		Expect(webhook.Generator{}.Generate(&genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}))

		By("loading the generated YAML")
		actualFile, err := ioutil.ReadFile(path.Join(outputDir, "manifests.yaml"))
		Expect(err).NotTo(HaveOccurred())
		actualMutating, actualValidating := unmarshalBoth(actualFile)

		By("loading the desired YAML")
		expectedFile, err := ioutil.ReadFile("manifests.yaml")
		Expect(err).NotTo(HaveOccurred())
		expectedMutating, expectedValidating := unmarshalBoth(expectedFile)

		By("comparing the two")
		assertSame := func(actual, expected interface{}) {
			ExpectWithOffset(1, actual).To(Equal(expected), "type not as expected, check pkg/webhook/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(actual, expected))
		}
		assertSame(actualMutating, expectedMutating)
		assertSame(actualValidating, expectedValidating)
	})
})

func unmarshalBoth(in []byte) (mutating admissionreg.MutatingWebhookConfiguration, validating admissionreg.ValidatingWebhookConfiguration) {
	documents := bytes.Split(in, []byte("\n---\n"))[1:]
	ExpectWithOffset(1, documents).To(HaveLen(2), "expected two documents in file, found %d", len(documents))

	ExpectWithOffset(1, yaml.UnmarshalStrict(documents[0], &mutating)).To(Succeed(), "expected the first document in the file to be a mutating webhook configuration")
	ExpectWithOffset(1, yaml.UnmarshalStrict(documents[1], &validating)).To(Succeed(), "expected the second document in the file to be a validating webhook configuration")
	return
}
