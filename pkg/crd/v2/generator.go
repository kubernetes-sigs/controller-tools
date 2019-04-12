/*
Copyright 2018 The Kubernetes Authors.

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

package v2

import (
	"fmt"
	"go/ast"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

const (
	defPrefix = "#/definitions/"
	inlineTag = "inline"
)

// Checks whether the typeName represents a simple json type

// Removes a character by replacing it with a space
func removeChar(str string, removedStr string) string {
	return strings.Replace(str, removedStr, " ", -1)
}

// This is a hacky function that does the one job of
// extracting the tag values in the structs
// Example struct:
// type MyType struct {
//   MyField string `yaml:"myField,omitempty"`
// }
//
// From the above example struct, we need to extract
// and return this: ("myField", "omitempty")
func extractFromTag(tag *ast.BasicLit) (string, string) {
	if tag == nil || tag.Value == "" {
		return "", ""
	}
	tagValue := tag.Value

	tagValue = removeChar(tagValue, "`")
	tagValue = removeChar(tagValue, `"`)
	tagValue = strings.TrimSpace(tagValue)

	var tagContent, tagKey string
	fmt.Sscanf(tagValue, `%s %s`, &tagKey, &tagContent)

	if tagKey != "json:" && tagKey != "yaml:" {
		return "", ""
	}

	if strings.Contains(tagContent, ",") {
		splitContent := strings.Split(tagContent, ",")
		return splitContent[0], splitContent[1]
	}
	return tagContent, ""
}

// exprToSchema converts ast.Expr to JSONSchemaProps
func (f *file) exprToSchema(t ast.Expr, doc string, comments []*ast.CommentGroup) (*v1beta1.JSONSchemaProps, []TypeReference) {
	var def *v1beta1.JSONSchemaProps
	var externalTypeRefs []TypeReference

	switch tt := t.(type) {
	case *ast.Ident:
		def = f.identToSchema(tt, comments)
	case *ast.ArrayType:
		def, externalTypeRefs = f.arrayTypeToSchema(tt, doc, comments)
	case *ast.MapType:
		def = f.mapTypeToSchema(tt, doc, comments)
	case *ast.SelectorExpr:
		def, externalTypeRefs = f.selectorExprToSchema(tt, comments)
	case *ast.StarExpr:
		def, externalTypeRefs = f.exprToSchema(tt.X, "", comments)
	case *ast.StructType:
		def, externalTypeRefs = f.structTypeToSchema(tt)
	case *ast.InterfaceType: // TODO: handle interface if necessary.
		return &v1beta1.JSONSchemaProps{}, []TypeReference{}
	}
	def.Description = filterDescription(doc)

	return def, externalTypeRefs
}

// identToSchema converts ast.Ident to JSONSchemaProps.
func (f *file) identToSchema(ident *ast.Ident, comments []*ast.CommentGroup) *v1beta1.JSONSchemaProps {
	def := &v1beta1.JSONSchemaProps{}
	if isSimpleType(ident.Name) {
		def.Type = jsonifyType(ident.Name)
	} else {
		def.Ref = getPrefixedDefLink(ident.Name, f.pkgPrefix)
	}
	processMarkersInComments(def, comments...)
	return def
}

// identToSchema converts ast.SelectorExpr to JSONSchemaProps.
func (f *file) selectorExprToSchema(selectorType *ast.SelectorExpr, comments []*ast.CommentGroup) (*v1beta1.JSONSchemaProps, []TypeReference) {
	pkgAlias := selectorType.X.(*ast.Ident).Name
	typeName := selectorType.Sel.Name

	typ := TypeReference{
		TypeName:    typeName,
		PackageName: f.importPaths[pkgAlias],
	}

	time := TypeReference{TypeName: "Time", PackageName: "k8s.io/apimachinery/pkg/apis/meta/v1"}
	duration := TypeReference{TypeName: "Duration", PackageName: "k8s.io/apimachinery/pkg/apis/meta/v1"}
	quantity := TypeReference{TypeName: "Quantity", PackageName: "k8s.io/apimachinery/pkg/api/resource"}
	unstructured := TypeReference{TypeName: "Unstructured", PackageName: "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"}
	rawExtension := TypeReference{TypeName: "RawExtension", PackageName: "k8s.io/apimachinery/pkg/runtime"}
	intOrString := TypeReference{TypeName: "IntOrString", PackageName: "k8s.io/apimachinery/pkg/util/intstr"}

	switch typ {
	case time:
		return &v1beta1.JSONSchemaProps{
			Type:   "string",
			Format: "date-time",
		}, []TypeReference{}
	case duration:
		return &v1beta1.JSONSchemaProps{
			Type: "string",
		}, []TypeReference{}
	case quantity:
		return &v1beta1.JSONSchemaProps{
			Type: "string",
		}, []TypeReference{}
	case unstructured, rawExtension:
		return &v1beta1.JSONSchemaProps{
			Type: "object",
		}, []TypeReference{}
	case intOrString:
		return &v1beta1.JSONSchemaProps{
			AnyOf: []v1beta1.JSONSchemaProps{
				{
					Type: "string",
				},
				{
					Type: "integer",
				},
			},
		}, []TypeReference{}
	}

	def := &v1beta1.JSONSchemaProps{
		Ref: getPrefixedDefLink(typeName, f.importPaths[pkgAlias]),
	}
	processMarkersInComments(def, comments...)
	return def, []TypeReference{{TypeName: typeName, PackageName: pkgAlias}}
}

// arrayTypeToSchema converts ast.ArrayType to JSONSchemaProps by examining the elements in the array.
func (f *file) arrayTypeToSchema(arrayType *ast.ArrayType, doc string, comments []*ast.CommentGroup) (*v1beta1.JSONSchemaProps, []TypeReference) {
	// not passing doc down to exprToSchema
	items, extRefs := f.exprToSchema(arrayType.Elt, "", comments)
	processMarkersInComments(items, comments...)

	def := &v1beta1.JSONSchemaProps{
		Type:        "array",
		Items:       &v1beta1.JSONSchemaPropsOrArray{Schema: items},
		Description: doc,
	}

	// TODO: clear the schema on the parent level, since it is on the children level.

	return def, extRefs
}

// mapTypeToSchema converts ast.MapType to JSONSchemaProps.
func (f *file) mapTypeToSchema(mapType *ast.MapType, doc string, comments []*ast.CommentGroup) *v1beta1.JSONSchemaProps {
	def := &v1beta1.JSONSchemaProps{}
	switch mapType.Value.(type) {
	case *ast.Ident:
		valueType := mapType.Value.(*ast.Ident)
		if def.AdditionalProperties == nil {
			def.AdditionalProperties = &v1beta1.JSONSchemaPropsOrBool{}
		}
		def.AdditionalProperties.Schema = new(v1beta1.JSONSchemaProps)

		if isSimpleType(valueType.Name) {
			def.AdditionalProperties.Schema.Type = valueType.Name
		} else {
			def.AdditionalProperties.Schema.Ref = getPrefixedDefLink(valueType.Name, f.pkgPrefix)
		}
	case *ast.InterfaceType:
		// No op
		panic("Map Interface Type")
	}
	def.Type = "object"
	def.Description = doc
	processMarkersInComments(def, comments...)
	return def
}

// structTypeToSchema converts ast.StructType to JSONSchemaProps by examining each field in the struct.
func (f *file) structTypeToSchema(structType *ast.StructType) (*v1beta1.JSONSchemaProps, []TypeReference) {
	def := &v1beta1.JSONSchemaProps{
		Type: "object",
	}
	externalTypeRefs := []TypeReference{}
	for _, field := range structType.Fields.List {
		yamlName, option := extractFromTag(field.Tag)

		if (yamlName == "" && option != inlineTag) || yamlName == "-" {
			continue
		}

		if option != inlineTag && option != "omitempty" {
			def.Required = append(def.Required, yamlName)
		}

		if def.Properties == nil {
			def.Properties = make(map[string]v1beta1.JSONSchemaProps)
		}

		propDef, propExternalTypeDefs := f.exprToSchema(field.Type, field.Doc.Text(), f.commentMap[field])

		externalTypeRefs = append(externalTypeRefs, propExternalTypeDefs...)

		if option == inlineTag {
			def.AllOf = append(def.AllOf, *propDef)
			continue
		}

		def.Properties[yamlName] = *propDef
	}

	return def, externalTypeRefs
}

func getReachableTypes(startingTypes map[string]bool, definitions v1beta1.JSONSchemaDefinitions) map[string]bool {
	pruner := DefinitionPruner{definitions, startingTypes}
	prunedTypes := pruner.Prune(true)
	return prunedTypes
}

type file struct {
	// name prefix of the package
	pkgPrefix string
	// importPaths contains a map from import alias to the import path for the file.
	importPaths map[string]string
	// commentMap is comment mapping for this file.
	commentMap ast.CommentMap
}
