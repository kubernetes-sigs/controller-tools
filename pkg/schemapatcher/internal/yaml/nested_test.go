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
)

var _ = Describe("GetNode", func() {
	tests := []struct {
		name    string
		obj     interface{}
		path    []string
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
			path: []string{"foo"},
		},
		{name: "non-empty, right path",
			obj: map[string]interface{}{
				"bar": int64(42),
			},
			path:  []string{"bar"},
			want:  int64(42),
			found: true,
		},
		{name: "non-empty, long path",
			obj: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": int64(42),
				},
			},
			path:  []string{"foo", "bar"},
			want:  int64(42),
			found: true,
		},
		{name: "invalid type",
			obj: map[string]interface{}{
				"foo": []interface{}{int64(42)},
			},
			path:    []string{"foo", "bar"},
			wantErr: true,
			found:   false,
		},
	}

	for i := range tests {
		tc := tests[i]
		It("should support "+tc.name, func() {
			By("setting up the test objects and desired values")
			obj, err := ToYAML(tc.obj)
			Expect(err).NotTo(HaveOccurred())
			want, err := ToYAML(tc.want)
			Expect(err).NotTo(HaveOccurred())
			// remove the document node
			want, _, _ = asCloseAsPossible(want)

			By("calling GetNode")
			node, found, err := GetNode(obj, tc.path...)
			if tc.wantErr {
				Expect(err).To(HaveOccurred())
				return // don't check output value if there's an error
			}

			Expect(err).NotTo(HaveOccurred())

			// remove line/column info to compare
			if node != nil {
				node.Line = 0
				node.Column = 0
			}
			if want != nil {
				want.Line = 0
				want.Column = 0
			}

			By("confirming that the returned node is as we expect it")
			Expect(found).To(Equal(tc.found))
			if found {
				Expect(node).To(Equal(want))
			}
		})
	}
})
