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

package controller

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"log"

	"sigs.k8s.io/controller-tools/pkg/scaffold"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
	"sigs.k8s.io/controller-tools/pkg/scaffold/resource"
)

var _ scaffold.Name = &AddController{}
var _ scaffold.Template = &AddController{}

// AddController scaffolds adds a new Controller.
type AddController struct {
	// OutputPath is the output file to write
	OutputPath string

	// Resource is a resource in the API group
	*resource.Resource

	// Project is the project
	Project project.Project
}

// Name implements scaffold.Name
func (AddController) Name() string {
	return "controller-controller-go"
}

// Path implements scaffold.Path.  Defaults to pkg/controller/add_kind.go
func (m AddController) Path() string {
	dir := filepath.Join("pkg", "controller", fmt.Sprintf(
		"add_%s.go", strings.ToLower(m.Kind)))
	if m.OutputPath != "" {
		dir = m.OutputPath
	}
	return dir
}

// SetBoilerplate implements scaffold.Boilerplate.
func (m *AddController) SetBoilerplate(b string) {
	m.Boilerplate = b
}

// SetProject injects the Project
func (m *AddController) SetProject(p project.Project) {
	m.Project = p
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (m AddController) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
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
	return temp.Execute(w, m)
}

var managerInitTemplate = `{{ .Boilerplate }}

package controller

import (
	
	"{{ .Project.Repo }}/pkg/controller/{{ lower .Kind }}"
)

func init() {
	// Create the Controller and add it to the Manager.
	AddToManagerFuncs = append(AddToManagerFuncs, {{ lower .Kind }}.Add)
}
`
