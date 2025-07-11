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
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// Schema flattening is done in a recursive mapping method.
// Start reading at infoToSchema.

const (
	// defPrefix is the prefix used to link to definitions in the OpenAPI schema.
	defPrefix = "#/definitions/"
)

// byteType is the types.Type for byte (see the types documention
// for why we need to look this up in the Universe), saved
// for quick comparison.
var byteType = types.Universe.Lookup("byte").Type()

// SchemaMarker is any marker that needs to modify the schema of the underlying type or field.
type SchemaMarker interface {
	// ApplyToSchema is called after the rest of the schema for a given type
	// or field is generated, to modify the schema appropriately.
	ApplyToSchema(*apiext.JSONSchemaProps) error
}

// applyFirstMarker is applied before any other markers.  It's a bit of a hack.
type applyFirstMarker interface {
	ApplyFirst()
}

// schemaRequester knows how to marker that another schema (e.g. via an external reference) is necessary.
type schemaRequester interface {
	NeedSchemaFor(typ TypeIdent)
}

// schemaContext stores and provides information across a hierarchy of schema generation.
type schemaContext struct {
	pkg  *loader.Package
	info *markers.TypeInfo

	schemaRequester schemaRequester
	PackageMarkers  markers.MarkerValues

	allowDangerousTypes    bool
	ignoreUnexportedFields bool
}

// newSchemaContext constructs a new schemaContext for the given package and schema requester.
// It must have type info added before use via ForInfo.
func newSchemaContext(pkg *loader.Package, req schemaRequester, allowDangerousTypes, ignoreUnexportedFields bool) *schemaContext {
	pkg.NeedTypesInfo()
	return &schemaContext{
		pkg:                    pkg,
		schemaRequester:        req,
		allowDangerousTypes:    allowDangerousTypes,
		ignoreUnexportedFields: ignoreUnexportedFields,
	}
}

// ForInfo produces a new schemaContext with containing the same information
// as this one, except with the given type information.
func (c *schemaContext) ForInfo(info *markers.TypeInfo) *schemaContext {
	return &schemaContext{
		pkg:                    c.pkg,
		info:                   info,
		schemaRequester:        c.schemaRequester,
		allowDangerousTypes:    c.allowDangerousTypes,
		ignoreUnexportedFields: c.ignoreUnexportedFields,
	}
}

// requestSchema asks for the schema for a type in the package with the
// given import path.
func (c *schemaContext) requestSchema(pkgPath, typeName string) {
	pkg := c.pkg
	if pkgPath != "" {
		pkg = c.pkg.Imports()[pkgPath]
	}
	c.schemaRequester.NeedSchemaFor(TypeIdent{
		Package: pkg,
		Name:    typeName,
	})
}

// infoToSchema creates a schema for the type in the given set of type information.
func infoToSchema(ctx *schemaContext) *apiext.JSONSchemaProps {
	if obj := ctx.pkg.Types.Scope().Lookup(ctx.info.Name); obj != nil {
		switch {
		// If the obj implements a JSON marshaler and has a marker, use the
		// markers value and do not traverse as the marshaler could be doing
		// anything. If there is no marker, fall back to traversing.
		case implements(obj.Type(), jsonMarshaler):
			schema := &apiext.JSONSchemaProps{}
			applyMarkers(ctx, ctx.info.Markers, schema, ctx.info.RawSpec.Type)
			if schema.Type != "" {
				return schema
			}

		// If the obj implements a text marshaler, encode it as a string.
		case implements(obj.Type(), textMarshaler):
			schema := &apiext.JSONSchemaProps{Type: "string"}
			applyMarkers(ctx, ctx.info.Markers, schema, ctx.info.RawSpec.Type)
			if schema.Type != "string" {
				err := fmt.Errorf("%q implements encoding.TextMarshaler but schema type is not string: %q", ctx.info.RawSpec.Name, schema.Type)
				ctx.pkg.AddError(loader.ErrFromNode(err, ctx.info.RawSpec.Type))
			}
			return schema
		}
	}
	return typeToSchema(ctx, ctx.info.RawSpec.Type)
}

type schemaMarkerWithName struct {
	SchemaMarker SchemaMarker
	Name         string
}

// applyMarkers applies schema markers given their priority to the given schema
func applyMarkers(ctx *schemaContext, markerSet markers.MarkerValues, props *apiext.JSONSchemaProps, node ast.Node) {
	markers := make([]schemaMarkerWithName, 0, len(markerSet))
	itemsMarkers := make([]schemaMarkerWithName, 0, len(markerSet))

	for markerName, markerValues := range markerSet {
		for _, markerValue := range markerValues {
			if schemaMarker, isSchemaMarker := markerValue.(SchemaMarker); isSchemaMarker {
				if strings.HasPrefix(markerName, crdmarkers.ValidationItemsPrefix) {
					itemsMarkers = append(itemsMarkers, schemaMarkerWithName{
						SchemaMarker: schemaMarker,
						Name:         markerName,
					})
				} else {
					markers = append(markers, schemaMarkerWithName{
						SchemaMarker: schemaMarker,
						Name:         markerName,
					})
				}
			}
		}
	}

	cmpPriority := func(markers []schemaMarkerWithName, i, j int) bool {
		var iPriority, jPriority crdmarkers.ApplyPriority

		switch m := markers[i].SchemaMarker.(type) {
		case crdmarkers.ApplyPriorityMarker:
			iPriority = m.ApplyPriority()
		case applyFirstMarker:
			iPriority = crdmarkers.ApplyPriorityFirst
		default:
			iPriority = crdmarkers.ApplyPriorityDefault
		}

		switch m := markers[j].SchemaMarker.(type) {
		case crdmarkers.ApplyPriorityMarker:
			jPriority = m.ApplyPriority()
		case applyFirstMarker:
			jPriority = crdmarkers.ApplyPriorityFirst
		default:
			jPriority = crdmarkers.ApplyPriorityDefault
		}

		return iPriority < jPriority
	}
	sort.Slice(markers, func(i, j int) bool { return cmpPriority(markers, i, j) })
	sort.Slice(itemsMarkers, func(i, j int) bool { return cmpPriority(itemsMarkers, i, j) })

	for _, schemaMarker := range markers {
		if err := schemaMarker.SchemaMarker.ApplyToSchema(props); err != nil {
			ctx.pkg.AddError(loader.ErrFromNode(err /* an okay guess */, node))
		}
	}

	for _, schemaMarker := range itemsMarkers {
		if props.Type != "array" || props.Items == nil || props.Items.Schema == nil {
			err := fmt.Errorf("must apply %s to an array value, found %s", schemaMarker.Name, props.Type)
			ctx.pkg.AddError(loader.ErrFromNode(err, node))
		} else {
			itemsSchema := props.Items.Schema
			if err := schemaMarker.SchemaMarker.ApplyToSchema(itemsSchema); err != nil {
				ctx.pkg.AddError(loader.ErrFromNode(err /* an okay guess */, node))
			}
		}
	}
}

// typeToSchema creates a schema for the given AST type.
func typeToSchema(ctx *schemaContext, rawType ast.Expr) *apiext.JSONSchemaProps {
	var props *apiext.JSONSchemaProps
	switch expr := rawType.(type) {
	case *ast.Ident:
		props = localNamedToSchema(ctx, expr)
	case *ast.SelectorExpr:
		props = namedToSchema(ctx, expr)
	case *ast.ArrayType:
		props = arrayToSchema(ctx, expr)
	case *ast.MapType:
		props = mapToSchema(ctx, expr)
	case *ast.StarExpr:
		props = typeToSchema(ctx, expr.X)
	case *ast.StructType:
		props = structToSchema(ctx, expr)
	default:
		ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("unsupported AST kind %T", expr), rawType))
		// NB(directxman12): we explicitly don't handle interfaces
		return &apiext.JSONSchemaProps{}
	}

	props.Description = ctx.info.Doc

	applyMarkers(ctx, ctx.info.Markers, props, rawType)

	return props
}

// qualifiedName constructs a JSONSchema-safe qualified name for a type
// (`<typeName>` or `<safePkgPath>~0<typeName>`, where `<safePkgPath>`
// is the package path with `/` replaced by `~1`, according to JSONPointer
// escapes).
func qualifiedName(pkgName, typeName string) string {
	if pkgName != "" {
		return strings.ReplaceAll(pkgName, "/", "~1") + "~0" + typeName
	}
	return typeName
}

// TypeRefLink creates a definition link for the given type and package.
func TypeRefLink(pkgName, typeName string) string {
	return defPrefix + qualifiedName(pkgName, typeName)
}

// localNamedToSchema creates a schema (ref) for a *potentially* local type reference
// (could be external from a dot-import).
func localNamedToSchema(ctx *schemaContext, ident *ast.Ident) *apiext.JSONSchemaProps {
	typeInfo := ctx.pkg.TypesInfo.TypeOf(ident)
	if typeInfo == types.Typ[types.Invalid] {
		ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("unknown type %s", ident.Name), ident))
		return &apiext.JSONSchemaProps{}
	}
	// This reproduces the behavior we had pre gotypesalias=1 (needed if this
	// project is compiled with default settings and Go >= 1.23).
	if aliasInfo, isAlias := typeInfo.(*types.Alias); isAlias {
		typeInfo = aliasInfo.Rhs()
	}
	if basicInfo, isBasic := typeInfo.(*types.Basic); isBasic {
		typ, fmt, err := builtinToType(basicInfo, ctx.allowDangerousTypes)
		if err != nil {
			ctx.pkg.AddError(loader.ErrFromNode(err, ident))
		}
		// Check for type aliasing to a basic type for gotypesalias=0. See more
		// in documentation https://pkg.go.dev/go/types#Alias:
		// > For gotypesalias=1, alias declarations produce an Alias type.
		// > Otherwise, the alias information is only in the type name, which
		// > points directly to the actual (aliased) type.
		if basicInfo.Name() != ident.Name {
			ctx.requestSchema("", ident.Name)
			link := TypeRefLink("", ident.Name)
			return &apiext.JSONSchemaProps{
				Type:   typ,
				Format: fmt,
				Ref:    &link,
			}
		}
		return &apiext.JSONSchemaProps{
			Type:   typ,
			Format: fmt,
		}
	}
	// NB(directxman12): if there are dot imports, this might be an external reference,
	// so use typechecking info to get the actual object
	typeNameInfo := typeInfo.(interface{ Obj() *types.TypeName }).Obj()
	pkg := typeNameInfo.Pkg()
	pkgPath := loader.NonVendorPath(pkg.Path())
	if pkg == ctx.pkg.Types {
		pkgPath = ""
	}
	ctx.requestSchema(pkgPath, typeNameInfo.Name())
	link := TypeRefLink(pkgPath, typeNameInfo.Name())
	return &apiext.JSONSchemaProps{
		Ref: &link,
	}
}

// namedSchema creates a schema (ref) for an explicitly external type reference.
func namedToSchema(ctx *schemaContext, named *ast.SelectorExpr) *apiext.JSONSchemaProps {
	typeInfoRaw := ctx.pkg.TypesInfo.TypeOf(named)
	if typeInfoRaw == types.Typ[types.Invalid] {
		ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("unknown type %v.%s", named.X, named.Sel.Name), named))
		return &apiext.JSONSchemaProps{}
	}
	typeInfo := typeInfoRaw.(interface{ Obj() *types.TypeName })
	typeNameInfo := typeInfo.Obj()
	nonVendorPath := loader.NonVendorPath(typeNameInfo.Pkg().Path())
	ctx.requestSchema(nonVendorPath, typeNameInfo.Name())
	link := TypeRefLink(nonVendorPath, typeNameInfo.Name())
	return &apiext.JSONSchemaProps{
		Ref: &link,
	}
	// NB(directxman12): we special-case things like resource.Quantity during the "collapse" phase.
}

// arrayToSchema creates a schema for the items of the given array, dealing appropriately
// with the special `[]byte` type (according to OpenAPI standards).
func arrayToSchema(ctx *schemaContext, array *ast.ArrayType) *apiext.JSONSchemaProps {
	eltType := ctx.pkg.TypesInfo.TypeOf(array.Elt)
	if eltType == byteType && array.Len == nil {
		// byte slices are represented as base64-encoded strings
		// (the format is defined in OpenAPI v3, but not JSON Schema)
		return &apiext.JSONSchemaProps{
			Type:   "string",
			Format: "byte",
		}
	}
	// TODO(directxman12): backwards-compat would require access to markers from base info
	items := typeToSchema(ctx.ForInfo(&markers.TypeInfo{}), array.Elt)

	return &apiext.JSONSchemaProps{
		Type:  "array",
		Items: &apiext.JSONSchemaPropsOrArray{Schema: items},
	}
}

// mapToSchema creates a schema for items of the given map.  Key types must eventually resolve
// to string (other types aren't allowed by JSON, and thus the kubernetes API standards).
func mapToSchema(ctx *schemaContext, mapType *ast.MapType) *apiext.JSONSchemaProps {
	keyInfo := ctx.pkg.TypesInfo.TypeOf(mapType.Key)
	// check that we've got a type that actually corresponds to a string
	for keyInfo != nil {
		switch typedKey := keyInfo.(type) {
		case *types.Basic:
			if typedKey.Info()&types.IsString == 0 {
				ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("map keys must be strings, not %s", keyInfo.String()), mapType.Key))
				return &apiext.JSONSchemaProps{}
			}
			keyInfo = nil // stop iterating
		case *types.Named:
			keyInfo = typedKey.Underlying()
		default:
			ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("map keys must be strings, not %s", keyInfo.String()), mapType.Key))
			return &apiext.JSONSchemaProps{}
		}
	}

	// TODO(directxman12): backwards-compat would require access to markers from base info
	var valSchema *apiext.JSONSchemaProps
	switch val := mapType.Value.(type) {
	case *ast.Ident:
		valSchema = localNamedToSchema(ctx.ForInfo(&markers.TypeInfo{}), val)
	case *ast.SelectorExpr:
		valSchema = namedToSchema(ctx.ForInfo(&markers.TypeInfo{}), val)
	case *ast.ArrayType:
		valSchema = arrayToSchema(ctx.ForInfo(&markers.TypeInfo{}), val)
	case *ast.StarExpr:
		valSchema = typeToSchema(ctx.ForInfo(&markers.TypeInfo{}), val)
	case *ast.MapType:
		valSchema = typeToSchema(ctx.ForInfo(&markers.TypeInfo{}), val)
	default:
		ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("not a supported map value type: %T", mapType.Value), mapType.Value))
		return &apiext.JSONSchemaProps{}
	}

	return &apiext.JSONSchemaProps{
		Type: "object",
		AdditionalProperties: &apiext.JSONSchemaPropsOrBool{
			Schema: valSchema,
			Allows: true, /* set automatically by serialization, but useful for testing */
		},
	}
}

// structToSchema creates a schema for the given struct.  Embedded fields are placed in AllOf,
// and can be flattened later with a Flattener.
//
//nolint:gocyclo
func structToSchema(ctx *schemaContext, structType *ast.StructType) *apiext.JSONSchemaProps {
	props := &apiext.JSONSchemaProps{
		Type:       "object",
		Properties: make(map[string]apiext.JSONSchemaProps),
	}

	if ctx.info.RawSpec.Type != structType {
		ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("encountered non-top-level struct (possibly embedded), those aren't allowed"), structType))
		return props
	}

	exactlyOneOf, err := oneOfValuesToSet(ctx.info.Markers[crdmarkers.ValidationExactlyOneOfPrefix])
	if err != nil {
		ctx.pkg.AddError(loader.ErrFromNode(err, structType))
		return props
	}
	atMostOneOf, err := oneOfValuesToSet(ctx.info.Markers[crdmarkers.ValidationAtMostOneOfPrefix])
	if err != nil {
		ctx.pkg.AddError(loader.ErrFromNode(err, structType))
		return props
	}

	for _, field := range ctx.info.Fields {
		// Skip if the field is not an inline field, ignoreUnexportedFields is true, and the field is not exported
		if field.Name != "" && ctx.ignoreUnexportedFields && !ast.IsExported(field.Name) {
			continue
		}

		jsonTag, hasTag := field.Tag.Lookup("json")
		if !hasTag {
			// if the field doesn't have a JSON tag, it doesn't belong in output (and shouldn't exist in a serialized type)
			ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("encountered struct field %q without JSON tag in type %q", field.Name, ctx.info.Name), field.RawField))
			continue
		}
		jsonOpts := strings.Split(jsonTag, ",")
		if len(jsonOpts) == 1 && jsonOpts[0] == "-" {
			// skipped fields have the tag "-" (note that "-," means the field is named "-")
			continue
		}

		inline := false
		omitEmpty := false
		for _, opt := range jsonOpts[1:] {
			switch opt {
			case "inline":
				inline = true
			case "omitempty":
				omitEmpty = true
			}
		}
		fieldName := jsonOpts[0]
		inline = inline || fieldName == "" // anonymous fields are inline fields in YAML/JSON

		// if no default required mode is set, default to required
		defaultMode := "required"
		if ctx.PackageMarkers.Get("kubebuilder:validation:Optional") != nil {
			defaultMode = "optional"
		}

		switch {
		case field.Markers.Get("kubebuilder:validation:Optional") != nil:
			// explicitly optional - kubebuilder
		case field.Markers.Get("kubebuilder:validation:Required") != nil:
			if exactlyOneOf.Has(fieldName) || atMostOneOf.Has(fieldName) {
				ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("field %s is part of OneOf constraint and cannot be marked as required", fieldName), structType))
				return props
			}
			// explicitly required - kubebuilder
			props.Required = append(props.Required, fieldName)
		case field.Markers.Get("optional") != nil:
			// explicitly optional - kubernetes
		case field.Markers.Get("required") != nil:
			if exactlyOneOf.Has(fieldName) || atMostOneOf.Has(fieldName) {
				ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("field %s is part of OneOf constraint and cannot be marked as required", fieldName), structType))
				return props
			}
			// explicitly required - kubernetes
			props.Required = append(props.Required, fieldName)

		// if this package isn't set to optional default...
		case defaultMode == "required":
			// ...everything that's not inline / omitempty is required
			if !inline && !omitEmpty {
				if exactlyOneOf.Has(fieldName) || atMostOneOf.Has(fieldName) {
					ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("field %s is part of OneOf constraint and must have omitempty tag", fieldName), structType))
					return props
				}
				props.Required = append(props.Required, fieldName)
			}

		// if this package isn't set to required default...
		case defaultMode == "optional":
			// implicitly optional
		}

		var propSchema *apiext.JSONSchemaProps
		if field.Markers.Get(crdmarkers.SchemalessName) != nil {
			propSchema = &apiext.JSONSchemaProps{}
		} else {
			propSchema = typeToSchema(ctx.ForInfo(&markers.TypeInfo{}), field.RawField.Type)
		}
		propSchema.Description = field.Doc

		applyMarkers(ctx, field.Markers, propSchema, field.RawField)

		if inline {
			props.AllOf = append(props.AllOf, *propSchema)
			continue
		}

		props.Properties[fieldName] = *propSchema
	}

	return props
}

func oneOfValuesToSet(oneOfGroups []any) (sets.Set[string], error) {
	set := sets.New[string]()
	for _, oneOf := range oneOfGroups {
		switch vals := oneOf.(type) {
		case crdmarkers.ExactlyOneOf:
			if err := validateOneOfValues(vals...); err != nil {
				return nil, fmt.Errorf("%s: %w", crdmarkers.ValidationExactlyOneOfPrefix, err)
			}
			set.Insert(vals...)
		case crdmarkers.AtMostOneOf:
			if err := validateOneOfValues(vals...); err != nil {
				return nil, fmt.Errorf("%s: %w", crdmarkers.ValidationAtMostOneOfPrefix, err)
			}
			set.Insert(vals...)
		default:
			return nil, fmt.Errorf("expected ExactlyOneOf or AtMostOneOf, got %T", oneOf)
		}
	}
	return set, nil
}

func validateOneOfValues(fields ...string) error {
	var invalid []string
	for _, field := range fields {
		if strings.Contains(field, ".") {
			// nested fields are not allowed in OneOf validation markers
			invalid = append(invalid, field)
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf("cannot reference nested fields: %s", strings.Join(invalid, ","))
	}
	return nil
}

// builtinToType converts builtin basic types to their equivalent JSON schema form.
// It *only* handles types allowed by the kubernetes API standards. Floats are not
// allowed unless allowDangerousTypes is true
func builtinToType(basic *types.Basic, allowDangerousTypes bool) (typ string, format string, err error) {
	// NB(directxman12): formats from OpenAPI v3 are slightly different than those defined
	// in JSONSchema.  This'll use the OpenAPI v3 ones, since they're useful for bounding our
	// non-string types.
	basicInfo := basic.Info()
	switch {
	case basicInfo&types.IsBoolean != 0:
		typ = "boolean"
	case basicInfo&types.IsString != 0:
		typ = "string"
	case basicInfo&types.IsInteger != 0:
		typ = "integer"
	case basicInfo&types.IsFloat != 0:
		if allowDangerousTypes {
			typ = "number"
		} else {
			return "", "", errors.New("found float, the usage of which is highly discouraged, as support for them varies across languages. Please consider serializing your float as string instead. If you are really sure you want to use them, re-run with crd:allowDangerousTypes=true")
		}
	default:
		return "", "", fmt.Errorf("unsupported type %q", basic.String())
	}

	switch basic.Kind() {
	case types.Int32, types.Uint32:
		format = "int32"
	case types.Int64, types.Uint64:
		format = "int64"
	}

	return typ, format, nil
}

// Open coded go/types representation of encoding/json.Marshaller
var jsonMarshaler = types.NewInterfaceType([]*types.Func{
	types.NewFunc(token.NoPos, nil, "MarshalJSON",
		types.NewSignatureType(nil, nil, nil, nil,
			types.NewTuple(
				types.NewVar(token.NoPos, nil, "", types.NewSlice(types.Universe.Lookup("byte").Type())),
				types.NewVar(token.NoPos, nil, "", types.Universe.Lookup("error").Type())), false)),
}, nil).Complete()

// Open coded go/types representation of encoding.TextMarshaler
var textMarshaler = types.NewInterfaceType([]*types.Func{
	types.NewFunc(token.NoPos, nil, "MarshalText",
		types.NewSignatureType(nil, nil, nil, nil,
			types.NewTuple(
				types.NewVar(token.NoPos, nil, "text", types.NewSlice(types.Universe.Lookup("byte").Type())),
				types.NewVar(token.NoPos, nil, "err", types.Universe.Lookup("error").Type())), false)),
}, nil).Complete()

func implements(typ types.Type, iface *types.Interface) bool {
	return types.Implements(typ, iface) || types.Implements(types.NewPointer(typ), iface)
}
