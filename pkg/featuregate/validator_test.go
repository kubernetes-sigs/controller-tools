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
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestValidateFeatureGateExpression(t *testing.T) {
	knownGates := sets.New("alpha", "beta", "gamma", "v1beta1", "my-feature", "under_score")

	tests := []struct {
		name          string
		expr          string
		knownGates    sets.Set[string]
		strict        bool
		expectError   bool
		errorContains string
	}{
		{
			name:       "empty expression",
			expr:       "",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "simple gate name",
			expr:       "alpha",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "gate with hyphen",
			expr:       "my-feature",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "gate with underscore",
			expr:       "under_score",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "gate with numbers",
			expr:       "v1beta1",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "OR expression",
			expr:       "alpha|beta",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "AND expression",
			expr:       "alpha&beta",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:          "mixed AND OR operators",
			expr:          "alpha&beta|gamma",
			knownGates:    knownGates,
			strict:        true,
			expectError:   true,
			errorContains: "cannot mix '&' and '|' operators",
		},
		{
			name:          "invalid character @",
			expr:          "alpha@beta",
			knownGates:    knownGates,
			strict:        true,
			expectError:   true,
			errorContains: "invalid character '@'",
		},
		{
			name:          "invalid character .",
			expr:          "alpha.beta",
			knownGates:    knownGates,
			strict:        true,
			expectError:   true,
			errorContains: "invalid character '.'",
		},
		{
			name:          "invalid character space",
			expr:          "alpha beta",
			knownGates:    knownGates,
			strict:        true,
			expectError:   true,
			errorContains: "invalid character ' '",
		},
		{
			name:          "unknown gate strict mode",
			expr:          "unknown",
			knownGates:    knownGates,
			strict:        true,
			expectError:   true,
			errorContains: "unknown feature gate 'unknown'",
		},
		{
			name:       "unknown gate non-strict mode",
			expr:       "unknown",
			knownGates: knownGates,
			strict:     false,
		},
		{
			name:          "unknown gate in OR expression strict mode",
			expr:          "alpha|unknown",
			knownGates:    knownGates,
			strict:        true,
			expectError:   true,
			errorContains: "unknown feature gate 'unknown'",
		},
		{
			name:       "unknown gate in OR expression non-strict mode",
			expr:       "alpha|unknown",
			knownGates: knownGates,
			strict:     false,
		},
		{
			name:       "no known gates provided - no validation",
			expr:       "anything",
			knownGates: nil,
			strict:     true,
		},
		{
			name:       "empty known gates set - no validation",
			expr:       "anything",
			knownGates: sets.New[string](),
			strict:     true,
		},
		// Complex precedence rule tests
		{
			name:       "simple parentheses",
			expr:       "(alpha)",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "parentheses with AND",
			expr:       "(alpha&beta)",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "parentheses with OR",
			expr:       "(alpha|beta)",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "complex precedence AND OR",
			expr:       "(alpha&beta)|gamma",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "complex precedence OR AND",
			expr:       "(alpha|beta)&gamma",
			knownGates: knownGates,
			strict:     true,
		},
		{
			name:       "nested parentheses",
			expr:       "((alpha&beta)|gamma)&delta",
			knownGates: sets.New("alpha", "beta", "gamma", "delta"),
			strict:     true,
		},
		{
			name:       "multiple OR groups",
			expr:       "(alpha&beta)|(gamma&delta)",
			knownGates: sets.New("alpha", "beta", "gamma", "delta"),
			strict:     true,
		},
		{
			name:          "unmatched opening parenthesis",
			expr:          "(alpha&beta",
			knownGates:    knownGates,
			strict:        true,
			expectError:   true,
			errorContains: "unmatched opening parenthesis",
		},
		{
			name:          "unmatched closing parenthesis",
			expr:          "alpha&beta)",
			knownGates:    knownGates,
			strict:        true,
			expectError:   true,
			errorContains: "unmatched closing parenthesis",
		},
		{
			name:          "empty parentheses",
			expr:          "()",
			knownGates:    knownGates,
			strict:        true,
			expectError:   true,
			errorContains: "empty gate name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			err := ValidateFeatureGateExpression(tt.expr, tt.knownGates, tt.strict)

			if tt.expectError {
				g.Expect(err).To(gomega.HaveOccurred())
				if tt.errorContains != "" {
					g.Expect(err.Error()).To(gomega.ContainSubstring(tt.errorContains))
				}
			} else {
				g.Expect(err).NotTo(gomega.HaveOccurred())
			}
		})
	}
}

func TestExtractGateNames(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected []string
	}{
		{
			name:     "single gate",
			expr:     "alpha",
			expected: []string{"alpha"},
		},
		{
			name:     "OR expression",
			expr:     "alpha|beta",
			expected: []string{"alpha", "beta"},
		},
		{
			name:     "AND expression",
			expr:     "alpha&beta",
			expected: []string{"alpha", "beta"},
		},
		{
			name:     "three gates OR",
			expr:     "alpha|beta|gamma",
			expected: []string{"alpha", "beta", "gamma"},
		},
		{
			name:     "three gates AND",
			expr:     "alpha&beta&gamma",
			expected: []string{"alpha", "beta", "gamma"},
		},
		{
			name:     "with spaces",
			expr:     " alpha | beta ",
			expected: []string{"alpha", "beta"},
		},
		{
			name:     "empty expression",
			expr:     "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			result := extractGateNames(tt.expr)
			g.Expect(result).To(gomega.Equal(tt.expected))
		})
	}
}

func TestIsValidCharacter(t *testing.T) {
	g := gomega.NewWithT(t)
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_&|()"
	invalidChars := "@#$%^*+=[]{}\\:;\"'<>?,./`~ "

	for _, char := range validChars {
		g.Expect(isValidCharacter(char)).To(gomega.BeTrue(), "Character '%c' should be valid", char)
	}

	for _, char := range invalidChars {
		g.Expect(isValidCharacter(char)).To(gomega.BeFalse(), "Character '%c' should be invalid", char)
	}
}
