/*
Copyright 2024 The Kubernetes Authors.

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

package version

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TestVersion", func() {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "empty returns unknown",
			version:  "",
			expected: "(devel)",
		},
		{
			name:     "set to a value returns it",
			version:  "1.2.3",
			expected: "1.2.3",
		},
	}
	for _, tc := range tests {
		It("Version set to "+tc.name, func() {
			versionBackup := version
			defer func() {
				version = versionBackup
			}()
			version = tc.version
			result := Version()
			Expect(result).To(Equal(tc.expected))
		})
	}
})
