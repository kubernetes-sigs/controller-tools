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

package yaml

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

// GetNestedFieldNoCopy gets a value at a path in a YAML in-memory object.
func GetNestedFieldNoCopy(x interface{}, pth ...string) (interface{}, bool, error) {
	if len(pth) == 0 {
		return x, true, nil
	}
	m, ok := x.(yaml.MapSlice)
	if !ok {
		return nil, false, fmt.Errorf("%s is not an object, but %T", strings.Join(pth, "."), x)
	}
	for _, item := range m {
		s, ok := item.Key.(string)
		if !ok {
			continue
		}
		if s == pth[0] {
			ret, found, err := GetNestedFieldNoCopy(item.Value, pth[1:]...)
			if err != nil {
				return ret, found, fmt.Errorf("%s.%s", pth[0], err)
			}
			return ret, found, nil
		}
	}
	return nil, false, nil
}

// GetNestedFieldSliceNoCopy gets a slice at a path in a YAML in-memory object.
func GetNestedFieldSliceNoCopy(x interface{}, pth ...string) ([]interface{}, bool, error) {
	v, found, err := GetNestedFieldNoCopy(x, pth...)
	if err != nil {
		return nil, false, err
	}
	if !found || v == nil {
		return nil, found, nil
	}
	if v, ok := v.([]interface{}); ok {
		return v, true, nil
	}
	return nil, false, fmt.Errorf("%s is not an array, but %T", strings.Join(pth, "."), v)
}

// GetNestedFieldString gets a string at a path in a YAML in-memory object.
func GetNestedFieldString(x interface{}, pth ...string) (string, bool, error) {
	v, found, err := GetNestedFieldNoCopy(x, pth...)
	if err != nil {
		return "", false, err
	}
	if !found {
		return "", found, nil
	}
	if v, ok := v.(string); ok {
		return v, true, nil
	}
	return "", false, fmt.Errorf("%s is not a string, but %T", strings.Join(pth, "."), v)
}
