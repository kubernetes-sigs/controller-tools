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

package resource

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"log"

	"sigs.k8s.io/controller-tools/pkg/scaffold"
)

var _ scaffold.Template = &TypesTest{}
var _ scaffold.Name = &TypesTest{}

// TypesTest scaffolds the types_test.go file for testing APIsGo
type TypesTest struct {
	// Resource is the resource to scaffold the types_test.go file for
	*Resource
	// OutputPath is the output file to write
	OutputPath string
}

// Name implements scaffold.Name
func (TypesTest) Name() string {
	return "types-test-go"
}

// Path implements scaffold.Path.  Defaults to pkg/apis/group/version/kind_types_test.go
func (t TypesTest) Path() string {
	dir := filepath.Join("pkg", "apis", t.Group, t.Version)
	if t.OutputPath != "" {
		dir = t.OutputPath
	}

	return filepath.Join(dir, fmt.Sprintf("%s_types_test.go", strings.ToLower(t.Kind)))
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (t TypesTest) Execute(b []byte, temp *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		return fmt.Errorf("%s already exists", t.Path())
	}

	temp, err := temp.Parse(typesTestTemplate)
	if err != nil {
		return err
	}

	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return temp.Execute(w, t)
}

var typesTestTemplate = `{{ .Boilerplate }}

package {{.Version}}

`
