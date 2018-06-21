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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-tools/pkg/scaffold/input"
)

var _ input.File = &Project{}

// Project scaffolds the PROJECT file with project metadata
type Project struct {
	// Path is the output file location - defaults to PROJECT
	Path string `yaml:",omitempty"`

	// Version is the project version - defaults to "2"
	Version string `yaml:"version,omitempty"`

	// Domain is the domain associated with the project and used for API groups
	Domain string `yaml:"domain,omitempty"`

	// Repo is the go package name of the project root
	Repo string `yaml:"repo,omitempty"`
}

// GetInput implements input.File
func (c *Project) GetInput() (input.Input, error) {
	if c.Path == "" {
		c.Path = "PROJECT"
	}
	if c.Repo == "" {
		r, err := c.defaultRepo()
		if err != nil {
			return input.Input{}, err
		}
		c.Repo = r
	}

	out, err := yaml.Marshal(c)
	if err != nil {
		return input.Input{}, err
	}

	return input.Input{
		Path:         c.Path,
		TemplateBody: string(out),
	}, nil
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
	i, _ := (&Project{}).GetInput()
	return i.Path
}

// DieIfNoProject checks to make sure the command is run from a directory containing a project file.
func DieIfNoProject() {
	if _, err := os.Stat(Path()); os.IsNotExist(err) {
		log.Fatalf("Command must be run from a diretory containing %s", Path())
	}
}
