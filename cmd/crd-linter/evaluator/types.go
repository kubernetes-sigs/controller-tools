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

package evaluator

import (
	"sort"

	"sigs.k8s.io/controller-tools/cmd/crd-linter/exceptions"
	"sigs.k8s.io/controller-tools/cmd/crd-linter/linters"
	"sigs.k8s.io/controller-tools/cmd/crd-linter/reader"
)

type Results struct {
	Evaluated reader.CRDToEvaluate

	Violations []ViolationList
}

func (r *Results) HasViolations() bool {
	if len(r.Violations) == 0 {
		return false
	}
	for _, v := range r.Violations {
		if len(v.Violations) > 0 {
			return true
		}
	}
	return false
}

func (r *Results) ViolatedLinters() []ViolationList {
	var list []ViolationList
	for _, v := range r.Violations {
		if len(v.Violations) == 0 {
			continue
		}

		list = append(list, v)
	}
	return list
}

type ResultsList []Results

func (results ResultsList) ToExceptionList() *exceptions.ExceptionList {
	list := exceptions.NewExceptionList()
	for _, r := range results {
		if !r.HasViolations() {
			continue
		}

		for _, l := range r.Violations {
			for _, v := range l.Violations {
				list.Add(r.Evaluated.OriginalFilename, r.Evaluated.CustomResourceDefinition.Name, l.Linter.Name(), v)
			}
		}
	}
	sort.SliceStable(list.Exceptions, func(i, j int) bool {
		return list.Exceptions[i].String() < list.Exceptions[j].String()
	})
	return list
}

func (results ResultsList) HasViolations() bool {
	for _, r := range results {
		if r.HasViolations() {
			return true
		}
	}
	return false
}

type ViolationList struct {
	// The Linter that triggered this set of violations
	Linter linters.Linter

	// An ordered list of violations triggered by this linter
	Violations []string
}
