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
	"sigs.k8s.io/controller-tools/cmd/crd-linter/exceptions"
	"sigs.k8s.io/controller-tools/cmd/crd-linter/linters"
	"sigs.k8s.io/controller-tools/cmd/crd-linter/reader"
)

func Evaluate(list linters.LinterList, eval []reader.CRDToEvaluate, excepted *exceptions.ExceptionList) ResultsList {
	results := make([]Results, len(eval))
	for i, e := range eval {
		results[i] = Results{
			Evaluated: e,
		}
		for _, l := range list {
			errs := l.Execute(e.CustomResourceDefinition)
			if len(errs) == 0 {
				continue
			}

			// If no exceptions are provided, skip the filtering logic
			if excepted == nil {
				results[i].Violations = append(results[i].Violations, ViolationList{
					Linter:     l,
					Violations: errs,
				})
				continue
			}

			// Check results against the exception list
			var filteredErrs []string
			for _, err := range errs {
				if !excepted.IsExcepted(e.OriginalFilename, e.CustomResourceDefinition.Name, l.Name(), err) {
					filteredErrs = append(filteredErrs, err)
				}
			}
			results[i].Violations = append(results[i].Violations, ViolationList{
				Linter:     l,
				Violations: filteredErrs,
			})
		}
	}
	return results
}
