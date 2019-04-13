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
	"go/build"
	"log"

	"github.com/spf13/afero"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

type listFilesFn func(pkgPath string) (dir string, files []string, e error)

var defaultListFiles = func(pkgPath string) (string, []string, error) {
	pkg, err := build.Import(pkgPath, "", 0)
	return pkg.Dir, pkg.GoFiles, err
}

type SingleVersionOptions struct {
	// InputPackage is the path of the input package that contains source files.
	InputPackage string
	// Types is a list of target types.
	Types []string
	// Flatten contains if we use a flattened structure or a embedded structure.
	Flatten bool

	// fs is provided FS. We can use afero.NewMemFs() for testing.
	fs          afero.Fs
	listFilesFn listFilesFn
}

type SingleVersionGenerator struct {
	SingleVersionOptions
	WriterOptions

	outputCRD bool
}

func (op *SingleVersionOptions) parse() (v1beta1.JSONSchemaDefinitions, crdSpecByKind) {
	startingPointMap := make(map[string]bool)
	for i := range op.Types {
		startingPointMap[op.Types[i]] = true
	}
	pr := prsr{
		listFilesFn: op.listFilesFn,
		fs:          op.fs,
	}
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

	pr.linkCRDSpec(defs, crdSpecs)
	return defs, crdSpecs
}

func (op *SingleVersionGenerator) Generate() {
	if len(op.InputPackage) == 0 || len(op.OutputPath) == 0 {
		log.Panic("Both input path and output paths need to be set")
	}

	if op.fs == nil {
		op.fs = afero.NewOsFs()
	}
	if op.listFilesFn == nil {
		op.listFilesFn = defaultListFiles
	}

	if op.outputCRD {
		// if generating CRD, we should always embed schemas.
		op.Flatten = false
	}

	op.defs, op.crdSpecs = op.parse()

	op.write(op.outputCRD, op.Types)
}

func (pr *prsr) linkCRDSpec(defs v1beta1.JSONSchemaDefinitions, crdSpecs crdSpecByKind) {
	for gk := range crdSpecs {
		if len(crdSpecs[gk].Versions) == 0 {
			log.Printf("no version for CRD %q", gk)
			continue
		}
		if len(crdSpecs[gk].Versions) > 1 {
			log.Fatalf("the number of versions in one package should not be more than 1")
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
}
