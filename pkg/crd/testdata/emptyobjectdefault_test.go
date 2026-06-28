/*
Copyright The Kubernetes Authors.

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
package cronjob

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var _ = Describe("EmptyObjectDefault CRD", func() {
	newObj := func(name string, policyAuditConfig map[string]any) *unstructured.Unstructured {
		spec := map[string]any{}
		if policyAuditConfig != nil {
			spec["policyAuditConfig"] = policyAuditConfig
		}
		return &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "testdata.kubebuilder.io/v1",
				"kind":       "EmptyObjectDefault",
				"metadata": map[string]any{
					"name":      name,
					"namespace": "default",
				},
				"spec": spec,
			},
		}
	}

	getPolicyAuditConfig := func(ctx SpecContext, obj *unstructured.Unstructured) map[string]any {
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj)).To(Succeed())
		cfg, found, err := unstructured.NestedMap(obj.Object, "spec", "policyAuditConfig")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue(), "policyAuditConfig should be defaulted to an object, not left unset")
		return cfg
	}

	It("should apply the empty object default and fill in all nested defaults when the field is omitted", func(ctx SpecContext) {
		obj := newObj("defaults-omitted", nil)
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())

		Expect(getPolicyAuditConfig(ctx, obj)).To(Equal(map[string]any{
			"rateLimit":      int64(20),
			"maxFileSize":    int64(50),
			"destination":    "null",
			"syslogFacility": "local0",
		}))
	})

	It("should keep explicitly set values and only default the missing nested fields", func(ctx SpecContext) {
		obj := newObj("defaults-partial", map[string]any{
			"rateLimit": int64(5),
		})
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())

		Expect(getPolicyAuditConfig(ctx, obj)).To(Equal(map[string]any{
			"rateLimit":      int64(5),
			"maxFileSize":    int64(50),
			"destination":    "null",
			"syslogFacility": "local0",
		}))
	})

	It("should reject a nested value that violates its validation even though the parent defaults to {}", func(ctx SpecContext) {
		obj := newObj("defaults-invalid", map[string]any{
			"rateLimit": int64(0),
		})
		Expect(k8sClient.Create(ctx, obj)).To(MatchError(ContainSubstring("greater than or equal to 1")))
	})
})

var _ = Describe("RequiredChildDefault CRD", func() {
	It("should be rejected by the API server because the {} default is missing the required child field", func(ctx SpecContext) {
		data, err := os.ReadFile("testdata.kubebuilder.io_requiredchilddefaults.yaml")
		Expect(err).NotTo(HaveOccurred())

		crd := &apiextensionsv1.CustomResourceDefinition{}
		Expect(yaml.UnmarshalStrict(data, crd)).To(Succeed())

		err = k8sClient.Create(ctx, crd)
		Expect(err).To(MatchError(And(
			ContainSubstring("default"),
			ContainSubstring("stage"),
		)))
	})
})
