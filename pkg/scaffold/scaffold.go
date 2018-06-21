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
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"fmt"

	"log"

	"sigs.k8s.io/controller-tools/pkg/scaffold/input"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
)

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

func (s *Scaffold) setFieldsAndValidate(t input.File) error {
	// Set boilerplate on templates
	if b, ok := t.(input.BoilerplatePath); ok {
		b.SetBoilerplatePath(s.BoilerplatePath)
	}
	if b, ok := t.(input.Boilerplate); ok {
		b.SetBoilerplate(s.Boilerplate)
	}
	if b, ok := t.(input.Domain); ok {
		b.SetDomain(s.Project.Domain)
	}
	if b, ok := t.(input.Version); ok {
		b.SetVersion(s.Project.Version)
	}
	if b, ok := t.(input.Repo); ok {
		b.SetRepo(s.Project.Repo)
	}

	// Validate the template is ok
	if v, ok := t.(input.Validate); ok {
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

func (s *Scaffold) defaultOptions(options *input.Options) error {
	// Use the default Boilerplate path if unset
	if options.BoilerplatePath == "" {
		options.BoilerplatePath = project.BoilerplatePath()
	}

	// Use the default Project path if unset
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

// Execute executes scaffolding the Files
func (s *Scaffold) Execute(options input.Options, files ...input.File) error {
	if err := s.defaultOptions(&options); err != nil {
		return err
	}
	for _, f := range files {
		if err := s.doFile(f); err != nil {
			return err
		}
	}
	return nil
}

// doFile scaffolds a single file
func (s *Scaffold) doFile(e input.File) error {
	// Set common fields
	err := s.setFieldsAndValidate(e)
	if err != nil {
		return err
	}

	// Get the template input params
	i, err := e.GetInput()
	if err != nil {
		return err
	}

	// Check if the file to write already exists
	if _, err := os.Stat(i.Path); err == nil {
		switch i.IfExistsAction {
		case input.Overwrite:
		case input.Skip:
			return nil
		case input.Error:
			return fmt.Errorf("%s already exists", i.Path)
		}
	}

	if err := s.doTemplate(i, e); err != nil {
		return err
	}
	return nil
}

// doTemplate executes the template for a file using the input
func (*Scaffold) doTemplate(i input.Input, e input.File) error {
	temp, err := newTemplate(e).Parse(i.TemplateBody)
	if err != nil {
		return err
	}
	f, err := newWriteCloser(i.Path)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return temp.Execute(f, e)
}

// newWriteCloser returns a WriteCloser to write scaffold to
func newWriteCloser(path string) (io.WriteCloser, error) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return nil, err
	}

	fi, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}

	return fi, nil
}

// newTemplate a new template with common functions
func newTemplate(t input.File) *template.Template {
	return template.New(fmt.Sprintf("%T", t)).Funcs(template.FuncMap{
		"title": strings.Title,
		"lower": strings.ToLower,
	})
}
