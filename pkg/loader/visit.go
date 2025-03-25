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

package loader

import (
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
)

// TypeCallback is a callback called for each raw AST (gendecl, typespec) combo.
type TypeCallback func(file *ast.File, decl *ast.GenDecl, spec *ast.TypeSpec)

// A ConstCallback is used to iterate over constant definitions. It provides
// the value spec only since it is the only thing needed yet.
type ConstCallback func(spec *ast.ValueSpec)

// EachType calls the given callback for each (gendecl, typespec) combo in the
// given package.  Generally, using markers.EachType is better when working
// with marker data, and has a more convinient representation.
func EachType(pkg *Package, cb TypeCallback) {
	visitor := &typeVisitor{
		callback: cb,
	}
	pkg.NeedSyntax()
	for _, file := range pkg.Syntax {
		visitor.file = file
		ast.Walk(visitor, file)
	}
}

// EachConstDecl calls the given callback for each
// *ast.ValueSpec associated with constant declarations.
func EachConstDecl(pkg *Package, cb ConstCallback) {
	pkg.NeedSyntax()
	for _, file := range pkg.Syntax {
		ast.Walk(cb, file)
	}
}

// Visit visits all top-level constant declarations.
func (v ConstCallback) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return v
	}
	var typedNode, ok = node.(*ast.GenDecl)
	if !ok {
		return v
	}
	if typedNode.Tok == token.CONST {
		for i := range typedNode.Specs {
			v(typedNode.Specs[i].(*ast.ValueSpec))
		}
	}
	return nil
}

// typeVisitor visits all TypeSpecs, calling the given callback for each.
type typeVisitor struct {
	callback TypeCallback
	decl     *ast.GenDecl
	file     *ast.File
}

// Visit visits all TypeSpecs.
func (v *typeVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		v.decl = nil
		return v
	}

	switch typedNode := node.(type) {
	case *ast.File:
		v.file = typedNode
		return v
	case *ast.GenDecl:
		v.decl = typedNode
		return v
	case *ast.TypeSpec:
		v.callback(v.file, v.decl, typedNode)
		return nil // don't recurse
	default:
		return nil
	}
}

// ParseAstTag parses the given raw tag literal into a reflect.StructTag.
func ParseAstTag(tag *ast.BasicLit) reflect.StructTag {
	if tag == nil {
		return reflect.StructTag("")
	}
	tagStr, err := strconv.Unquote(tag.Value)
	if err != nil {
		return reflect.StructTag("")
	}
	return reflect.StructTag(tagStr)
}
