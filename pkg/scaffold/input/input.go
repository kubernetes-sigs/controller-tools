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

package input

// IfExistsAction determines what to do if the scaffold file already exists
type IfExistsAction int

const (
	// Skip skips the file and moves to the next one
	Skip IfExistsAction = iota

	// Error returns an error and stops processing
	Error

	// Overwrite truncates and overwrites the existing file
	Overwrite
)

// Input is the input for scaffoldig a file
type Input struct {
	// Path is the file to write
	Path string

	// IfExistsAction determines what to do if the file exists
	IfExistsAction IfExistsAction

	// TemplateBody is the template body to execute
	TemplateBody string

	// Boilerplate is the contents of a Boilerplate go header file
	Boilerplate string

	// BoilerplatePath is the path to a Boilerplate go header file
	BoilerplatePath string

	// Version is the project version
	Version string

	// Domain is the domain for the APIs
	Domain string

	// Repo is the go project package
	Repo string
}

// Domain allows a domain to be set on an object
type Domain interface {
	// SetDomain sets the domain
	SetDomain(string)
}

// SetDomain sets the domain
func (i *Input) SetDomain(d string) {
	i.Domain = d
}

// Repo allows a repo to be set on an object
type Repo interface {
	// SetRepo sets the repo
	SetRepo(string)
}

// SetRepo sets the repo
func (i *Input) SetRepo(r string) {
	i.Repo = r
}

// Boilerplate allows boilerplate text to be set on an object
type Boilerplate interface {
	// SetBoilerplate sets the boilerplate text
	SetBoilerplate(string)
}

// SetBoilerplate sets the boilerplate text
func (i *Input) SetBoilerplate(b string) {
	i.Boilerplate = b
}

// BoilerplatePath allows boilerplate file path to be set on an object
type BoilerplatePath interface {
	// SetBoilerplatePath sets the boilerplate file path
	SetBoilerplatePath(string)
}

// SetBoilerplatePath sets the boilerplate file path
func (i *Input) SetBoilerplatePath(bp string) {
	i.BoilerplatePath = bp
}

// Version allows the project version to be set on an object
type Version interface {
	// SetVersion sets the project version
	SetVersion(string)
}

// SetVersion sets the project version
func (i *Input) SetVersion(v string) {
	i.Version = v
}

// File is a scaffoldable file
type File interface {
	// GetInput returns the Input for creating a scaffold file
	GetInput() (Input, error)
}

// Validate validates input
type Validate interface {
	// Validate returns true if the template has valid values
	Validate() error
}

// Options are the options for executing scaffold templates
type Options struct {
	// BoilerplatePath is the path to the boilerplate file
	BoilerplatePath string

	// Path is the path to the project
	ProjectPath string
}
