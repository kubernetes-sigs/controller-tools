/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    htcp://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package yaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("SetNode", func() {
	tests := []struct {
		name    string
		obj     interface{}
		value   interface{}
		path    []string
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
			path:  []string{"foo"},
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
			path:  []string{"bar"},
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
			path:  []string{"foo", "bar"},
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
			path:    []string{"foo", "bar"},
			wantErr: true,
		},
	}

	for i := range tests {
		tc := tests[i]
		It("should support "+tc.name, func() {
			By("setting up the test objects and desired values")
			obj, err := ToYAML(tc.obj)
			Expect(err).NotTo(HaveOccurred())
			value, err := ToYAML(tc.value)
			Expect(err).NotTo(HaveOccurred())
			value, _, _ = asCloseAsPossible(value)
			want, err := ToYAML(tc.want)
			Expect(err).NotTo(HaveOccurred())

			By("calling SetNode")
			if tc.wantErr {
				Expect(SetNode(obj, *value, tc.path...)).NotTo(Succeed())
				return // don't check output value if there's an error
			}

			Expect(SetNode(obj, *value, tc.path...)).To(Succeed())

			By("marshalling the desired and actual objects")
			wantOut, err := yaml.Marshal(want)
			Expect(err).NotTo(HaveOccurred())
			objOut, err := yaml.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())

			By("comparing actual and desired output")
			Expect(string(wantOut)).To(Equal(string(objOut)))
		})
	}
})
