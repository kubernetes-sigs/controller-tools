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

package featuregate

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestParseFeatureGates(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      FeatureGateMap
		expectError   bool
		errorContains string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: FeatureGateMap{},
		},
		{
			name:  "single gate enabled",
			input: "alpha=true",
			expected: FeatureGateMap{
				"alpha": true,
			},
		},
		{
			name:  "single gate disabled",
			input: "alpha=false",
			expected: FeatureGateMap{
				"alpha": false,
			},
		},
		{
			name:  "multiple gates",
			input: "alpha=true,beta=false,gamma=true",
			expected: FeatureGateMap{
				"alpha": true,
				"beta":  false,
				"gamma": true,
			},
		},
		{
			name:  "gates with spaces",
			input: " alpha = true , beta = false ",
			expected: FeatureGateMap{
				"alpha": true,
				"beta":  false,
			},
		},
		{
			name:          "invalid format",
			input:         "alpha=true,invalid,beta=false",
			expectError:   true,
			errorContains: "invalid feature gate format",
		},
		{
			name:          "invalid value",
			input:         "alpha=true,beta=maybe",
			expectError:   true,
			errorContains: "invalid feature gate value",
		},
		{
			name:  "complex gate names",
			input: "v1beta1=true,my-feature=false,under_score=true",
			expected: FeatureGateMap{
				"v1beta1":     true,
				"my-feature":  false,
				"under_score": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			result, err := ParseFeatureGates(tt.input)

			if tt.expectError {
				g.Expect(err).To(gomega.HaveOccurred())
				if tt.errorContains != "" {
					g.Expect(err.Error()).To(gomega.ContainSubstring(tt.errorContains))
				}
			} else {
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(result).To(gomega.Equal(tt.expected))
			}
		})
	}
}
