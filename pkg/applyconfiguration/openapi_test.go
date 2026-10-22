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

package applyconfiguration

import (
	"encoding/json"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("OpenAPI schema generation", func() {
	var originalCWD string

	BeforeEach(func() {
		tmpDir, err := os.MkdirTemp("", "applyconfiguration-openapi-test")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, tmpDir)
		Expect(os.CopyFS(tmpDir, os.DirFS(cronjobDir))).To(Succeed())

		originalCWD, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		// Cleanups run last-in-first-out, so the working directory is restored before the temporary directory is removed.
		DeferCleanup(os.Chdir, originalCWD)
		Expect(os.Chdir(tmpDir)).To(Succeed()) // go modules are directory-sensitive
	})

	It("should qualify references of types embedded from another package", func() {
		optionsRegistry := &markers.Registry{}
		Expect(genall.RegisterOptionsMarkers(optionsRegistry)).To(Succeed())
		Expect(optionsRegistry.Register(markers.Must(markers.MakeDefinition("applyconfiguration", markers.DescribesPackage, Generator{})))).To(Succeed())
		Expect((Generator{}).RegisterMarkers(optionsRegistry)).To(Succeed())

		roots, err := loader.LoadRoots("./inlineref")
		Expect(err).NotTo(HaveOccurred())
		Expect(roots).To(HaveLen(1))

		ctx := &ObjectGenCtx{
			Collector: &markers.Collector{Registry: optionsRegistry},
			Checker:   &loader.TypeChecker{},
		}

		schemaFile, err := ctx.buildOpenAPISchema(roots[0], schema.GroupVersion{Group: "testdata.kubebuilder.io", Version: "v1"})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.Remove, schemaFile)

		data, err := os.ReadFile(schemaFile)
		Expect(err).NotTo(HaveOccurred())
		var swagger struct {
			Definitions map[string]any `json:"definitions"`
		}
		Expect(json.Unmarshal(data, &swagger)).To(Succeed())

		By("Checking that every reference points to a known definition")
		refs := collectRefs(swagger.Definitions, nil)
		Expect(refs).NotTo(BeEmpty())
		for _, ref := range refs {
			Expect(swagger.Definitions).To(HaveKey(strings.TrimPrefix(ref, "#/definitions/")), "reference %q should resolve", ref)
		}
	})
})

// collectRefs returns all $ref values found in v.
func collectRefs(v any, refs []string) []string {
	switch v := v.(type) {
	case map[string]any:
		for k, val := range v {
			if s, ok := val.(string); ok && k == "$ref" {
				refs = append(refs, s)
				continue
			}
			refs = collectRefs(val, refs)
		}
	case []any:
		for _, val := range v {
			refs = collectRefs(val, refs)
		}
	}
	return refs
}
