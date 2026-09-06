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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("DeclarativeValidation CRD", func() {
	newObj := func(name string, items []any, value int64) *unstructured.Unstructured {
		return &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "testdata.kubebuilder.io/v1",
			"kind":       "DeclarativeValidation",
			"metadata": map[string]any{
				"name":      name,
				"namespace": "default",
			},
			"spec": map[string]any{
				"items": items,
				"value": value,
			},
		}}
	}

	It("should accept values within the declarative validation bounds", func(ctx SpecContext) {
		Expect(k8sClient.Create(ctx, newObj("valid-bounds", []any{"one", "two"}, 2))).To(Succeed())
	})

	It("should enforce k8s:maxItems", func(ctx SpecContext) {
		err := k8sClient.Create(ctx, newObj("too-many-items", []any{"one", "two", "three"}, 0))
		Expect(err).To(MatchError(ContainSubstring("must have at most 2 items")))
	})

	DescribeTable("should enforce numeric declarative validation bounds",
		func(ctx SpecContext, name string, value int64, message string) {
			err := k8sClient.Create(ctx, newObj(name, []any{}, value))
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("minimum", "below-minimum", int64(-3), "greater than or equal to -2"),
		Entry("maximum", "above-maximum", int64(3), "less than or equal to 2"),
	)
})
