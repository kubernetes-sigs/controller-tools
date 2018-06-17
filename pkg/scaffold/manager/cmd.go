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
	"path/filepath"
	"text/template"

	"log"

	flag "github.com/spf13/pflag"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
)

// Cmd scaffolds a manager.go to run Controllers
type Cmd struct {
	// OutputPath is the output file to write
	OutputPath string

	// Boilerplate is the boilerplate header to write
	Boilerplate string

	// Project is the project
	Project project.Project
}

// Name is the name of the template
func (Cmd) Name() string {
	return "manager-go"
}

// Path implements scaffold.Path.  Defaults to hack/boilerplate.go.txt
func (m *Cmd) Path() string {
	dir := filepath.Join("cmd", "manager", "main.go")
	if m.OutputPath != "" {
		dir = m.OutputPath
	}
	return dir
}

// SetBoilerplate implements scaffold.Boilerplate.
func (m *Cmd) SetBoilerplate(b string) {
	m.Boilerplate = b
}

// SetProject implements scaffold.Project.
func (m *Cmd) SetProject(p project.Project) {
	m.Project = p
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (m *Cmd) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		// Do nothing if the file exists
		return nil
	}
	temp, err := t.Parse(managerTemplate)
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

// ForFlags registers flags for Main fields and returns the Main
func ForFlags(_ *flag.FlagSet) *Cmd {
	return &Cmd{}
}

var managerTemplate = `{{ .Boilerplate }}

package main

import (
	"log"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"{{ .Project.Repo }}/pkg/apis"
	"{{ .Project.Repo }}/pkg/controller"
)

func main() {
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mrg, err := manager.New(cfg, manager.Options{})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Registering Components.")

	// Setup all Resources
	apis.AddToScheme(mrg.GetScheme())

	// Setup all Controllers
	if err := controller.AddToManager(mrg); err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting the Cmd.")

	// Start the Cmd
	log.Fatal(mrg.Start(signals.SetupSignalHandler()))
}
`
