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
	"go/build"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

// Project scaffolds the PROJECT file with project metadata
type Project struct {
	// OutputPath is the output file location - defaults to PROJECT
	OutputPath string `yaml:",omitempty"`

	// Version is the project version - defaults to "2"
	Version string `yaml:"version,omitempty"`

	// Domain is the domain associated with the project and used for API groups
	Domain string `yaml:"domain,omitempty"`

	// Repo is the go package name of the project root
	Repo string `yaml:"repo,omitempty"`
}

// Name is the name of the template
func (Project) Name() string {
	return "project"
}

// Path implements scaffold.Path.  Defaults to hack/boilerplate.go.txt
func (p *Project) Path() string {
	dir := filepath.Join("PROJECT")
	if p.OutputPath != "" {
		dir = p.OutputPath
	}
	return dir
}

func (Project) defaultRepo() (string, error) {
	// Assume the working dir is the root of the repo
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Strip the GOPATH from the working dir to get the go package of the repo
	gopath := os.Getenv("GOPATH")
	if len(gopath) == 0 {
		gopath = build.Default.GOPATH
	}
	goSrc := filepath.Join(gopath, "src")

	// Make sure the GOPATH is set and the working dir is under the GOPATH
	if !strings.HasPrefix(filepath.Dir(wd), goSrc) {
		return "", fmt.Errorf("kubebuilder must be run from the project root under $GOPATH/src/<package>. "+
			"\nCurrent GOPATH=%s.  \nCurrent directory=%s", gopath, wd)
	}

	// Prune the base path from the go package for the repo
	repo := strings.Replace(wd, fmt.Sprintf("%s%s", goSrc, string(filepath.Separator)), "", 1)

	// Make sure the prune did what it was supposed to
	if strings.Contains(repo, goSrc) {
		return "", fmt.Errorf("could not parse go package for repo: %s", repo)
	}
	return repo, err
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (p *Project) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		// Do nothing if the file exists
		return nil
	}

	if p.Repo == "" {
		r, err := p.defaultRepo()
		if err != nil {
			return err
		}
		p.Repo = r
	}

	out, err := yaml.Marshal(p)
	if err != nil {
		return err
	}

	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	_, err = w.Write(out)
	return err
}

// GetProject reads the project file and deserializes it into a Project
func GetProject(path string) (Project, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return Project{}, err
	}
	p := Project{}
	err = yaml.Unmarshal(in, &p)
	if err != nil {
		return Project{}, err
	}
	return p, nil
}

// ForFlags registers flags for Project fields and returns the Project
func ForFlags(f *flag.FlagSet) *Project {
	p := &Project{}
	f.StringVar(&p.Domain, "domain", "k8s.io", "domain for groups")
	f.StringVar(&p.Version, "project-version", "2", "project version")
	f.StringVar(&p.Repo, "repo", "", "name of the github repo.  "+
		"defaults to the go package of the current working directory.")
	return p
}

// Path returns the default location for the PROJECT file
func Path() string {
	return (&Project{}).Path()
}

// DieIfNoProject checks to make sure the command is run from a directory containing a project file.
func DieIfNoProject() {
	if _, err := os.Stat(Path()); os.IsNotExist(err) {
		log.Fatalf("Command must be run from a diretory containing %s", Path())
	}
}
