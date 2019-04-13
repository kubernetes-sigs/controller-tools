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
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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

	if !skipCRD {
		// process top-level (not tied to a struct field) markers.
		// e.g. group name marker +groupName=<group-name>
		pr.processTopLevelMarkers(node.Comments)
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
				pr.generatorOptions = &pkglevelGeneratorOptions{group: group}
			} else {
				pr.generatorOptions.group = group
			}
		}

		if cs.hasTag("versionName") {
			version := cs.getTag("versionName", "=")
			if len(version) == 0 {
				log.Fatalf("can't use an empty name for the +versionName marker")
			}
			if pr.generatorOptions != nil && len(pr.generatorOptions.version) > 0 && version != pr.generatorOptions.version {
				log.Fatalf("can't have different version names %q and %q one package", pr.generatorOptions.version, version)
			}
			if pr.generatorOptions == nil {
				pr.generatorOptions = &pkglevelGeneratorOptions{version: version}
			} else {
				pr.generatorOptions.version = version
			}
		}
	}
}

func (pr *prsr) parseTypesInPackage(pkgName string, referencedTypes map[string]bool, rootPackage, skipCRD bool) (
	v1beta1.JSONSchemaDefinitions, crdSpecByKind) {
	pkgDefs := make(v1beta1.JSONSchemaDefinitions)
	pkgExternalTypes := make(ExternalReferences)
	pkgCRDSpecs := make(crdSpecByKind)

	pkgDir, listOfFiles, err := pr.listFilesFn(pkgName)
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

	return pkgDefs, pr.fanoutPkgLevelOptions(pkgCRDSpecs)
}

type pkglevelGeneratorOptions struct {
	group   string
	version string
}

type prsr struct {
	generatorOptions *pkglevelGeneratorOptions

	listFilesFn listFilesFn
	fs          afero.Fs
}

func (pr *prsr) fanoutPkgLevelOptions(crdSpecs crdSpecByKind) crdSpecByKind {
	rtCRDSpecs := crdSpecByKind{}
	for gk := range crdSpecs {
		if pr.generatorOptions != nil {
			crdSpecs[gk].Group = pr.generatorOptions.group
			if len(crdSpecs[gk].Versions) == 1 {
				crdSpecs[gk].Versions[0].Name = pr.generatorOptions.version
			}
			rtCRDSpecs[schema.GroupKind{Group: pr.generatorOptions.group, Kind: gk.Kind}] = crdSpecs[gk]
		} else {
			rtCRDSpecs[gk] = crdSpecs[gk]
		}
	}
	return rtCRDSpecs
}
