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
	"fmt"
	gotypes "go/types"
	"reflect"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clientgenutil "k8s.io/code-generator/cmd/client-gen/generators/util"
	"k8s.io/gengo/v2/types"

	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// checkUnsupportedTypes reports the types of pkg for which applyconfiguration-gen
// would generate a broken apply configuration. These are bugs in code-generator,
// so we fail the generation with a clear error instead of writing the broken code.
//
// Each check can be dropped once controller-tools uses a code-generator release
// with the fix for the linked issue. Its testdata case then has to generate
// without errors.
func (ctx *ObjectGenCtx) checkUnsupportedTypes(root *loader.Package, pkg *types.Package) error {
	typeMarkers := make(map[string]markers.MarkerValues)
	if err := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
		typeMarkers[info.Name] = info.Markers
	}); err != nil {
		return err
	}

	var errs []error
	for _, t := range generatedApplyConfigTypes(pkg) {
		errs = append(errs, checkUnexportedEmbeds(t)...)
		errs = append(errs, checkCustomJSON(t, typeMarkers[t.Name.Name])...)
		errs = append(errs, checkShadowedFields(t)...)
	}
	return kerrors.NewAggregate(errs)
}

// generatedApplyConfigTypes returns the types of pkg that get their own apply
// configuration, sorted by name. It follows the rules of applyconfiguration-gen:
// the struct types with JSON-tagged fields that are reachable from a +genclient
// type with the apply verb. Types of other packages, including the ones in
// ExternalApplyConfigurations, never get one, so they are not checked.
func generatedApplyConfigTypes(pkg *types.Package) []*types.Type {
	reachable := map[types.Name]bool{}
	for _, t := range pkg.Types {
		tags, err := clientgenTags(t)
		if err != nil {
			// The generator reports the invalid tags itself.
			continue
		}
		if tags.GenerateClient && (tags.HasVerb("apply") || tags.HasVerb("applyStatus")) {
			findReachable(t, reachable)
		}
	}

	var generated []*types.Type
	for _, t := range pkg.Types {
		if !reachable[t.Name] {
			continue
		}
		if tags, err := clientgenTags(t); err == nil && tags.GenerateClient && !hasEmbeddedObjectMeta(t) {
			// applyconfiguration-gen skips client types without ObjectMeta.
			continue
		}
		if st := structOf(t); st != nil && hasJSONTaggedMember(st) {
			generated = append(generated, t)
		}
	}
	slices.SortFunc(generated, func(a, b *types.Type) int {
		return strings.Compare(a.Name.Name, b.Name.Name)
	})
	return generated
}

// checkUnexportedEmbeds reports the embedded fields of an unexported type. The
// generated apply configuration refers to an unexported name for them, which is
// not defined, so the generated package does not compile.
//
// TODO: drop this check once controller-tools uses a code-generator with the fix
// for https://github.com/kubernetes/kubernetes/issues/143414 (controller-tools#1497).
func checkUnexportedEmbeds(t *types.Type) []error {
	var errs []error
	for _, m := range structOf(t).Members {
		if !m.Embedded || !hasJSONTag(m) {
			continue
		}
		embedded := deref(m.Type)
		if embedded.Name.Package == "" || isExported(embedded.Name.Name) {
			continue
		}
		errs = append(errs, fmt.Errorf("type %s embeds the unexported type %s, and applyconfiguration-gen would generate a reference to an undefined type (https://github.com/kubernetes/kubernetes/issues/143414); export the embedded type to work around it",
			t.Name, embedded.Name.Name))
	}
	return errs
}

// checkCustomJSON reports a struct type with a custom MarshalJSON or UnmarshalJSON
// method whose schema is not an object. Its generated apply configuration is
// serialized field by field as an object, so the apply request does not match the
// schema. A custom method that keeps the object form, like the default one,
// works, so a type is only reported when +kubebuilder:validation:Type says that
// its JSON form is something else.
//
// TODO: drop this check once controller-tools uses a code-generator with the fix
// for https://github.com/kubernetes/kubernetes/issues/143415 (controller-tools#1498).
func checkCustomJSON(t *types.Type, typeMarkers markers.MarkerValues) []error {
	method := customJSONMethod(t)
	if method == "" {
		return nil
	}
	schemaType, ok := typeMarkers.Get("kubebuilder:validation:Type").(crdmarkers.Type)
	if !ok || schemaType == "object" {
		return nil
	}
	return []error{fmt.Errorf("type %s has a custom %s method and +kubebuilder:validation:Type=%s, but applyconfiguration-gen would generate an apply configuration that is serialized as an object (https://github.com/kubernetes/kubernetes/issues/143415); there is no workaround other than disabling apply configurations for the API types that use it with +kubebuilder:ac:generate=false",
		t.Name, method, schemaType)}
}

// checkShadowedFields reports a field that shadows a field promoted from an
// embedded struct declared before it. applyconfiguration-gen generates one
// With<Field> function per name, in declaration order, so the function sets the
// promoted field, which encoding/json does not serialize.
//
// TODO: drop this check once controller-tools uses a code-generator with the fix
// for https://github.com/kubernetes/kubernetes/issues/143416 (controller-tools#1499).
func checkShadowedFields(t *types.Type) []error {
	type firstSeen struct {
		path  string
		depth int
	}
	seen := map[string]firstSeen{}
	var errs []error

	// walk follows the order in which applyconfiguration-gen generates the
	// With<Field> functions, see generateWithFuncs in code-generator.
	var walk func(st *types.Type, prefix string, depth int)
	walk = func(st *types.Type, prefix string, depth int) {
		for _, m := range st.Members {
			if blocklistedMember(st, m) || !hasJSONTag(m) {
				continue
			}
			if first, ok := seen[m.Name]; ok {
				if depth < first.depth {
					errs = append(errs, fmt.Errorf("field %s of type %s shadows the promoted field %s, and applyconfiguration-gen would generate a With%s function that sets the promoted field, so the value is lost (https://github.com/kubernetes/kubernetes/issues/143416); declare %s before the embedded field to work around it",
						prefix+m.Name, t.Name, first.path, m.Name, m.Name))
				}
				continue
			}
			seen[m.Name] = firstSeen{path: prefix + m.Name, depth: depth}
			if m.Embedded {
				walk(m.Type, prefix+m.Name+".", depth+1)
			}
		}
	}
	walk(structOf(t), "", 0)
	return errs
}

// customJSONMethod returns MarshalJSON or UnmarshalJSON if t or *t has it,
// including methods promoted from embedded fields, or "" otherwise.
func customJSONMethod(t *types.Type) string {
	for _, name := range []string{"MarshalJSON", "UnmarshalJSON"} {
		if t.GoType != nil {
			if gotypes.NewMethodSet(gotypes.NewPointer(t.GoType)).Lookup(nil, name) != nil {
				return name
			}
		} else if _, ok := t.Methods[name]; ok {
			return name
		}
	}
	return ""
}

func clientgenTags(t *types.Type) (clientgenutil.Tags, error) {
	return clientgenutil.ParseClientGenTags(append(t.SecondClosestCommentLines, t.CommentLines...))
}

func findReachable(t *types.Type, reachable map[types.Name]bool) {
	if reachable[t.Name] {
		return
	}
	reachable[t.Name] = true
	for _, next := range []*types.Type{t.Elem, t.Underlying, t.Key} {
		if next != nil {
			findReachable(next, reachable)
		}
	}
	for _, m := range t.Members {
		findReachable(m.Type, reachable)
	}
}

// structOf returns the struct type of t, following aliases, or nil.
func structOf(t *types.Type) *types.Type {
	for t.Kind == types.Alias {
		t = t.Underlying
	}
	if t.Kind != types.Struct {
		return nil
	}
	return t
}

func deref(t *types.Type) *types.Type {
	for t.Kind == types.Pointer {
		t = t.Elem
	}
	return t
}

func hasJSONTaggedMember(t *types.Type) bool {
	return slices.ContainsFunc(t.Members, hasJSONTag)
}

func hasJSONTag(m types.Member) bool {
	tag, ok := reflect.StructTag(m.Tags).Lookup("json")
	return ok && tag != "-"
}

var objectMetaName = types.Name{Package: "k8s.io/apimachinery/pkg/apis/meta/v1", Name: "ObjectMeta"}

func hasEmbeddedObjectMeta(t *types.Type) bool {
	return slices.ContainsFunc(t.Members, func(m types.Member) bool {
		return m.Embedded && m.Type.Name == objectMetaName
	})
}

// blocklistedMember returns true for the fields that applyconfiguration-gen
// leaves out of the apply configurations.
func blocklistedMember(t *types.Type, m types.Member) bool {
	if t.Name == objectMetaName && (m.Name == "ManagedFields" || m.Name == "SelfLink") {
		return true
	}
	return strings.HasPrefix(m.Name, "ZZZ_")
}

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}
