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
	"log"
	"path/filepath"
	"text/template"

	"sigs.k8s.io/controller-tools/pkg/scaffold"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
)

var _ scaffold.Name = &AddResource{}
var _ scaffold.Template = &AddResource{}

// AddResource scaffolds the manager init code.
type AddResource struct {
	// OutputPath is the output file to write
	OutputPath string

	// Resource is a resource in the API group
	*Resource

	// Project is the project
	Project project.Project
}

// Name implements scaffold.Name
func (AddResource) Name() string {
	return "pkg-resource-go"
}

// Path implements scaffold.Path.  Defaults to cmd/manager/setup/group_version_kind_init
func (a AddResource) Path() string {
	dir := filepath.Join("pkg", "apis", fmt.Sprintf(
		"add_%s_%s.go", a.Group, a.Version))
	if a.OutputPath != "" {
		dir = a.OutputPath
	}
	return dir
}

// SetBoilerplate implements scaffold.Boilerplate.
func (a *AddResource) SetBoilerplate(b string) {
	a.Boilerplate = b
}

// SetProject injects the Project
func (a *AddResource) SetProject(p project.Project) {
	a.Project = p
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (a AddResource) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	// Already exists, do nothing
	if len(b) > 0 {
		return nil
	}

	temp, err := t.Parse(managerInitTemplate)
	if err != nil {
		return err
	}
	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return temp.Execute(w, a)
}

var managerInitTemplate = `{{ .Boilerplate }}

package apis

import (
	"{{ .Project.Repo }}/pkg/apis/{{ .Group }}/{{ .Version }}"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,  v1beta1.AddToScheme)
}
`
