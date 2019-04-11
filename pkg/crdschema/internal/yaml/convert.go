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
	"reflect"

	yaml2 "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/json"
)

// ToYAML converts some JSON in-memory value into in-memory YAML.
func ToYAML(x interface{}) (interface{}, error) {
	if x == nil {
		return nil, nil
	}

	bs, err := json.Marshal(x)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %v", err)
	}

	t := reflect.TypeOf(x)
	nativeStruct := t.Kind() == reflect.Struct || (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct)
	if _, ok := x.(map[string]interface{}); ok || nativeStruct {
		var y yaml2.MapSlice
		if err := yaml2.Unmarshal(bs, &y); err != nil {
			return nil, fmt.Errorf("failed to unmarshal schema: %v", err)
		}
		return y, nil
	}

	var y interface{}
	if err := yaml2.Unmarshal(bs, &y); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %v", err)
	}
	return y, nil
}
