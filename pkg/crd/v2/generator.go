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
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/afero"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func (pr *prsr) parseTypesInFile(filePath string, curPkgPrefix string, skipCRD bool) (
	v1beta1.JSONSchemaDefinitions, ExternalReferences, crdSpecByKind) {
	// Open the input go file and parse the Abstract Syntax Tree
	fset := token.NewFileSet()
	srcFile, err := pr.fs.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	node, err := parser.ParseFile(fset, filePath, srcFile, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	if !skipCRD {
		// process top-level (not tied to a struct field) markers.
		// e.g. group name marker +groupName=<group-name>
		pr.processTopLevelMarkers(node.Comments)
	}

	definitions := make(v1beta1.JSONSchemaDefinitions)
	externalRefs := make(ExternalReferences)

	// Parse import statements to get "alias: pkgName" mapping
	importPaths := make(map[string]string)
	for _, importItem := range node.Imports {
		pathValue := strings.Trim(importItem.Path.Value, "\"")
		if importItem.Name != nil {
			// Process aliased import
			importPaths[importItem.Name.Name] = pathValue
		} else if strings.Contains(pathValue, "/") {
			// Process unnamed imports with "/"
			segments := strings.Split(pathValue, "/")
			importPaths[segments[len(segments)-1]] = pathValue
		} else {
			importPaths[pathValue] = pathValue
		}
	}

	// Create an ast.CommentMap from the ast.File's comments.
	// This helps keeping the association between comments and AST nodes.
	// TODO: if necessary, support our own rules of comments ownership, golang's
	// builtin rules are listed at https://golang.org/pkg/go/ast/#NewCommentMap.
	// It seems it can meet our need at the moment.
	cmap := ast.NewCommentMap(fset, node, node.Comments)

	f := &file{
		pkgPrefix:   curPkgPrefix,
		importPaths: importPaths,
		commentMap:  cmap,
	}

	crdSpecs := crdSpecByKind{}
	for i := range node.Decls {
		declaration, ok := node.Decls[i].(*ast.GenDecl)
		if !ok {
			continue
		}

		// Skip it if it's not type declaration.
		if declaration.Tok != token.TYPE {
			continue
		}

		// We support the following format
		// // TreeNode doc
		// type TreeNode struct {
		//   left, right *TreeNode
		//   value *Comparable
		// }
		// but not
		// type (
		//   // Point doc
		//   Point struct{ x, y float64 }
		//   // Point2 doc
		//   Point2 struct{ x, y int }
		// )
		// since the latter format is rarely used in k8s.
		if len(declaration.Specs) != 1 {
			continue
		}
		ts := declaration.Specs[0]
		typeSpec, ok := ts.(*ast.TypeSpec)
		if !ok {
			fmt.Printf("spec type is: %T\n", ts)
			continue
		}

		typeName := typeSpec.Name.Name
		typeDescription := declaration.Doc.Text()

		fmt.Println("Generating schema definition for type:", typeName)
		def, refTypes := f.exprToSchema(typeSpec.Type, typeDescription, []*ast.CommentGroup{})
		definitions[getFullName(typeName, curPkgPrefix)] = *def
		externalRefs[getFullName(typeName, curPkgPrefix)] = refTypes

		var comments []string
		for _, c := range f.commentMap[node.Decls[i]] {
			comments = append(comments, strings.Split(c.Text(), "\n")...)
		}

		if !skipCRD {
			crdSpec := parseCRDs(comments)
			if crdSpec != nil {
				crdSpec.Names.Kind = typeName
				gk := schema.GroupKind{Kind: typeName}
				crdSpecs[gk] = crdSpec
				// TODO: validate the CRD spec for one version.
			}
		}
	}

	// Overwrite import aliases with actual package names
	for typeName := range externalRefs {
		for i, ref := range externalRefs[typeName] {
			externalRefs[typeName][i].PackageName = importPaths[ref.PackageName]
		}
	}

	return definitions, externalRefs, crdSpecs
}

// processTopLevelMarkers process top-level (not tied to a struct field) markers.
// e.g. group name marker +groupName=<group-name>
func (pr *prsr) processTopLevelMarkers(comments []*ast.CommentGroup) {
	for _, c := range comments {
		commentLines := strings.Split(c.Text(), "\n")
		cs := Comments(commentLines)
		if cs.hasTag("groupName") {
			group := cs.getTag("groupName", "=")
			if len(group) == 0 {
				log.Fatalf("can't use an empty name for the +groupName marker")
			}
			if pr.generatorOptions != nil && len(pr.generatorOptions.group) > 0 && group != pr.generatorOptions.group {
				log.Fatalf("can't have different group names %q and %q one package", pr.generatorOptions.group, group)
			}
			if pr.generatorOptions == nil {
				pr.generatorOptions = &toplevelGeneratorOptions{group: group}
			} else {
				pr.generatorOptions.group = group
			}
		}
	}
}

// mock this in testing.
var listFiles = func(pkgPath string) (string, []string, error) {
	pkg, err := build.Import(pkgPath, "", 0)
	return pkg.Dir, pkg.GoFiles, err
}

func (pr *prsr) parseTypesInPackage(pkgName string, referencedTypes map[string]bool, rootPackage, skipCRD bool) (
	v1beta1.JSONSchemaDefinitions, crdSpecByKind) {
	pkgDefs := make(v1beta1.JSONSchemaDefinitions)
	pkgExternalTypes := make(ExternalReferences)
	pkgCRDSpecs := make(crdSpecByKind)

	pkgDir, listOfFiles, err := listFiles(pkgName)
	if err != nil {
		log.Fatal(err)
	}

	pkgPrefix := strings.Replace(pkgName, "/", ".", -1)
	if rootPackage {
		pkgPrefix = ""
	}
	fmt.Println("pkgPrefix=", pkgPrefix)
	for _, fileName := range listOfFiles {
		fmt.Println("Processing file ", fileName)
		fileDefs, fileExternalRefs, fileCRDSpecs := pr.parseTypesInFile(filepath.Join(pkgDir, fileName), pkgPrefix, skipCRD)
		mergeDefs(pkgDefs, fileDefs)
		mergeExternalRefs(pkgExternalTypes, fileExternalRefs)
		mergeCRDSpecs(pkgCRDSpecs, fileCRDSpecs)
	}

	// Add pkg prefix to referencedTypes
	newReferencedTypes := make(map[string]bool)
	for key := range referencedTypes {
		altKey := getFullName(key, pkgPrefix)
		newReferencedTypes[altKey] = referencedTypes[key]
	}
	referencedTypes = newReferencedTypes

	fmt.Println("referencedTypes")
	debugPrint(referencedTypes)

	allReachableTypes := getReachableTypes(referencedTypes, pkgDefs)
	for key := range pkgDefs {
		if _, exists := allReachableTypes[key]; !exists {
			delete(pkgDefs, key)
			delete(pkgExternalTypes, key)
		}
	}
	fmt.Println("allReachableTypes")
	debugPrint(allReachableTypes)
	fmt.Println("pkgDefs")
	debugPrint(pkgDefs)
	fmt.Println("pkgExternalTypes")
	debugPrint(pkgExternalTypes)

	uniquePkgTypeRefs := make(map[string]map[string]bool)
	for _, item := range pkgExternalTypes {
		for _, typeRef := range item {
			if _, ok := uniquePkgTypeRefs[typeRef.PackageName]; !ok {
				uniquePkgTypeRefs[typeRef.PackageName] = make(map[string]bool)
			}
			uniquePkgTypeRefs[typeRef.PackageName][typeRef.TypeName] = true
		}
	}

	for childPkgName := range uniquePkgTypeRefs {
		childTypes := uniquePkgTypeRefs[childPkgName]
		childPkgPr := prsr{fs: pr.fs}
		childDefs, _ := childPkgPr.parseTypesInPackage(childPkgName, childTypes, false, true)
		mergeDefs(pkgDefs, childDefs)
	}

	return pkgDefs, pkgCRDSpecs
}

type SingleVersionOptions struct {
	// InputPackage is the path of the input package that contains source files.
	InputPackage string
	// Types is a list of target types.
	Types []string
	// Flatten contains if we use a flattened structure or a embedded structure.
	Flatten bool

	// fs is provided FS. We can use afero.NewMemFs() for testing.
	fs afero.Fs
}

type WriterOptions struct {
	// OutputPath is the path that the schema will be written to.
	OutputPath string
	// OutputFormat should be either json or yaml. Default to json
	OutputFormat string

	defs     v1beta1.JSONSchemaDefinitions
	crdSpecs crdSpecByKind
}

type SingleVersionGenerator struct {
	SingleVersionOptions
	WriterOptions

	outputCRD bool
}

type toplevelGeneratorOptions struct {
	group string
}

type prsr struct {
	generatorOptions *toplevelGeneratorOptions

	fs afero.Fs
}

func (op *SingleVersionGenerator) Generate() {
	if len(op.InputPackage) == 0 || len(op.OutputPath) == 0 {
		log.Panic("Both input path and output paths need to be set")
	}

	if op.fs == nil {
		op.fs = afero.NewOsFs()
	}

	if op.outputCRD {
		// if generating CRD, we should always embed schemas.
		op.Flatten = false
	}

	op.defs, op.crdSpecs = op.parse()

	op.write(op.outputCRD, op.Types)
}

func (pr *prsr) linkCRDSpec(defs v1beta1.JSONSchemaDefinitions, crdSpecs crdSpecByKind) crdSpecByKind {
	rtCRDSpecs := crdSpecByKind{}
	for gk := range crdSpecs {
		if pr.generatorOptions != nil {
			crdSpecs[gk].Group = pr.generatorOptions.group
			rtCRDSpecs[schema.GroupKind{Group: pr.generatorOptions.group, Kind: gk.Kind}] = crdSpecs[gk]
		} else {
			rtCRDSpecs[gk] = crdSpecs[gk]
		}

		if len(crdSpecs[gk].Versions) == 0 {
			log.Printf("no version for CRD %q", gk)
			continue
		}
		if len(crdSpecs[gk].Versions) > 1 {
			log.Fatalf("the number of versions in one package is more than 1")
		}
		def, ok := defs[gk.Kind]
		if !ok {
			log.Printf("can't get json shchema for %q", gk)
			continue
		}
		crdSpecs[gk].Versions[0].Schema = &v1beta1.CustomResourceValidation{
			OpenAPIV3Schema: &def,
		}
	}
	return rtCRDSpecs
}

func (op *SingleVersionOptions) parse() (v1beta1.JSONSchemaDefinitions, crdSpecByKind) {
	startingPointMap := make(map[string]bool)
	for i := range op.Types {
		startingPointMap[op.Types[i]] = true
	}
	pr := prsr{fs: op.fs}
	defs, crdSpecs := pr.parseTypesInPackage(op.InputPackage, startingPointMap, true, false)

	// flattenAllOf only flattens allOf tags
	flattenAllOf(defs)

	reachableTypes := getReachableTypes(startingPointMap, defs)
	for key := range defs {
		if _, exists := reachableTypes[key]; !exists {
			delete(defs, key)
		}
	}

	checkDefinitions(defs, startingPointMap)

	if !op.Flatten {
		defs = embedSchema(defs, startingPointMap)

		newDefs := v1beta1.JSONSchemaDefinitions{}
		for name := range startingPointMap {
			newDefs[name] = defs[name]
		}
		defs = newDefs
	}

	return defs, pr.linkCRDSpec(defs, crdSpecs)
}

func (op *WriterOptions) write(outputCRD bool, types []string) {
	var toSerilizeList []interface{}
	if outputCRD {
		for gk, spec := range op.crdSpecs {
			crd := &v1beta1.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apiextensions.k8s.io/v1beta1",
					Kind:       "CustomResourceDefinition",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   strings.ToLower(gk.Kind),
					Labels: map[string]string{"controller-tools.k8s.io": "1.0"},
				},
				Spec: *spec,
			}
			toSerilizeList = append(toSerilizeList, crd)
		}
	} else {
		schema := v1beta1.JSONSchemaProps{Definitions: op.defs}
		schema.Type = "object"
		schema.AnyOf = []v1beta1.JSONSchemaProps{}
		for _, typeName := range types {
			schema.AnyOf = append(schema.AnyOf, v1beta1.JSONSchemaProps{Ref: getDefLink(typeName)})
		}
		toSerilizeList = []interface{}{schema}
	}

	// TODO: create dir is not exist.
	out, err := os.Create(op.OutputPath)
	if err != nil {
		log.Panic(err)
	}

	for i := range toSerilizeList {
		switch strings.ToLower(op.OutputFormat) {
		// default to json
		case "json", "":
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			err = enc.Encode(toSerilizeList[i])
			if err2 := out.Close(); err == nil {
				err = err2
			}
			if err != nil {
				log.Panic(err)
			}
		case "yaml":
			m, err := yaml.Marshal(toSerilizeList[i])
			if err != nil {
				log.Panic(err)
			}
			err = ioutil.WriteFile(op.OutputPath, m, 0644)
			if err != nil {
				log.Panic(err)
			}
		}
	}
}
