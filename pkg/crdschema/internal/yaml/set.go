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

var tombStone = &struct{}{}

// SetNestedField sets a value at a path in a YAML in-memory object,
// creating objects on the way.
func SetNestedField(x interface{}, v interface{}, pth ...string) (interface{}, error) {
	if len(pth) == 0 {
		return v, nil
	}

	if x == nil {
		if v == tombStone {
			return nil, nil
		}
		result, err := SetNestedField(nil, v, pth[1:]...)
		if err != nil {
			return nil, fmt.Errorf("%s.%s", pth[0], err)
		}
		return yaml.MapSlice{yaml.MapItem{Key: pth[0], Value: result}}, nil
	}

	m, ok := x.(yaml.MapSlice)
	if !ok {
		return nil, fmt.Errorf("%s is not an object", strings.Join(pth, "."))
	}

	foundAt := -1
	for i, item := range m {
		s, ok := item.Key.(string)
		if !ok {
			continue
		}
		if s == pth[0] {
			foundAt = i
			break
		}
	}

	if foundAt < 0 {
		if v == tombStone {
			return x, nil
		}
		ret := make(yaml.MapSlice, len(m), len(m)+1)
		copy(ret, m)
		result, err := SetNestedField(nil, v, pth[1:]...)
		if err != nil {
			return nil, fmt.Errorf("%s.%s", pth[0], err)
		}
		return append(ret, yaml.MapItem{Key: pth[0], Value: result}), nil
	}

	if len(pth) == 1 && v == tombStone {
		return append(m[:foundAt], m[foundAt+1:]...), nil
	}

	result, err := SetNestedField(m[foundAt].Value, v, pth[1:]...)
	ret := make(yaml.MapSlice, len(m))
	copy(ret, m)
	if err != nil {
		return nil, fmt.Errorf("%s.%s", pth[0], err)
	}
	ret[foundAt].Value = result
	return ret, nil
}

// DeleteNestedField deleted the value at a path in a YAML in-memory object. It's a noop of
// the path does not exist.
func DeleteNestedField(x interface{}, pth ...string) (interface{}, error) {
	return SetNestedField(x, tombStone, pth...)
}
