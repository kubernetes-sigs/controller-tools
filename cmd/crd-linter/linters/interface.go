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

import v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

// A Linter evaluates a CRD resource and provides information about policy violations.
type Linter interface {
	// Name returns the name of this linter, used during report output
	Name() string

	// Execute runs the linter against the given custom resource definition
	Execute(*v1.CustomResourceDefinition) WarningList

	// Description prints human-readable, actionable information and references about
	// what the linter does.
	Description() string
}

type Warning struct {
	Message string
}

type WarningList []Warning

func (w WarningList) Len() int           { return len(w) }
func (w WarningList) Swap(i, j int)      { w[i], w[j] = w[j], w[i] }
func (w WarningList) Less(i, j int) bool { return w[i].Message < w[j].Message }

// Used internally to construct a warning list from a set of strings.
// This function is not exported because as the Warning struct grows, the arguments
// to the function is likely to change in backward-incompatible ways.
func newWarningList(messages ...string) WarningList {
	var w WarningList
	for _, msg := range messages {
		w = append(w, Warning{Message: msg})
	}
	return w
}
