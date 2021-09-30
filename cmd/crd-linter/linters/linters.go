/*
Copyright 2021 The Kubernetes Authors.

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

package linters

import (
	"fmt"
)

var allLinters = []Linter{
	// Add new linters to this list
}

var namedLinters = make(map[string]Linter)

func init() {
	for _, linter := range allLinters {
		namedLinters[linter.Name()] = linter
	}
}

// All returns all registered linters
func All() LinterList {
	linters := make([]Linter, len(allLinters))
	copy(linters, allLinters)
	return linters
}

// Named returns a list of Linters with the given names, or an error if any
// names passed to it are not recognised
func Named(names []string) (LinterList, error) {
	var linters []Linter
	for _, n := range names {
		if l, ok := namedLinters[n]; ok {
			linters = append(linters, l)
			continue
		}

		return nil, fmt.Errorf("unrecognised linter: %s", n)
	}
	return linters, nil
}

// LinterList is a list of linters, used to provide methods that work across a whole list
type LinterList []Linter

// Names calls Name() on each Linter in the LinterList
func (list LinterList) Names() []string {
	var names []string
	for _, l := range list {
		names = append(names, l.Name())
	}
	return names
}
