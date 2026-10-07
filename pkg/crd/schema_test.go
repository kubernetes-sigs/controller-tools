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

package crd

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
	pkgstest "golang.org/x/tools/go/packages/packagestest"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	testloader "sigs.k8s.io/controller-tools/pkg/loader/testutils"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

func transform(t *testing.T, expr string) *apiextensionsv1.JSONSchemaProps {
	// this is *very* hacky but I haven’t found a simple way
	// to get an ast.Expr with all the associated metadata required
	// to run typeToSchema upon it:

	moduleName := "sigs.k8s.io/controller-tools/pkg/crd"
	modules := []pkgstest.Module{
		{
			Name: moduleName,
			Files: map[string]any{
				"test.go": `
				package crd 
				type Test ` + expr,
			},
		},
	}

	pkgs, exported, err := testloader.LoadFakeRoots(pkgstest.Modules, modules, moduleName)
	if exported != nil {
		t.Cleanup(exported.Cleanup)
	}

	if err != nil {
		t.Fatalf("unable to load fake package: %s", err)
	}

	if len(pkgs) != 1 {
		t.Fatal("expected to parse only one package")
	}

	pkg := pkgs[0]
	pkg.NeedTypesInfo()
	failIfErrors(t, pkg.Errors)

	schemaContext := newSchemaContext(pkg, nil, true, false).ForInfo(&markers.TypeInfo{})
	// yick: grab the only type definition
	definedType := pkg.Syntax[0].Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type
	result := typeToSchema(schemaContext, definedType)
	failIfErrors(t, pkg.Errors)
	return result
}

// transformPackage loads pkgContents as a fake package and returns the schema
// produced for the type definition named typeName, together with the package so
// callers can inspect (or assert) any errors raised during schema generation.
// Unlike transform, it does not fail the test when schema generation errors,
// which lets negative cases assert that an error was produced.
func transformPackage(t *testing.T, pkgContents, typeName string) (*apiextensionsv1.JSONSchemaProps, *loader.Package) {
	moduleName := "sigs.k8s.io/controller-tools/pkg/crd"
	modules := []pkgstest.Module{
		{
			Name: moduleName,
			Files: map[string]any{
				"test.go": pkgContents,
			},
		},
	}

	pkgs, exported, err := testloader.LoadFakeRoots(pkgstest.Modules, modules, moduleName)
	if exported != nil {
		t.Cleanup(exported.Cleanup)
	}

	if err != nil {
		t.Fatalf("unable to load fake package: %s", err)
	}

	if len(pkgs) != 1 {
		t.Fatal("expected to parse only one package")
	}

	pkg := pkgs[0]
	pkg.NeedTypesInfo()

	schemaContext := newSchemaContext(pkg, nil, true, false).ForInfo(&markers.TypeInfo{})

	var definedType ast.Expr
	for _, decl := range pkg.Syntax[0].Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if ok && typeSpec.Name.Name == typeName {
				definedType = typeSpec.Type
			}
		}
	}
	if definedType == nil {
		t.Fatalf("could not find type %q in fake package", typeName)
	}

	return typeToSchema(schemaContext, definedType), pkg
}

const textMarshalerKeyPackage = `
	package crd

	// textKey implements encoding.TextMarshaler, so it serializes to a string
	// and is a valid (JSON string) map key.
	type textKey struct{}

	func (textKey) MarshalText() ([]byte, error) { return nil, nil }

	// nonStringKey is a struct that does not implement encoding.TextMarshaler.
	type nonStringKey struct{}

	type TextMarshalerKeyMap map[textKey]string
	type NonStringKeyMap map[nonStringKey]string
`

// Test_Schema_MapOfTextMarshalerKey verifies that a map keyed by a type that
// implements encoding.TextMarshaler is accepted (mirroring how such types are
// accepted as struct fields) and produces an ordinary string-keyed object
// schema with no key schema emitted.
func Test_Schema_MapOfTextMarshalerKey(t *testing.T) {
	g := gomega.NewWithT(t)

	output, pkg := transformPackage(t, textMarshalerKeyPackage, "TextMarshalerKeyMap")
	failIfErrors(t, pkg.Errors)
	g.Expect(output).To(gomega.Equal(&apiextensionsv1.JSONSchemaProps{
		Type: "object",
		AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
			Allows: true,
			Schema: &apiextensionsv1.JSONSchemaProps{
				Type: "string",
			},
		},
	}))
}

// Test_Schema_MapOfNonStringKey verifies that a struct key that does not
// implement encoding.TextMarshaler is still rejected.
func Test_Schema_MapOfNonStringKey(t *testing.T) {
	g := gomega.NewWithT(t)

	_, pkg := transformPackage(t, textMarshalerKeyPackage, "NonStringKeyMap")
	g.Expect(pkg.Errors).ToNot(gomega.BeEmpty())
	g.Expect(pkg.Errors[0].Msg).To(gomega.ContainSubstring("map keys must be strings"))
}

func failIfErrors(t *testing.T, errs []packages.Error) {
	if len(errs) > 0 {
		msgs := make([]string, 0, len(errs))
		for _, e := range errs {
			msgs = append(msgs, e.Msg)
		}

		t.Fatalf("error loading fake package: %s", strings.Join(msgs, "; "))
	}
}

var arrayOfNumbersSchema *apiextensionsv1.JSONSchemaProps = &apiextensionsv1.JSONSchemaProps{
	Type: "array",
	Items: &apiextensionsv1.JSONSchemaPropsOrArray{
		Schema: &apiextensionsv1.JSONSchemaProps{
			Type: "number",
		},
	},
}

func Test_Schema_ArrayOfFloat32(t *testing.T) {
	g := gomega.NewWithT(t)

	output := transform(t, "[]float32")
	g.Expect(output).To(gomega.Equal(arrayOfNumbersSchema))
}

func Test_Schema_MapOfStringToArrayOfFloat32(t *testing.T) {
	g := gomega.NewWithT(t)

	output := transform(t, "map[string][]float32")
	g.Expect(output).To(gomega.Equal(&apiextensionsv1.JSONSchemaProps{
		Type: "object",
		AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
			Allows: true,
			Schema: arrayOfNumbersSchema,
		},
	}))
}

func Test_Schema_ApplyMarkers(t *testing.T) {
	g := gomega.NewWithT(t)

	props := &apiextensionsv1.JSONSchemaProps{}
	ctx := &schemaContext{}

	var invocations []string

	applyMarkers(ctx, markers.MarkerValues{
		"blah": []any{
			&testPriorityMarker{
				priority: 0, callback: func() {
					invocations = append(invocations, "0")
				},
			},
			&testPriorityMarker{priority: 2, callback: func() {
				invocations = append(invocations, "2")
			}},
			&testPriorityMarker{priority: 11, callback: func() {
				invocations = append(invocations, "11")
			}},
			&defaultPriorityMarker{callback: func() {
				invocations = append(invocations, "default")
			}},
			&testapplyFirstMarker{callback: func() {
				invocations = append(invocations, "applyFirst")
			}},
		}}, props, nil)

	g.Expect(invocations).To(gomega.Equal([]string{"0", "applyFirst", "2", "default", "11"}))
}

type defaultPriorityMarker struct {
	callback func()
}

func (m *defaultPriorityMarker) ApplyToSchema(*crdmarkers.SchemaContext, *apiextensionsv1.JSONSchemaProps) error {
	m.callback()
	return nil
}

type testPriorityMarker struct {
	priority crdmarkers.ApplyPriority
	callback func()
}

func (m *testPriorityMarker) ApplyPriority() crdmarkers.ApplyPriority {
	return m.priority
}

func (m *testPriorityMarker) ApplyToSchema(*crdmarkers.SchemaContext, *apiextensionsv1.JSONSchemaProps) error {
	m.callback()
	return nil
}

type testapplyFirstMarker struct {
	callback func()
}

func (m *testapplyFirstMarker) ApplyFirst() {}
func (m *testapplyFirstMarker) ApplyToSchema(*crdmarkers.SchemaContext, *apiextensionsv1.JSONSchemaProps) error {
	m.callback()
	return nil
}

// Test_Schema_TypeAlias_Map verifies that a type alias to a map type
// (type X = map[string]string) produces a $ref to the alias instead of
// erroring out. This is the fix for https://github.com/kubernetes-sigs/controller-tools/issues/1462.
func Test_Schema_TypeAlias_Map(t *testing.T) {
	g := gomega.NewWithT(t)

	pkgContents := `
package crd

// +kubebuilder:validation:XValidation:rule="self.all(key, size(key) > 0)",message="keys must be non-empty"
type MapAlias = map[string]string
`
	pkg := loadTestPackage(t, pkgContents)

	// Find the MapAlias identifier from the alias type declaration's AST.
	// We need an *ast.Ident to trigger localNamedToSchema.
	ident := findTypeIdent(t, pkg, "MapAlias")
	// Use a no-op schemaRequester since we only test the immediate output.
	ctx := newSchemaContext(pkg, &noopSchemaRequester{}, true, false).ForInfo(&markers.TypeInfo{})
	result := localNamedToSchema(ctx, ident)
	failIfErrors(t, pkg.Errors)

	g.Expect(result.Ref).ToNot(gomega.BeNil(), "map type alias should produce a $ref")
	g.Expect(*result.Ref).To(gomega.ContainSubstring("MapAlias"))
}

// Test_Schema_TypeAlias_Slice verifies that a type alias to a slice type
// produces a $ref to the alias instead of erroring out.
func Test_Schema_TypeAlias_Slice(t *testing.T) {
	g := gomega.NewWithT(t)

	pkgContents := `
package crd

type SliceAlias = []string
`
	pkg := loadTestPackage(t, pkgContents)
	ident := findTypeIdent(t, pkg, "SliceAlias")
	ctx := newSchemaContext(pkg, &noopSchemaRequester{}, true, false).ForInfo(&markers.TypeInfo{})
	result := localNamedToSchema(ctx, ident)
	failIfErrors(t, pkg.Errors)

	g.Expect(result.Ref).ToNot(gomega.BeNil(), "slice type alias should produce a $ref")
	g.Expect(*result.Ref).To(gomega.ContainSubstring("SliceAlias"))
}

// Test_Schema_TypeAlias_Basic verifies that a basic type alias
// (type X = string) still produces an inline type with a $ref
// (preserving existing behavior for field-level markers like MinLength).
func Test_Schema_TypeAlias_Basic(t *testing.T) {
	g := gomega.NewWithT(t)

	pkgContents := `
package crd

type StringAlias = string
`
	pkg := loadTestPackage(t, pkgContents)
	ident := findTypeIdent(t, pkg, "StringAlias")
	ctx := newSchemaContext(pkg, &noopSchemaRequester{}, true, false).ForInfo(&markers.TypeInfo{})
	result := localNamedToSchema(ctx, ident)
	failIfErrors(t, pkg.Errors)

	g.Expect(result.Type).To(gomega.Equal("string"), "basic type alias should preserve inline type")
	g.Expect(result.Ref).ToNot(gomega.BeNil(), "basic type alias should also have a $ref")
}

// Test_Schema_TypeAlias_Struct verifies that a struct type alias
// (type X = SomeStruct) still produces a $ref to the underlying named type
// (preserving existing behavior for inline/embedded struct aliases).
func Test_Schema_TypeAlias_Struct(t *testing.T) {
	g := gomega.NewWithT(t)

	pkgContents := `
package crd

type EmbeddedStruct struct {
	Name string ` + "`json:\"name\"`" + `
}

type StructAlias = EmbeddedStruct
`
	pkg := loadTestPackage(t, pkgContents)
	ident := findTypeIdent(t, pkg, "StructAlias")
	ctx := newSchemaContext(pkg, &noopSchemaRequester{}, true, false).ForInfo(&markers.TypeInfo{})
	result := localNamedToSchema(ctx, ident)
	failIfErrors(t, pkg.Errors)

	g.Expect(result.Ref).ToNot(gomega.BeNil(), "struct type alias should produce a $ref")
	// The $ref should point to the underlying named type (EmbeddedStruct),
	// not the alias name, preserving existing behavior for inline/embedded
	// struct aliases and the applyconfiguration generator.
	g.Expect(*result.Ref).To(gomega.ContainSubstring("EmbeddedStruct"))
}

// noopSchemaRequester is a schemaRequester that does nothing, for unit tests
// that only check the immediate schema output without resolving references.
type noopSchemaRequester struct{}

func (noopSchemaRequester) NeedSchemaFor(typ TypeIdent) {}
func (noopSchemaRequester) LookupType(pkg *loader.Package, name string) *markers.TypeInfo {
	return nil
}

// loadTestPackage loads pkgContents as a fake package for unit testing.
func loadTestPackage(t *testing.T, pkgContents string) *loader.Package {
	moduleName := "sigs.k8s.io/controller-tools/pkg/crd"
	modules := []pkgstest.Module{
		{
			Name: moduleName,
			Files: map[string]any{
				"test.go": pkgContents,
			},
		},
	}

	pkgs, exported, err := testloader.LoadFakeRoots(pkgstest.Modules, modules, moduleName)
	if exported != nil {
		t.Cleanup(exported.Cleanup)
	}
	if err != nil {
		t.Fatalf("unable to load fake package: %s", err)
	}
	if len(pkgs) != 1 {
		t.Fatal("expected to parse only one package")
	}

	pkg := pkgs[0]
	pkg.NeedTypesInfo()
	return pkg
}

// findTypeIdent returns an *ast.Ident for the type declaration with the given name.
// This is used to test localNamedToSchema directly.
func findTypeIdent(t *testing.T, pkg *loader.Package, typeName string) *ast.Ident {
	t.Helper()
	for _, decl := range pkg.Syntax[0].Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if ok && typeSpec.Name.Name == typeName {
				return typeSpec.Name
			}
		}
	}
	t.Fatalf("could not find type %q in fake package", typeName)
	return nil
}
