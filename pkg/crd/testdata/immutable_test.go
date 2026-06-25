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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Immutable CRD", func() {
	newObj := func(name string) *unstructured.Unstructured {
		return &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "testdata.kubebuilder.io/v1",
				"kind":       "ImmutableType",
				"metadata": map[string]any{
					"name":      name,
					"namespace": "default",
				},
				"spec": map[string]any{
					"requiredString": "original",
					"requiredStruct": map[string]any{
						"key":   "k",
						"value": "v",
					},
					"requiredSlice": []any{"a", "b"},
					"requiredMap":   map[string]any{"k1": "v1"},
					"mutableString": "can-change",
				},
			},
		}
	}

	It("should allow creation", func(ctx SpecContext) {
		obj := newObj("test-create")
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())
	})

	It("should allow update with identical immutable values", func(ctx SpecContext) {
		obj := newObj("test-noop")
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())

		got := obj.DeepCopy()
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(got), got)).To(Succeed())

		Expect(unstructured.SetNestedField(got.Object, "updated", "spec", "mutableString")).To(Succeed())
		Expect(k8sClient.Update(ctx, got)).To(Succeed())
	})

	DescribeTable("should reject changing an immutable required field",
		func(ctx SpecContext, name string, mutate, clear func(*unstructured.Unstructured)) {
			obj := newObj(name)
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			got := obj.DeepCopy()
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(got), got)).To(Succeed())

			mutate(got)
			Expect(k8sClient.Update(ctx, got)).To(MatchError(ContainSubstring("field is immutable")))

			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(got), got)).To(Succeed())
			clear(got)
			Expect(k8sClient.Update(ctx, got)).To(MatchError(ContainSubstring("Required value")))
		},
		Entry("string", "test-req-str",
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedField(obj.Object, "changed", "spec", "requiredString")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				unstructured.RemoveNestedField(obj.Object, "spec", "requiredString")
			},
		),
		Entry("struct", "test-req-struct",
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedField(obj.Object, map[string]any{
					"key":   "k",
					"value": "changed",
				}, "spec", "requiredStruct")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				unstructured.RemoveNestedField(obj.Object, "spec", "requiredStruct")
			},
		),
		Entry("slice", "test-req-slice",
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedSlice(obj.Object, []any{"a", "c"}, "spec", "requiredSlice")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				unstructured.RemoveNestedField(obj.Object, "spec", "requiredSlice")
			},
		),
		Entry("map", "test-req-map",
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedStringMap(obj.Object, map[string]string{"k1": "changed"}, "spec", "requiredMap")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				unstructured.RemoveNestedField(obj.Object, "spec", "requiredMap")
			},
		),
	)

	DescribeTable("should allow setting an optional field for the first time but reject further changes",
		func(ctx SpecContext, name string, setInitial, setChanged, clear func(*unstructured.Unstructured)) {
			obj := newObj(name)
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			got := obj.DeepCopy()
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(got), got)).To(Succeed())

			Expect(unstructured.SetNestedField(got.Object, "changed", "spec", "requiredString")).To(Succeed())
			Expect(k8sClient.Update(ctx, got)).To(MatchError(ContainSubstring("field is immutable")))

			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(got), got)).To(Succeed())

			setInitial(got)
			Expect(k8sClient.Update(ctx, got)).To(Succeed())

			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(got), got)).To(Succeed())
			setChanged(got)
			Expect(k8sClient.Update(ctx, got)).To(MatchError(ContainSubstring("field is immutable")))

			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(got), got)).To(Succeed())
			clear(got)
			Expect(k8sClient.Update(ctx, got)).To(MatchError(ContainSubstring("is immutable once set")))
		},
		Entry("string", "test-opt-str",
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedField(obj.Object, "first-value", "spec", "optionalString")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedField(obj.Object, "second-value", "spec", "optionalString")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				unstructured.RemoveNestedField(obj.Object, "spec", "optionalString")
			},
		),
		Entry("literal string", "test-opt-literal-str",
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedField(obj.Object, "first-value", "spec", "optionalLiteralString")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedField(obj.Object, "second-value", "spec", "optionalLiteralString")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				unstructured.RemoveNestedField(obj.Object, "spec", "optionalLiteralString")
			},
		),
		Entry("struct", "test-opt-struct",
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedField(obj.Object, map[string]any{
					"key":   "k",
					"value": "v",
				}, "spec", "optionalStruct")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedField(obj.Object, map[string]any{
					"key":   "k",
					"value": "changed",
				}, "spec", "optionalStruct")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				unstructured.RemoveNestedField(obj.Object, "spec", "optionalStruct")
			},
		),
		Entry("omitZeroStruct", "test-opt-zero-struct",
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedField(obj.Object, map[string]any{
					"key":   "k",
					"value": "v",
				}, "spec", "optionalOmitZeroStruct")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedField(obj.Object, map[string]any{
					"key":   "k",
					"value": "changed",
				}, "spec", "optionalOmitZeroStruct")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				unstructured.RemoveNestedField(obj.Object, "spec", "optionalOmitZeroStruct")
			},
		),
		Entry("slice", "test-opt-slice",
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedSlice(obj.Object, []any{"x", "y"}, "spec", "optionalSlice")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedSlice(obj.Object, []any{"x", "z"}, "spec", "optionalSlice")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				unstructured.RemoveNestedField(obj.Object, "spec", "optionalSlice")
			},
		),
		Entry("map", "test-opt-map",
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedStringMap(obj.Object, map[string]string{"a": "1"}, "spec", "optionalMap")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				Expect(unstructured.SetNestedStringMap(obj.Object, map[string]string{"a": "2"}, "spec", "optionalMap")).To(Succeed())
			},
			func(obj *unstructured.Unstructured) {
				unstructured.RemoveNestedField(obj.Object, "spec", "optionalMap")
			},
		),
	)
})
