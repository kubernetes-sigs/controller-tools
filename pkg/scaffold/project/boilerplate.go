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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"log"

	flag "github.com/spf13/pflag"
)

// Boilerplate scaffolds a boilerplate header file.
type Boilerplate struct {
	// OutputPath is the output location of the file to write
	OutputPath string

	// License is the License type to write
	License string

	// Owner is the copyright owner - e.g. "The Kubernetes Authors"
	Owner string

	// Year is the copyright year
	Year string

	// Boilerplate is the boilerplate to write.  Defaults to License + Owner + Year
	Boilerplate []byte
}

// Name is the name of the boilerplate template
func (p Boilerplate) Name() string {
	switch p.License {
	case "", "apache2":
		return "boilerplate-apache"
	case "none":
		return "boilerplate-none"
	default:
		return "boilerplate-unspecified"
	}
}

// Path implements scaffold.Path.  Defaults to hack/boilerplate.go.txt
func (p *Boilerplate) Path() string {
	dir := filepath.Join("hack", "boilerplate.go.txt")
	if p.OutputPath != "" {
		dir = p.OutputPath
	}
	return dir
}

func (p *Boilerplate) write(t *template.Template, wr func() io.WriteCloser) error {
	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return t.Execute(w, p)
}

func (p *Boilerplate) writeBoilerplate(wr func() io.WriteCloser) error {
	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	_, err := w.Write([]byte(p.Boilerplate))
	return err
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (p *Boilerplate) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		// Do nothing if the file exists
		return nil
	}

	if p.Year == "" {
		p.Year = fmt.Sprintf("%v", time.Now().Year())
	}

	err := os.MkdirAll(filepath.Dir(p.Path()), 0700)
	if err != nil {
		return err
	}

	if len(p.Boilerplate) <= 0 {
		switch p.License {
		case "", "apache2":
			t, err = t.Parse(apache)
			if err != nil {
				return err
			}
			return p.write(t, wr)
		case "none":
			t, err = t.Parse(none)
			if err != nil {
				return err
			}
			return p.write(t, wr)
		default:
			return fmt.Errorf("unrecognized LICENSE %s", p.License)
		}
	}

	return p.writeBoilerplate(wr)
}

var apache = `/*
{{ if .Owner }}Copyright {{ .Year }} {{ .Owner }}.
{{ end }}
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/`

var none = `/*
{{ if .Owner }}Copyright {{ .Year }} {{ .Owner }} {{ end }}.
*/`

// BoilerplateForFlags registers flags for Boilerplate fields and returns the Boilerplate
func BoilerplateForFlags(f *flag.FlagSet) *Boilerplate {
	b := &Boilerplate{}
	f.StringVar(&b.OutputPath, "path", "", "domain for groups")
	f.StringVar(&b.License, "license", "apache2",
		"license to use to boilerplate.  Maybe one of apache2,none")
	f.StringVar(&b.Owner, "owner", "",
		"Owner to add to the copyright")
	return b
}

// GetBoilerplate reads the boilerplate file
func GetBoilerplate(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	return string(b), err
}

// BoilerplatePath returns the default path to the boilerplate file
func BoilerplatePath() string {
	return (&Boilerplate{}).Path()
}
