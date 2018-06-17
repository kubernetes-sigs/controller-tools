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
	"log"
	"regexp"
	"strings"

	"github.com/markbates/inflect"
	flag "github.com/spf13/pflag"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
)

// Resource contains the information required to scaffold files for a resource.
type Resource struct {
	// Boilerplate is the contents of the go boilerplate header to write at the top of go files
	Boilerplate string

	// Namespaced is true if the resource is namespaced
	Namespaced bool

	// Group is the API Group.  Does not contain the domain.
	Group string

	// Domain is the domain of the API Group.
	Project project.Project

	// Version is the API version - e.g. v1beta1
	Version string

	// Kind is the API Kind.
	Kind string

	// Resource is the API Resource.
	Resource string

	// ShortNames is the list of resource shortnames.
	ShortNames []string
}

// SetBoilerplate implements scaffold.Boilerplate.
func (r *Resource) SetBoilerplate(b string) {
	r.Boilerplate = b
}

// SetProject injects the project
func (r *Resource) SetProject(p project.Project) {
	r.Project = p
}

// Validate checks the Resource values to make sure they are valid.
func (r *Resource) Validate() error {
	if len(r.Project.Domain) == 0 {
		return fmt.Errorf("Must specify a Project domain")
	}
	if len(r.Project.Repo) == 0 {
		return fmt.Errorf("Must specify a Project repo")
	}
	if len(r.Group) == 0 {
		return fmt.Errorf("Must specify --group")
	}
	if len(r.Version) == 0 {
		return fmt.Errorf("Must specify --version")
	}
	if len(r.Kind) == 0 {
		log.Fatal("Must specify --kind")
	}

	rs := inflect.NewDefaultRuleset()
	if len(r.Resource) == 0 {
		r.Resource = rs.Pluralize(strings.ToLower(r.Kind))
	}

	groupMatch := regexp.MustCompile("^[a-z]+$")
	if !groupMatch.MatchString(r.Group) {
		return fmt.Errorf("--group must match regex ^[a-z]+$ but was (%s)", r.Group)
	}
	versionMatch := regexp.MustCompile("^v\\d+(alpha\\d+|beta\\d+)*$")
	if !versionMatch.MatchString(r.Version) {
		return fmt.Errorf(
			"--version has bad format. must match ^v\\d+(alpha\\d+|beta\\d+)*$.  "+
				"e.g. v1alpha1,v1beta1,v1 but was (%s)", r.Version)
	}

	kindMatch := regexp.MustCompile("^[A-Z]+[A-Za-z0-9]*$")
	if !kindMatch.MatchString(r.Kind) {
		return fmt.Errorf("--kind must match regex ^[A-Z]+[A-Za-z0-9]*$ but was (%s)", r.Kind)
	}

	return nil
}

// ForFlags registers flags for Resource fields and returns the Resource
func ForFlags(f *flag.FlagSet) *Resource {
	r := &Resource{}
	f.StringVar(&r.Kind, "kind", "", "resource Kind")
	f.StringVar(&r.Group, "group", "", "resource Group")
	f.StringVar(&r.Version, "version", "", "resource Version")
	f.BoolVar(&r.Namespaced, "namespaced", true, "true if the resource is namespaced")
	return r
}
