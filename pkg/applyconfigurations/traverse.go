/*
Copyright 2021 The Kubernetes Authors.

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

package applyconfigurations

import (
	"fmt"
	"go/ast"
	"go/types"

	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/util"
)

var (
	applyConfigAppendString = "ApplyConfiguration"
)

// namingInfo holds package and syntax for referencing a field, type,
// etc.  It's used to allow lazily marking import usage.
// You should generally retrieve the syntax using Syntax.
type namingInfo struct {
	// typeInfo is the type being named.
	typeInfo     types.Type
	nameOverride string
}

// Syntax calculates the code representation of the given type or name,
// and returns the apply representation
func (n *namingInfo) Syntax(universe *Universe, basePkg *loader.Package, imports *util.ImportsList) string {
	if n.nameOverride != "" {
		return n.nameOverride
	}
	switch typeInfo := n.typeInfo.(type) {
	case *types.Named:
		// register that we need an import for this type,
		// so we can get the appropriate alias to use.

		var lastType types.Type
		appendString := applyConfigAppendString
		for underlyingType := typeInfo.Underlying(); underlyingType != lastType; lastType, underlyingType = underlyingType, underlyingType.Underlying() {
			// ApplyConfigurations are not necessary for basic types
			if _, ok := underlyingType.(*types.Basic); ok {
				appendString = ""
			}
		}

		typeName := typeInfo.Obj()
		otherPkg := typeName.Pkg()

		// ApplyConfiguration is in the same package
		if otherPkg == basePkg.Types && universe.IsApplyConfigGenerated(typeInfo, otherPkg.Path()) {
			return typeName.Name() + appendString
		}

		// ApplyConfiguration is in different package
		path, isAc := universe.GetApplyConfigPath(typeInfo, otherPkg.Path())
		alias := imports.NeedImport(path)
		if !isAc {
			appendString = ""
		}
		return alias + "." + typeName.Name() + appendString
	case *types.Basic:
		return typeInfo.String()
	case *types.Pointer:
		return "*" + (&namingInfo{typeInfo: typeInfo.Elem()}).Syntax(universe, basePkg, imports)
	case *types.Slice:
		return "[]" + (&namingInfo{typeInfo: typeInfo.Elem()}).Syntax(universe, basePkg, imports)
	case *types.Map:
		return fmt.Sprintf(
			"map[%s]%s",
			(&namingInfo{typeInfo: typeInfo.Key()}).Syntax(universe, basePkg, imports),
			(&namingInfo{typeInfo: typeInfo.Elem()}).Syntax(universe, basePkg, imports))
	case *types.Interface:
		return "interface{}"
	case *types.Signature:
		return typeInfo.String()
	default:
		return typeInfo.String()
	}
}

// copyMethodMakers makes apply configurations for Go types,
// writing them to its codeWriter.
type applyConfigurationMaker struct {
	pkg *loader.Package
	*util.ImportsList
	*codeWriter
}

// GenerateTypesFor makes makes apply configuration types for the given type, when appropriate
func (c *applyConfigurationMaker) GenerateTypesFor(universe *Universe, root *loader.Package, info *markers.TypeInfo) {
	typeInfo := root.TypesInfo.TypeOf(info.RawSpec.Name)
	if typeInfo == types.Typ[types.Invalid] {
		root.AddError(loader.ErrFromNode(fmt.Errorf("unknown type: %s", info.Name), info.RawSpec))
	}
	if len(info.Fields) == 0 {
		return
	}

	c.Linef("// %sApplyConfiguration represents a declarative configuration of the %s type for use", info.Name, info.Name)
	c.Linef("// with apply.")
	c.Linef("type %sApplyConfiguration struct {", info.Name)
	for _, field := range info.Fields {
		fieldName := field.Name
		fieldType := root.TypesInfo.TypeOf(field.RawField.Type)
		fieldNamingInfo := namingInfo{typeInfo: fieldType}
		fieldTypeString := fieldNamingInfo.Syntax(universe, root, c.ImportsList)

		if tags, ok := lookupJSONTags(field); ok {
			if tags.inline {
				c.Linef("%s %s `json:\"%s\"`", fieldName, fieldTypeString, tags.String())
			} else if isPointer(fieldNamingInfo.typeInfo) || isMap(fieldNamingInfo.typeInfo) || isList(fieldNamingInfo.typeInfo) {
				tags.omitempty = true
				c.Linef("%s %s `json:\"%s\"`", fieldName, fieldTypeString, tags.String())
			} else {
				tags.omitempty = true
				c.Linef("%s *%s `json:\"%s\"`", fieldName, fieldTypeString, tags.String())
			}
		}
	}
	c.Linef("}")
}

// These functions implement a specific interface as required by controller-runtime
func (c *applyConfigurationMaker) GenerateRootFunctions(universe *Universe, root *loader.Package, info *markers.TypeInfo) {
	// For TypeMeta
	c.Linef("func (ac * %[1]sApplyConfiguration) WithKind (value string)  *%[1]sApplyConfiguration {", info.Name)
	c.Linef("ac.Kind = &value")
	c.Linef("return ac")
	c.Linef("}")

	c.Linef("func (ac * %[1]sApplyConfiguration) WithAPIVersion (value string)  *%[1]sApplyConfiguration {", info.Name)
	c.Linef("ac.APIVersion = &value")
	c.Linef("return ac")
	c.Linef("}")

	metav1ac := c.NeedImport("k8s.io/client-go/applyconfigurations/meta/v1")
	c.Linef("func (ac * %[1]sApplyConfiguration) ensureObjectMetaApplyConfigurationExists () {", info.Name)
	c.Linef("if ac.ObjectMetaApplyConfiguration == nil {")
	c.Linef("ac.ObjectMetaApplyConfiguration = &%s.ObjectMetaApplyConfiguration{}", metav1ac)
	c.Linef("}")
	c.Linef("}")

	// For ObjectMeta
	c.Linef("func (ac * %[1]sApplyConfiguration) WithName(value string)  *%[1]sApplyConfiguration {", info.Name)
	c.Linef("ac.ensureObjectMetaApplyConfigurationExists()")
	c.Linef("ac.Name = &value")
	c.Linef("return ac")
	c.Linef("}")

	c.Linef("func (ac * %[1]sApplyConfiguration) WithNamespace(value string)  *%[1]sApplyConfiguration {", info.Name)
	c.Linef("ac.Namespace = &value")
	c.Linef("return ac")
	c.Linef("}")

	c.Linef("func (ac * %[1]sApplyConfiguration) GetName() string {", info.Name)
	c.Linef("ac.ensureObjectMetaApplyConfigurationExists()")
	c.Linef("return *ac.Name")
	c.Linef("}")

	c.Linef("func (ac * %[1]sApplyConfiguration) GetNamespace() string {", info.Name)
	c.Linef("ac.ensureObjectMetaApplyConfigurationExists()")
	c.Linef("return *ac.Namespace")
	c.Linef("}")

}

func (c *applyConfigurationMaker) GenerateStructConstructor(root *loader.Package, info *markers.TypeInfo) {
	c.Linef("// %sApplyConfiguration represents a declarative configuration of the %s type for use", info.Name, info.Name)
	c.Linef("// with apply.")
	c.Linef("func %s() *%sApplyConfiguration {", info.Name, info.Name)
	c.Linef("return &%sApplyConfiguration{}", info.Name)
	c.Linef("}")
}

func (c *applyConfigurationMaker) GenerateRootStructConstructor(root *loader.Package, info *markers.TypeInfo, clusterScope bool, group, version string) {
	c.Linef("// %sApplyConfiguration represents a declarative configuration of the %s type for use", info.Name, info.Name)
	c.Linef("// with apply.")
	if clusterScope {
		c.Linef("func %s(name string) *%sApplyConfiguration {", info.Name, info.Name)
	} else {
		c.Linef("func %s(name, namespace string) *%sApplyConfiguration {", info.Name, info.Name)
	}
	c.Linef("ac := &%sApplyConfiguration{}", info.Name)
	c.Linef("ac.WithName(name)")
	if !clusterScope {
		c.Linef("ac.WithNamespace(namespace)")
	}
	c.Linef("ac.WithKind(\"%s\")", info.Name)
	c.Linef("ac.WithAPIVersion(\"%s/%s\")", group, version)
	c.Linef("return ac")
	c.Linef("}")
}

func isApplyConfig(t types.Type) bool {
	switch typeInfo := t.(type) {
	case *types.Named:
		return isApplyConfig(typeInfo.Underlying())
	case *types.Struct:
		return true
	case *types.Pointer:
		return isApplyConfig(typeInfo.Elem())
	default:
		return false
	}
}

func isList(t types.Type) bool {
	switch t.(type) {
	case *types.Slice:
		return true
	default:
		return false
	}
}

func isMap(t types.Type) bool {
	switch t.(type) {
	case *types.Map:
		return true
	default:
		return false
	}
}

func isPointer(t types.Type) bool {
	switch t.(type) {
	case *types.Pointer:
		return true
	default:
		return false
	}
}

func (c *applyConfigurationMaker) GenerateMemberSet(universe *Universe, field markers.FieldInfo, root *loader.Package, info *markers.TypeInfo) {
	fieldType := root.TypesInfo.TypeOf(field.RawField.Type)
	fieldNamingInfo := namingInfo{typeInfo: fieldType}
	fieldTypeString := fieldNamingInfo.Syntax(universe, root, c.ImportsList)

	c.Linef("// With%s sets the %s field in the declarative configuration to the given value", field.Name, field.Name)
	if isApplyConfig(fieldNamingInfo.typeInfo) && !isPointer(fieldNamingInfo.typeInfo) {
		fieldTypeString = "*" + fieldTypeString
	}
	c.Linef("func (b *%sApplyConfiguration) With%s(value %s) *%sApplyConfiguration {", info.Name, field.Name, fieldTypeString, info.Name)

	if tags, ok := lookupJSONTags(field); ok {
		if tags.inline {
			c.Linef("if value != nil {")
			c.Linef("b.%s = *value", field.Name)
			c.Linef("}")
		} else if isApplyConfig(fieldNamingInfo.typeInfo) {
			c.Linef("b.%s = value", field.Name)
		} else if isPointer(fieldNamingInfo.typeInfo) {
			c.Linef("b.%s = value", field.Name)
		} else {
			c.Linef("b.%s = &value", field.Name)
		}
	}
	c.Linef("return b")
	c.Linef("}")
}

func (c *applyConfigurationMaker) GenerateMemberSetForSlice(universe *Universe, field markers.FieldInfo, root *loader.Package, info *markers.TypeInfo) {
	fieldType := root.TypesInfo.TypeOf(field.RawField.Type)
	fieldNamingInfo := namingInfo{typeInfo: fieldType}

	sliceType := fieldType.(*types.Slice)
	listVal := (&namingInfo{typeInfo: sliceType.Elem()}).Syntax(universe, root, c.ImportsList)

	c.Linef("// With%s sets the %s field in the declarative configuration to the given value", field.Name, field.Name)
	c.Linef("func (b *%sApplyConfiguration) With%s(values ...%s) *%sApplyConfiguration {", info.Name, field.Name, listVal, info.Name)

	c.Linef("for i := range values {")
	if isApplyConfig(fieldNamingInfo.typeInfo) || isPointer(fieldNamingInfo.typeInfo) {
		c.Linef("if values[i] == nil {")
		c.Linef("panic(\"nil value passed to With%s\")", field.Name)
		c.Linef("}")
		c.Linef("b.%s[1] = append(b.%s[1], values[i]", field.Name)
	}
	c.Linef("b.%[1]s = append(b.%[1]s, values[i])", field.Name)
	c.Linef("}")
	c.Linef("return b")
	c.Linef("}")
}

func (c *applyConfigurationMaker) GenerateMemberSetForMap(universe *Universe, field markers.FieldInfo, root *loader.Package, info *markers.TypeInfo) {
	fieldType := root.TypesInfo.TypeOf(field.RawField.Type)
	fieldNamingInfo := namingInfo{typeInfo: fieldType}
	fieldTypeString := fieldNamingInfo.Syntax(universe, root, c.ImportsList)

	mapType := fieldType.(*types.Map)
	k := (&namingInfo{typeInfo: mapType.Key()}).Syntax(universe, root, c.ImportsList)
	v := (&namingInfo{typeInfo: mapType.Elem()}).Syntax(universe, root, c.ImportsList)

	c.Linef("// With%s sets the %s field in the declarative configuration to the given value", field.Name, field.Name)
	if isApplyConfig(fieldNamingInfo.typeInfo) && !isPointer(fieldNamingInfo.typeInfo) {
		fieldTypeString = "*" + fieldTypeString
	}
	c.Linef("func (b *%sApplyConfiguration) With%s(entries %s) *%sApplyConfiguration {", info.Name, field.Name, fieldTypeString, info.Name)
	c.Linef("if b.%s == nil && len(entries) > 0 {", field.Name)
	c.Linef("b.%s = make(map[%s]%s, len(entries))", field.Name, k, v)
	c.Linef("}")
	c.Linef("for k, v := range entries {")
	c.Linef("b.%s[k] = v", field.Name)
	c.Linef("}")
	c.Linef("return b")
	c.Linef("}")
}

func (c *applyConfigurationMaker) GenerateListMapAlias(root *loader.Package, info *markers.TypeInfo) {
	var listAlias = `
// %[1]sList represents a listAlias of %[1]sApplyConfiguration.
type %[1]sList []*%[1]sApplyConfiguration
`
	var mapAlias = `
// %[1]sMap represents a map of %[1]sApplyConfiguration.
type %[1]sMap map[string]%[1]sApplyConfiguration
`

	c.Linef(listAlias, info.Name)
	c.Linef(mapAlias, info.Name)
}

// shouldBeApplyConfiguration checks if we're supposed to make apply configurations for the given type.
//
// TODO(jpbetz): Copy over logic for inclusion from requiresApplyConfiguration
func shouldBeApplyConfiguration(pkg *loader.Package, info *markers.TypeInfo) bool {
	if !ast.IsExported(info.Name) {
		return false
	}

	typeInfo := pkg.TypesInfo.TypeOf(info.RawSpec.Name)
	if typeInfo == types.Typ[types.Invalid] {
		pkg.AddError(loader.ErrFromNode(fmt.Errorf("unknown type: %s", info.Name), info.RawSpec))
		return false
	}

	// according to gengo, everything named is an alias, except for an alias to a pointer,
	// which is just a pointer, afaict.  Just roll with it.
	if asPtr, isPtr := typeInfo.(*types.Named).Underlying().(*types.Pointer); isPtr {
		typeInfo = asPtr
	}

	lastType := typeInfo
	if _, isNamed := typeInfo.(*types.Named); isNamed {
		for underlyingType := typeInfo.Underlying(); underlyingType != lastType; lastType, underlyingType = underlyingType, underlyingType.Underlying() {
		}
	}

	// structs are the only thing that's not a basic that apply configurations are generated for
	_, isStruct := lastType.(*types.Struct)
	if !isStruct {
		return false
	}
	if _, ok := excludeTypes[info.Name]; ok { // TODO(jpbetz): What to do here?
		return false
	}
	var hasJSONTaggedMembers bool
	for _, field := range info.Fields {
		if _, ok := lookupJSONTags(field); ok {
			hasJSONTaggedMembers = true
		}
	}
	return hasJSONTaggedMembers
}

var (
	rawExtension = "k8s.io/apimachinery/pkg/runtime/RawExtension"
	unknown      = "k8s.io/apimachinery/pkg/runtime/Unknown"
)

// excludeTypes contains well known types that we do not generate apply configurations for.
// Hard coding because we only have two, very specific types that serve a special purpose
// in the type system here.
var excludeTypes = map[string]struct{}{
	rawExtension: {},
	unknown:      {},
	// DO NOT ADD TO THIS LIST. If we need to exclude other types, we should consider allowing the
	// go type declarations to be annotated as excluded from this generator.
}
