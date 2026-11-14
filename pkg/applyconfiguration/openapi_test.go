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

package applyconfiguration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resolveRefDefinitions", func() {
	const containerSettings = "example.com/v1.ContainerSettings"

	// settingsSchema is the only definition of the chain below that has a schema
	// of its own; everything else is declared in terms of it.
	settingsSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"image": map[string]any{"type": "string"},
		},
	}

	// declaredTypeChain builds `type WorkerSettings ContainerSettings` and
	// `type GPUWorkerSettings WorkerSettings`, where the schema of a declared
	// type is a $ref to the type it is declared from.
	declaredTypeChain := func() map[string]any {
		return map[string]any{
			containerSettings: settingsSchema,
			"example.com/v1.WorkerSettings": map[string]any{
				"$ref": "#/definitions/" + containerSettings,
			},
			"example.com/v1.GPUWorkerSettings": map[string]any{
				"$ref": "#/definitions/example.com/v1.WorkerSettings",
			},
			"example.com/v1.ClusterSpec": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"workers": map[string]any{
						"$ref": "#/definitions/example.com/v1.WorkerSettings",
					},
				},
			},
		}
	}

	It("should resolve chains of declared types in any order", func() {
		// The order in which the definitions are visited decides how far a
		// chain gets resolved in a single pass, so repeat to catch any
		// dependence on it.
		for range 50 {
			definitions := declaredTypeChain()
			resolveRefDefinitions(definitions)

			for name, def := range definitions {
				Expect(def).ToNot(HaveKey("$ref"), "%s was left a reference", name)
			}
			Expect(definitions["example.com/v1.WorkerSettings"]).To(Equal(settingsSchema))
			Expect(definitions["example.com/v1.GPUWorkerSettings"]).To(Equal(settingsSchema))

			// Only the definition that declared the field is a real schema, and
			// its nested reference still points at the named type it referenced.
			spec, ok := definitions["example.com/v1.ClusterSpec"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(spec["properties"]).To(Equal(map[string]any{
				"workers": map[string]any{"$ref": "#/definitions/example.com/v1.WorkerSettings"},
			}))
		}
	})

	It("should keep the extensions declared along the chain", func() {
		definitions := declaredTypeChain()
		definitions["example.com/v1.WorkerSettings"].(map[string]any)["x-kubernetes-map-type"] = "atomic"
		definitions["example.com/v1.GPUWorkerSettings"].(map[string]any)["x-kubernetes-list-type"] = "set"
		resolveRefDefinitions(definitions)

		Expect(definitions["example.com/v1.GPUWorkerSettings"]).To(HaveKeyWithValue("x-kubernetes-list-type", "set"))
		Expect(definitions["example.com/v1.GPUWorkerSettings"]).To(HaveKeyWithValue("x-kubernetes-map-type", "atomic"))
	})

	It("should leave references it cannot resolve", func() {
		// A reference to a definition that does not exist, and a chain that
		// loops, are both broken input: keep them as they are instead of
		// guessing.
		definitions := map[string]any{
			"example.com/v1.Dangling": map[string]any{
				"$ref": "#/definitions/example.com/v1.Missing",
			},
			"example.com/v1.First": map[string]any{
				"$ref": "#/definitions/example.com/v1.Second",
			},
			"example.com/v1.Second": map[string]any{
				"$ref": "#/definitions/example.com/v1.First",
			},
		}

		resolveRefDefinitions(definitions)

		Expect(definitions).To(Equal(map[string]any{
			"example.com/v1.Dangling": map[string]any{
				"$ref": "#/definitions/example.com/v1.Missing",
			},
			"example.com/v1.First": map[string]any{
				"$ref": "#/definitions/example.com/v1.Second",
			},
			"example.com/v1.Second": map[string]any{
				"$ref": "#/definitions/example.com/v1.First",
			},
		}))
	})
})
