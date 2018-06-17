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

package manager

import (
	"io"
	"log"
	"path/filepath"
	"text/template"
)

// Controller scaffolds a controller.go to add Controllers to a manager.Cmd
type Controller struct {
	// OutputPath is the output file to write
	OutputPath string

	// Boilerplate is the boilerplate header to write
	Boilerplate string
}

// Name is the name of the template
func (Controller) Name() string {
	return "manager-controller-go"
}

// Path implements scaffold.Path.  Defaults to pkg/controller/controller.go
func (m *Controller) Path() string {
	dir := filepath.Join("pkg", "controller", "controller.go")
	if m.OutputPath != "" {
		dir = m.OutputPath
	}
	return dir
}

// SetBoilerplate implements scaffold.Boilerplate.
func (m *Controller) SetBoilerplate(b string) {
	m.Boilerplate = b
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (m *Controller) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		// Do nothing if the file exists
		return nil
	}
	temp, err := t.Parse(controllerTemplate)
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

var controllerTemplate = `{{ .Boilerplate }}

package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Adds registers functions to add Controllers to the Cmd
var AddToManagerFuncs []func(manager.Manager) error

// Add adds all Controllers to the Cmd
func AddToManager(m manager.Manager) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}
`
