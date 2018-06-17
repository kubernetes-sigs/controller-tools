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

package scaffold

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
)

// Template implements a scaffoldable component
type Template interface {
	// Path returns the path of the file to write
	Path() string

	// Execute parses t and executes it to write to wr
	Execute(old []byte, t *template.Template, wr func() io.WriteCloser) error
}

// Validate allows a Template to validate
type Validate interface {
	// Validate returns true if the template has valid values
	Validate() error
}

// Name allows a Template to have a name
type Name interface {
	// Name returns the name of the template
	Name() string
}

// Project allows a Template to get the Project
type Project interface {
	// SetProject sets the project
	SetProject(project.Project)
}

// Boilerplate allows a Template to get the Boilerplate to put at the top of go files
type Boilerplate interface {
	// SetBoilerplate sets the boilerplate file content
	SetBoilerplate(string)
}

// BoilerplatePath allows a Template to get the path to the boilerplate file
type BoilerplatePath interface {
	// SetBoilerplatePath sets the path to the boilerplate file
	SetBoilerplatePath(string)
}

// Options are the options for executing scaffold templates
type Options struct {
	// BoilerplatePath is the path to the boilerplate file
	BoilerplatePath string

	// Path is the path to the project
	ProjectPath string
}

// Scaffold writes Templates to scaffold new files
type Scaffold struct {
	// BoilerplatePath is the path to the boilerplate file
	BoilerplatePath string

	// Boilerplate is the contents of the boilerplate file for code generation
	Boilerplate string

	BoilerplateOptional bool

	// Project is the project
	Project project.Project

	ProjectOptional bool
}

func (s *Scaffold) setFieldsAndValidate(t Template) error {
	// Set boilerplate on templates
	if b, ok := t.(BoilerplatePath); ok {
		b.SetBoilerplatePath(s.BoilerplatePath)
	}
	if b, ok := t.(Boilerplate); ok {
		b.SetBoilerplate(s.Boilerplate)
	}
	if b, ok := t.(Project); ok {
		b.SetProject(s.Project)
	}

	// Validate the template is ok
	if v, ok := t.(Validate); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scaffold) getBoilerplate(path string) (string, error) {
	return project.GetBoilerplate(path)
}

func (s *Scaffold) getProject(path string) (project.Project, error) {
	return project.GetProject(path)
}

func (s *Scaffold) defaultFunc(options *Options) error {
	if options.BoilerplatePath == "" {
		options.BoilerplatePath = project.BoilerplatePath()
	}
	if options.ProjectPath == "" {
		options.ProjectPath = project.Path()
	}

	s.BoilerplatePath = options.BoilerplatePath

	var err error
	s.Boilerplate, err = s.getBoilerplate(options.BoilerplatePath)
	if !s.BoilerplateOptional && err != nil {
		return err
	}

	s.Project, err = s.getProject(options.ProjectPath)
	if !s.ProjectOptional && err != nil {
		return err
	}
	return nil
}

// Execute executes the Templates
func (s *Scaffold) Execute(options Options, elem ...Template) error {
	if err := s.defaultFunc(&options); err != nil {
		return err
	}
	for _, t := range elem {
		err := s.setFieldsAndValidate(t)
		if err != nil {
			return err
		}

		b, _ := ioutil.ReadFile(t.Path())

		// Write the template
		temp := newTemplate(t)
		err = t.Execute(b, temp, func() io.WriteCloser { return newWriteCloser(t.Path()) })
		if err != nil {
			return err
		}
	}
	return nil
}

func newWriteCloser(path string) io.WriteCloser {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		log.Fatal(err)
	}

	fi, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}

	return fi
}

func newTemplate(t Template) *template.Template {
	name := ""
	if n, ok := t.(Name); ok {
		name = n.Name()
	}
	return template.New(name).Funcs(template.FuncMap{
		"title": strings.Title,
		"lower": strings.ToLower,
	})
}
