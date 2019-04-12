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
	"go/build"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/spf13/afero"
)

type MultiVersionOptions struct {
	// InputPackage is the path of the input package that contains source files.
	InputPackage string
	// Types is a list of target types.
	Types []string

	// fs is provided FS. We can use afero.NewMemFs() for testing.
	fs afero.Fs
}

type MultiVersionGenerator struct {
	MultiVersionOptions
	WriterOptions
}

func (op *MultiVersionGenerator) Generate() {
	if len(op.InputPackage) == 0 || len(op.OutputPath) == 0 {
		log.Panic("Both input path and output paths need to be set")
	}

	if op.fs == nil {
		op.fs = afero.NewOsFs()
	}

	op.crdSpecs = op.parse()

	op.write(true, op.Types)
}

func listDirs(path string) ([]string, error) {
	pkg, err := build.Import(path, "", 0)
	if err != nil {
		return nil, err
	}
	infos, err := ioutil.ReadDir(pkg.Dir)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, info := range infos {
		if info.IsDir() {
			dirs = append(dirs, info.Name())
		}
	}
	return dirs, nil
}

func (op *MultiVersionOptions) parse() crdSpecByKind {
	startingPointMap := make(map[string]bool)
	for i := range op.Types {
		startingPointMap[op.Types[i]] = true
	}

	dirs, err := listDirs(op.InputPackage)
	if err != nil {
		panic(err)
	}
	fmt.Println(dirs)

	crdSpecs := crdSpecByKind{}
	for _, dir := range dirs {
		singleVer := SingleVersionOptions{
			Types:        op.Types,
			InputPackage: filepath.Join(op.InputPackage, dir),
			Flatten:      false,
			fs:           op.fs,
		}
		_, crdSingleVersionSpecs := singleVer.parse()
		// merge crd versions
		err = mergeCRDVersions(crdSpecs, crdSingleVersionSpecs)
		if err != nil {
			panic(err)
		}
	}

	return crdSpecs
}
