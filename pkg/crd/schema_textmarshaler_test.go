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

package crd

import (
	"go/ast"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pkgstest "golang.org/x/tools/go/packages/packagestest"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/loader"
	testloader "sigs.k8s.io/controller-tools/pkg/loader/testutils"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("mapToSchema", func() {
	Context("TextMarshaler map key validation", func() {
		// cleanup holds the teardown function returned by the fake package loader.
		// AfterEach calls it so temp directories are removed after every It block.
		var cleanup func()

		AfterEach(func() {
			if cleanup != nil {
				cleanup()
				cleanup = nil
			}
		})

		// schemaForType loads src as a fake package and returns the schema for the
		// named type, together with the package so callers can inspect errors.
		// It sets cleanup so AfterEach can release temp files after the It block ends.
		schemaForType := func(src, typeName string) (*apiextensionsv1.JSONSchemaProps, *loader.Package) {
			const mod = "sigs.k8s.io/controller-tools/pkg/crd"
			pkgs, exported, err := testloader.LoadFakeRoots(pkgstest.Modules, []pkgstest.Module{{
				Name:  mod,
				Files: map[string]any{"test.go": src},
			}}, mod)
			if exported != nil {
				cleanup = exported.Cleanup
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))

			pkg := pkgs[0]
			pkg.NeedTypesInfo()

			schemaCtx := newSchemaContext(pkg, nil, true, false).ForInfo(&markers.TypeInfo{})

			var definedType ast.Expr
			for _, decl := range pkg.Syntax[0].Decls {
				if gd, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range gd.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == typeName {
							definedType = ts.Type
						}
					}
				}
			}
			Expect(definedType).NotTo(BeNil(), "type %q not found in fake package", typeName)
			return typeToSchema(schemaCtx, definedType), pkg
		}

		// src defines three map types covering the three cases below.
		const src = `
package crd

// valueKey implements encoding.TextMarshaler via a value receiver.
// encoding/json can call MarshalText on map keys of this type because
// the method is reachable without a pointer.
type valueKey struct{}
func (valueKey) MarshalText() ([]byte, error) { return nil, nil }

// ptrKey implements encoding.TextMarshaler via a pointer receiver only.
// Map keys are not addressable, so encoding/json cannot call *ptrKey.MarshalText
// and will return an error at runtime — this type must be rejected.
type ptrKey struct{}
func (*ptrKey) MarshalText() ([]byte, error) { return nil, nil }

// noMarshalKey is a plain struct with no TextMarshaler implementation.
type noMarshalKey struct{}

type ValueKeyMap    map[valueKey]string
type PtrKeyMap      map[ptrKey]string
type NoMarshalKeyMap map[noMarshalKey]string
`

		It("accepts a value-receiver TextMarshaler type as a map key", func() {
			output, pkg := schemaForType(src, "ValueKeyMap")
			Expect(pkg.Errors).To(BeEmpty())
			Expect(output).To(Equal(&apiextensionsv1.JSONSchemaProps{
				Type: "object",
				AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
					Allows: true,
					Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"},
				},
			}))
		})

		It("rejects a pointer-receiver-only TextMarshaler type as a map key", func() {
			_, pkg := schemaForType(src, "PtrKeyMap")
			Expect(pkg.Errors).NotTo(BeEmpty())
			Expect(pkg.Errors[0].Msg).To(ContainSubstring("map keys must be strings"))
		})

		It("rejects a struct that does not implement TextMarshaler", func() {
			_, pkg := schemaForType(src, "NoMarshalKeyMap")
			Expect(pkg.Errors).NotTo(BeEmpty())
			Expect(pkg.Errors[0].Msg).To(ContainSubstring("map keys must be strings"))
		})
	})
})
