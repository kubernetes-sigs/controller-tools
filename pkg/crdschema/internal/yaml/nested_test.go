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
	"reflect"
	"testing"
)

func TestGetNestedFieldNoCopy(t *testing.T) {
	tests := []struct {
		name    string
		obj     interface{}
		pth     []string
		want    interface{}
		found   bool
		wantErr bool
	}{
		{name: "null",
			found: true,
		},
		{name: "empty",
			obj:   map[string]interface{}{},
			found: true,
			want:  map[string]interface{}{},
		},
		{name: "non-empty, wrong path",
			obj: map[string]interface{}{
				"bar": int64(42),
			},
			pth: []string{"foo"},
		},
		{name: "non-empty, right path",
			obj: map[string]interface{}{
				"bar": int64(42),
			},
			pth:   []string{"bar"},
			want:  int64(42),
			found: true,
		},
		{name: "non-empty, long path",
			obj: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": int64(42),
				},
			},
			pth:   []string{"foo", "bar"},
			want:  int64(42),
			found: true,
		},
		{name: "invalid type",
			obj: map[string]interface{}{
				"foo": []interface{}{int64(42)},
			},
			pth:     []string{"foo", "bar"},
			wantErr: true,
			found:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, _ := ToYAML(tt.obj)

			x, found, err := GetNestedFieldNoCopy(obj, tt.pth...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNestedFieldNoCopy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want, _ := ToYAML(tt.want)
			if !reflect.DeepEqual(x, want) {
				t.Errorf("GetNestedFieldNoCopy() got = %+v : %T, want %+v : %T", x, x, want, want)
			}
			if found != tt.found {
				t.Errorf("GetNestedFieldNoCopy() found = %v, expected to be found = %v", found, tt.found)
			}
		})
	}
}
