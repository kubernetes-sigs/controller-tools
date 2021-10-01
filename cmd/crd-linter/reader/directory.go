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

package reader

import (
	"bytes"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type accumulator struct {
	crds []CRDToEvaluate
}

type CRDToEvaluate struct {
	// OriginalFilename is the name of the file that the CRD was read from
	OriginalFilename string

	// OriginalAPIVersion is the API version of the resource as it is on disk
	OriginalGroupVersionKind schema.GroupVersionKind

	// CustomResourceDefinition is the apiextensions.k8s.io/v1 representation of the CRD
	CustomResourceDefinition *v1.CustomResourceDefinition
}

func DiscoverAllCRDs(path string) ([]CRDToEvaluate, error) {
	a := &accumulator{}
	if err := filepath.WalkDir(path, a.processDirectory); err != nil {
		return nil, err
	}
	return a.list(), nil
}

func (a *accumulator) processDirectory(path string, d fs.DirEntry, err error) error {
	if err != nil {
		log.Printf("Error reading path %q: %v", path, err)
		return nil
	}

	if d.IsDir() {
		return nil
	}

	if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
		log.Printf("Skipping file with non-yaml extension %q: %s", filepath.Ext(path), path)
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	dataChunks := bytes.Split(data, []byte("---\n"))
	for _, data := range dataChunks {
		obj, gvk, err := DecodeCustomResourceDefinition(data, crdV1GVK.GroupVersion())
		if err != nil {
			log.Printf("Failed to decode file %q: %v", path, err)
			return nil
		}

		a.crds = append(a.crds, CRDToEvaluate{
			OriginalFilename:         path,
			OriginalGroupVersionKind: *gvk,
			CustomResourceDefinition: obj.(*v1.CustomResourceDefinition),
		})
	}

	return nil
}

func (a *accumulator) list() []CRDToEvaluate {
	return a.crds
}
