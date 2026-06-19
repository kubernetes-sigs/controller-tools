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

package genall

import (
	. "github.com/onsi/ginkgo"
	g "github.com/onsi/gomega"
)

var _ = Describe("yamlMarshal", func() {
	It("should preserve integer types without adding decimal points", func() {
		type schemaLike struct {
			Minimum *float64 `json:"minimum,omitempty"`
			Maximum *float64 `json:"maximum,omitempty"`
			Port    *int32   `json:"port,omitempty"`
			Count   *int64   `json:"count,omitempty"`
		}

		minVal := -0.5
		maxVal := 1.5
		port := int32(8080)
		count := int64(2)

		out, err := yamlMarshal(schemaLike{Minimum: &minVal, Maximum: &maxVal, Port: &port, Count: &count})
		g.Expect(err).NotTo(g.HaveOccurred())

		y := string(out)
		g.Expect(y).To(g.ContainSubstring("minimum: -0.5"))
		g.Expect(y).To(g.ContainSubstring("maximum: 1.5"))
		g.Expect(y).To(g.ContainSubstring("port: 8080"))
		g.Expect(y).To(g.ContainSubstring("count: 2"))
		g.Expect(y).NotTo(g.ContainSubstring("2.0"))
		g.Expect(y).NotTo(g.ContainSubstring("8080.0"))
	})

	It("should marshal nil as null", func() {
		out, err := yamlMarshal(nil)
		g.Expect(err).NotTo(g.HaveOccurred())
		g.Expect(string(out)).To(g.ContainSubstring("null"))
	})

	It("should marshal string values correctly", func() {
		obj := map[string]any{
			"name":    "hello-world",
			"special": "value: with colon",
			"empty":   "",
		}
		out, err := yamlMarshal(obj)
		g.Expect(err).NotTo(g.HaveOccurred())

		y := string(out)
		g.Expect(y).To(g.ContainSubstring("name: hello-world"))
		g.Expect(y).To(g.Or(
			g.ContainSubstring("value: with colon"),
			g.ContainSubstring("'value: with colon'"),
			g.ContainSubstring(`"value: with colon"`),
		))
	})

	It("should marshal arrays correctly", func() {
		obj := map[string]any{
			"items": []any{"a", "b", "c"},
		}
		out, err := yamlMarshal(obj)
		g.Expect(err).NotTo(g.HaveOccurred())

		y := string(out)
		g.Expect(y).To(g.ContainSubstring("- a"))
		g.Expect(y).To(g.ContainSubstring("- b"))
		g.Expect(y).To(g.ContainSubstring("- c"))
	})

	It("should marshal deeply nested maps correctly", func() {
		obj := map[string]any{
			"spec": map[string]any{
				"template": map[string]any{
					"metadata": map[string]any{
						"labels": map[string]any{
							"app": "myapp",
						},
					},
				},
			},
		}
		out, err := yamlMarshal(obj)
		g.Expect(err).NotTo(g.HaveOccurred())
		g.Expect(string(out)).To(g.ContainSubstring("app: myapp"))
	})

	It("should marshal booleans correctly", func() {
		obj := map[string]any{"enabled": true, "disabled": false}
		out, err := yamlMarshal(obj)
		g.Expect(err).NotTo(g.HaveOccurred())

		y := string(out)
		g.Expect(y).To(g.ContainSubstring("enabled: true"))
		g.Expect(y).To(g.ContainSubstring("disabled: false"))
	})

	It("should apply multiple transforms in order", func() {
		obj := map[string]any{
			"metadata": map[string]any{
				"name":              "example",
				"creationTimestamp": "2024-01-01T00:00:00Z",
			},
			"extra": "keep",
		}

		removeExtra := func(o map[string]any) error {
			delete(o, "extra")
			return nil
		}

		out, err := yamlMarshal(obj,
			WithTransform(TransformRemoveCreationTimestamp),
			WithTransform(removeExtra),
		)
		g.Expect(err).NotTo(g.HaveOccurred())

		y := string(out)
		g.Expect(y).NotTo(g.ContainSubstring("creationTimestamp"))
		g.Expect(y).NotTo(g.ContainSubstring("extra"))
		g.Expect(y).To(g.ContainSubstring("name: example"))
	})

	It("should apply transforms and produce valid YAML for a CRD-like object", func() {
		obj := map[string]any{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]any{
				"name":              "foos.example.com",
				"creationTimestamp": "2024-01-01T00:00:00Z",
			},
			"spec": map[string]any{
				"group": "example.com",
				"versions": []any{
					map[string]any{"name": "v1", "served": true, "storage": true},
				},
			},
		}

		out, err := yamlMarshal(obj, WithTransform(TransformRemoveCreationTimestamp))
		g.Expect(err).NotTo(g.HaveOccurred())

		y := string(out)
		g.Expect(y).NotTo(g.ContainSubstring("creationTimestamp"))
		g.Expect(y).To(g.ContainSubstring("name: foos.example.com"))
		g.Expect(y).To(g.ContainSubstring("group: example.com"))
	})
})

var _ = Describe("TransformRemoveCreationTimestamp", func() {
	It("should remove creationTimestamp when present", func() {
		obj := map[string]any{
			"metadata": map[string]any{
				"name":              "test",
				"creationTimestamp": "2024-01-01T00:00:00Z",
			},
		}
		g.Expect(TransformRemoveCreationTimestamp(obj)).To(g.Succeed())

		meta := obj["metadata"].(map[string]any)
		g.Expect(meta).NotTo(g.HaveKey("creationTimestamp"))
		g.Expect(meta).To(g.HaveKey("name"))
	})

	It("should be a no-op when metadata is absent", func() {
		obj := map[string]any{"kind": "CronJob"}
		g.Expect(TransformRemoveCreationTimestamp(obj)).To(g.Succeed())
	})

	It("should be a no-op when creationTimestamp is already absent", func() {
		obj := map[string]any{
			"metadata": map[string]any{"name": "test"},
		}
		g.Expect(TransformRemoveCreationTimestamp(obj)).To(g.Succeed())

		meta := obj["metadata"].(map[string]any)
		g.Expect(meta).To(g.HaveKey("name"))
	})

	It("should be a no-op when metadata is a non-map type", func() {
		obj := map[string]any{"metadata": "not-a-map"}
		g.Expect(TransformRemoveCreationTimestamp(obj)).To(g.Succeed())
	})

	It("should be a no-op when metadata is nil", func() {
		obj := map[string]any{"metadata": nil}
		g.Expect(TransformRemoveCreationTimestamp(obj)).To(g.Succeed())
	})
})
