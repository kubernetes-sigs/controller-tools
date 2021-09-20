/*
Copyright 2019 The Kubernetes Authors.

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

package rbac

import "sort"

type stringSet map[string]struct{}

func toSet(strings []string) stringSet {
	result := make(stringSet, len(strings))
	for _, s := range strings {
		result[s] = struct{}{}
	}

	return result
}

func (ss stringSet) AddAll(other stringSet) {
	for k := range other {
		ss[k] = struct{}{}
	}
}

func (ss stringSet) IsEmpty() bool {
	return len(ss) == 0
}

func (ss stringSet) IsSuperSetOf(other stringSet) bool {
	if len(other) > len(ss) {
		return false
	}

	for k := range other {
		if _, ok := ss[k]; !ok {
			return false
		}
	}

	return true
}

func (ss stringSet) ToSorted() []string {
	if len(ss) == 0 {
		return nil // for back-compat
	}

	result := make([]string, len(ss))
	i := 0
	for k := range ss {
		result[i] = k
		i++
	}

	sort.Strings(result)
	return result
}
