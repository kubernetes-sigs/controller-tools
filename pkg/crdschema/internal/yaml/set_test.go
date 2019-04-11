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

func TestSetNestedFieldNoCopy(t *testing.T) {
	tests := []struct {
		name    string
		obj     interface{}
		value   interface{}
		pth     []string
		want    interface{}
		wantErr bool
	}{
		{name: "null"},
		{name: "empty, null value",
			obj:   map[string]interface{}{},
			value: nil,
			want:  nil,
		},
		{name: "empty, non-null value",
			obj:   map[string]interface{}{},
			value: int64(64),
			want:  int64(64),
		},
		{name: "non-empty, unknown path",
			obj: map[string]interface{}{
				"bar": int64(42),
			},
			value: "foo",
			pth:   []string{"foo"},
			want: map[string]interface{}{
				"bar": int64(42),
				"foo": "foo",
			},
		},
		{name: "non-empty, existing path",
			obj: map[string]interface{}{
				"bar": int64(42),
			},
			value: "foo",
			pth:   []string{"bar"},
			want: map[string]interface{}{
				"bar": "foo",
			},
		},
		{name: "non-empty, long path",
			obj: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": int64(42),
				},
			},
			value: "foo",
			pth:   []string{"foo", "bar"},
			want: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "foo",
				},
			},
		},
		{name: "invalid type",
			obj: map[string]interface{}{
				"foo": []interface{}{int64(42)},
			},
			value:   "foo",
			pth:     []string{"foo", "bar"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, _ := ToYAML(tt.obj)
			value, _ := ToYAML(tt.value)

			x, err := SetNestedField(obj, value, tt.pth...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetNestedFieldNoCopy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want, _ := ToYAML(tt.want)
			if !reflect.DeepEqual(x, want) {
				t.Errorf("SetNestedFieldNoCopy() got = %+v : %T, want %+v : %T", x, x, want, want)
			}
		})
	}
}
