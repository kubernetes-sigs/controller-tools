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
	"io"
	"log"
	"path/filepath"
	"text/template"

	"sigs.k8s.io/controller-tools/pkg/scaffold"
)

var _ scaffold.Name = &Group{}
var _ scaffold.Template = &Group{}

// Group scaffolds the group.go
type Group struct {
	// OutputPath is the output file to write
	OutputPath string

	// Resource is a resource in the API group
	*Resource
}

// Name implements scaffold.Name
func (Group) Name() string {
	return "group-go"
}

// Path implements scaffold.Path.  Defaults to pkg/apis/group/group.go
func (g Group) Path() string {
	dir := filepath.Join("pkg", "apis", g.Group, "group.go")
	if g.OutputPath != "" {
		dir = g.OutputPath
	}
	return dir
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (g Group) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		return nil
	}

	temp, err := t.Parse(groupTemplate)
	if err != nil {
		return err
	}

	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return temp.Execute(w, g)
}

var groupTemplate = `{{ .Boilerplate }}

// Package {{ .Group }} contains {{ .Group }} API versions
package {{ .Group }}
`
