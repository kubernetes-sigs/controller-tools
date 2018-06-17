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

package project

import (
	"io"
	"path/filepath"
	"text/template"

	"log"

	flag "github.com/spf13/pflag"
)

// Main scaffolds a simple main file
type Main struct {
	// OutputPath is the output file location
	OutputPath string

	// Boilerplate is the boilerplate header to use
	Boilerplate string
}

// Name is the name of the template
func (Main) Name() string {
	return "doc-go"
}

// Path implements scaffold.Path.  Defaults to hack/boilerplate.go.txt
func (d *Main) Path() string {
	dir := filepath.Join("main.go")
	if d.OutputPath != "" {
		dir = d.OutputPath
	}
	return dir
}

// SetBoilerplate implements scaffold.Boilerplate.
func (d *Main) SetBoilerplate(b string) {
	d.Boilerplate = b
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (d *Main) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		// Do nothing if the file exists
		return nil
	}
	temp, err := t.Parse(mainTemplate)
	if err != nil {
		return err
	}

	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return temp.Execute(w, d)
}

var mainTemplate = `{{.Boilerplate}}

package main

func main() {}
`

// MainForFlags registers flags for Main fields and returns the Main
func MainForFlags(_ *flag.FlagSet) *Main {
	return &Main{}
}
