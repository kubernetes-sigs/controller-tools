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

func TestFeatureGateMap_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		gates    FeatureGateMap
		gateName string
		expected bool
	}{
		{
			name: "enabled gate returns true",
			gates: FeatureGateMap{
				"alpha": true,
				"beta":  false,
			},
			gateName: "alpha",
			expected: true,
		},
		{
			name: "disabled gate returns false",
			gates: FeatureGateMap{
				"alpha": true,
				"beta":  false,
			},
			gateName: "beta",
			expected: false,
		},
		{
			name: "missing gate returns false",
			gates: FeatureGateMap{
				"alpha": true,
			},
			gateName: "gamma",
			expected: false,
		},
		{
			name:     "empty map returns false",
			gates:    FeatureGateMap{},
			gateName: "alpha",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			result := tt.gates.IsEnabled(tt.gateName)
			g.Expect(result).To(gomega.Equal(tt.expected))
		})
	}
}

func TestFeatureGateEvaluator_EvaluateExpression(t *testing.T) {
	evaluator := NewFeatureGateEvaluator(FeatureGateMap{
		"alpha": true,
		"beta":  false,
		"gamma": true,
		"delta": false,
	})

	tests := []struct {
		name     string
		expr     string
		expected bool
	}{
		{
			name:     "empty expression always true",
			expr:     "",
			expected: true,
		},
		{
			name:     "single enabled gate",
			expr:     "alpha",
			expected: true,
		},
		{
			name:     "single disabled gate",
			expr:     "beta",
			expected: false,
		},
		{
			name:     "single missing gate",
			expr:     "missing",
			expected: false,
		},
		{
			name:     "OR expression - first enabled",
			expr:     "alpha|beta",
			expected: true,
		},
		{
			name:     "OR expression - second enabled",
			expr:     "beta|gamma",
			expected: true,
		},
		{
			name:     "OR expression - both enabled",
			expr:     "alpha|gamma",
			expected: true,
		},
		{
			name:     "OR expression - none enabled",
			expr:     "beta|delta",
			expected: false,
		},
		{
			name:     "OR expression - three gates, one enabled",
			expr:     "beta|delta|gamma",
			expected: true,
		},
		{
			name:     "AND expression - both enabled",
			expr:     "alpha&gamma",
			expected: true,
		},
		{
			name:     "AND expression - first disabled",
			expr:     "beta&gamma",
			expected: false,
		},
		{
			name:     "AND expression - second disabled",
			expr:     "alpha&delta",
			expected: false,
		},
		{
			name:     "AND expression - both disabled",
			expr:     "beta&delta",
			expected: false,
		},
		{
			name:     "AND expression - three gates, all enabled",
			expr:     "alpha&gamma&alpha", // Using alpha twice to test multiple enabled
			expected: true,
		},
		{
			name:     "AND expression - three gates, one disabled",
			expr:     "alpha&gamma&beta",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			result := evaluator.EvaluateExpression(tt.expr)
			g.Expect(result).To(gomega.Equal(tt.expected))
		})
	}
}

func TestNewFeatureGateEvaluator(t *testing.T) {
	g := gomega.NewWithT(t)
	gates := FeatureGateMap{
		"alpha": true,
		"beta":  false,
	}

	evaluator := NewFeatureGateEvaluator(gates)

	g.Expect(evaluator).NotTo(gomega.BeNil())
	g.Expect(evaluator.gates).To(gomega.Equal(gates))
}
