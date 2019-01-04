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

package annotation

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// Annotation defines a generic spec of annotations
// The schema is [header]:[module]:[submodule]:[key-value elements], submodule could be optional and multiple
type Annotation interface {
	Header(string)
	Module(*Module)
	HasModule(string) bool
	GetModule(string) *Module
	Parse(string) error
}

type defaultAnnotation struct {
	Headers   sets.String
	Modules   sets.String
	ModuleMap map[string]*Module
}

func (a *defaultAnnotation) Header(header string) {
	a.Headers.Insert(header)
}

func (a *defaultAnnotation) Module(m *Module) {
	a.Modules.Insert(m.Name)
	a.ModuleMap[m.Name] = m
}

func (a *defaultAnnotation) HasModule(name string) bool {
	return a.Modules.Has(name)
}

func (a *defaultAnnotation) GetModule(name string) *Module {
	for _, m := range a.ModuleMap {
		if m.Name == name {
			return m
		}
	}
	return nil
}

// Parse takes single line comment and validates each token.
func (a *defaultAnnotation) Parse(comments string) error {
	for _, comment := range strings.Split(comments, "\n") {
		comment = strings.TrimSpace(comment)
		for k := range a.Headers.Union(a.Modules) {
			if !strings.HasPrefix(comment, prefixName(k)) {
				continue
			}
			// parsing sigle whole line of comment into tokens split by comma (1st level delimiter)
			// This requires all key-values of same module/submodule should reside in the same comment line
			tokens := strings.Split(strings.TrimPrefix(comment, "+"), ":")
			if err := a.parseTokens(tokens); err != nil {
				return err
			}
		}
	}
	return nil
}

// Complete process annotaion string into Tokens
func (a *defaultAnnotation) parseTokens(tokens []string) error {
	if a.Headers.Has(tokens[0]) {
		// competitable for annotations without header starting with "+[module]"
		tokens = tokens[1:]
	}
	if a.Modules.Has(tokens[0]) {
		return a.GetModule(tokens[0]).parseModule(tokens)
	}
	return fmt.Errorf("annotation %+v format error", tokens)
}

// Module defines functional feature for annotation. Header may contain multiple modules,
// single module may contain submodules. Do() function defines what this module can do
type Module struct {
	Name       string
	Meta       interface{}
	SubModules map[string]*Module
	Do         func(string) error
}

// HasSubModule verify if given token string is a valid subresource
func (m *Module) HasSubModule(name string) bool {
	for _, v := range m.SubModules {
		if v.Name == name {
			return true
		}
	}
	return false
}

func (m *Module) parseModule(tokens []string) error {
	if len(tokens) == 1 {
		return m.Do(tokens[0])
	}
	// [module]:[submodule]:[element-values]
	if len(tokens) > 2 {
		s := tokens[1]
		if !m.HasSubModule(s) {
			return fmt.Errorf("annotation (%s) format error, has incorrect submodule %s", tokens, s)
		}
		return m.SubModules[s].parseModule(tokens[1:])
	}
	return m.Do(tokens[1])
}

// Build returns initialized default annotation
func Build() Annotation {
	return &defaultAnnotation{
		Headers:   sets.NewString(),
		Modules:   sets.NewString(),
		ModuleMap: map[string]*Module{},
	}
}
