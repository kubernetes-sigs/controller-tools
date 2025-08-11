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

func TestNewRegistry(t *testing.T) {
	g := gomega.NewWithT(t)
	knownGates := []string{"alpha", "beta", "gamma"}
	registry := NewRegistry(knownGates, true)

	g.Expect(registry.strict).To(gomega.BeTrue())
	g.Expect(registry.knownGates.Len()).To(gomega.Equal(3))
	g.Expect(registry.IsKnownGate("alpha")).To(gomega.BeTrue())
	g.Expect(registry.IsKnownGate("beta")).To(gomega.BeTrue())
	g.Expect(registry.IsKnownGate("gamma")).To(gomega.BeTrue())
	g.Expect(registry.IsKnownGate("unknown")).To(gomega.BeFalse())
}

func TestRegistry_ParseAndValidate(t *testing.T) {
	registry := NewRegistry([]string{"alpha", "beta"}, true)

	tests := []struct {
		name            string
		featureGatesStr string
		expression      string
		expectError     bool
		expectedGates   FeatureGateMap
	}{
		{
			name:            "valid parsing and expression",
			featureGatesStr: "alpha=true,beta=false",
			expression:      "alpha|beta",
			expectedGates: FeatureGateMap{
				"alpha": true,
				"beta":  false,
			},
		},
		{
			name:            "invalid feature gate format",
			featureGatesStr: "alpha=true,invalid",
			expression:      "alpha",
			expectError:     true,
		},
		{
			name:            "invalid expression",
			featureGatesStr: "alpha=true",
			expression:      "alpha&beta|gamma",
			expectError:     true,
		},
		{
			name:            "unknown gate in expression",
			featureGatesStr: "alpha=true",
			expression:      "unknown",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			gates, err := registry.ParseAndValidate(tt.featureGatesStr, tt.expression)

			if tt.expectError {
				g.Expect(err).To(gomega.HaveOccurred())
			} else {
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(gates).To(gomega.Equal(tt.expectedGates))
			}
		})
	}
}

func TestRegistry_CreateEvaluator(t *testing.T) {
	registry := NewRegistry([]string{"alpha", "beta"}, true)

	tests := []struct {
		name            string
		featureGatesStr string
		expectError     bool
	}{
		{
			name:            "valid feature gates",
			featureGatesStr: "alpha=true,beta=false",
		},
		{
			name:            "invalid format",
			featureGatesStr: "alpha=true,invalid",
			expectError:     true,
		},
		{
			name:            "empty string",
			featureGatesStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			evaluator, err := registry.CreateEvaluator(tt.featureGatesStr)

			if tt.expectError {
				g.Expect(err).To(gomega.HaveOccurred())
				g.Expect(evaluator).To(gomega.BeNil())
			} else {
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(evaluator).NotTo(gomega.BeNil())
			}
		})
	}
}

func TestRegistry_ValidateExpression(t *testing.T) {
	registry := NewRegistry([]string{"alpha", "beta"}, true)

	tests := []struct {
		name        string
		expr        string
		expectError bool
	}{
		{
			name: "valid expression",
			expr: "alpha|beta",
		},
		{
			name:        "unknown gate",
			expr:        "unknown",
			expectError: true,
		},
		{
			name:        "mixed operators",
			expr:        "alpha&beta|gamma",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			err := registry.ValidateExpression(tt.expr)

			if tt.expectError {
				g.Expect(err).To(gomega.HaveOccurred())
			} else {
				g.Expect(err).NotTo(gomega.HaveOccurred())
			}
		})
	}
}

func TestRegistry_AddKnownGate(t *testing.T) {
	g := gomega.NewWithT(t)
	registry := NewRegistry([]string{"alpha"}, true)

	g.Expect(registry.IsKnownGate("alpha")).To(gomega.BeTrue())
	g.Expect(registry.IsKnownGate("beta")).To(gomega.BeFalse())

	registry.AddKnownGate("beta")

	g.Expect(registry.IsKnownGate("alpha")).To(gomega.BeTrue())
	g.Expect(registry.IsKnownGate("beta")).To(gomega.BeTrue())
}

func TestRegistry_AddKnownGates(t *testing.T) {
	g := gomega.NewWithT(t)
	registry := NewRegistry([]string{"alpha"}, true)

	g.Expect(registry.knownGates.Len()).To(gomega.Equal(1))

	registry.AddKnownGates([]string{"beta", "gamma"})

	g.Expect(registry.knownGates.Len()).To(gomega.Equal(3))
	g.Expect(registry.IsKnownGate("alpha")).To(gomega.BeTrue())
	g.Expect(registry.IsKnownGate("beta")).To(gomega.BeTrue())
	g.Expect(registry.IsKnownGate("gamma")).To(gomega.BeTrue())
}

func TestRegistry_GetKnownGates(t *testing.T) {
	g := gomega.NewWithT(t)
	registry := NewRegistry([]string{"alpha", "beta"}, true)

	gates := registry.GetKnownGates()

	g.Expect(gates.Len()).To(gomega.Equal(2))
	g.Expect(gates.Has("alpha")).To(gomega.BeTrue())
	g.Expect(gates.Has("beta")).To(gomega.BeTrue())

	// Verify it's a copy - modifying returned set shouldn't affect registry
	gates.Insert("gamma")
	g.Expect(registry.IsKnownGate("gamma")).To(gomega.BeFalse())
}

func TestRegistry_Integration(t *testing.T) {
	g := gomega.NewWithT(t)
	// Test a complete workflow
	registry := NewRegistry([]string{"alpha", "beta", "gamma"}, true)

	// Create an evaluator
	evaluator, err := registry.CreateEvaluator("alpha=true,beta=false,gamma=true")
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(evaluator).NotTo(gomega.BeNil())

	// Validate and evaluate expressions
	testExpressions := []struct {
		expr     string
		expected bool
	}{
		{"", true},
		{"alpha", true},
		{"beta", false},
		{"alpha|beta", true},
		{"beta|gamma", true},
		{"alpha&gamma", true},
		{"alpha&beta", false},
	}

	for _, tt := range testExpressions {
		// Validate expression
		err := registry.ValidateExpression(tt.expr)
		g.Expect(err).NotTo(gomega.HaveOccurred(), "Expression validation failed for: %s", tt.expr)

		// Evaluate expression
		result := evaluator.EvaluateExpression(tt.expr)
		g.Expect(result).To(gomega.Equal(tt.expected), "Expression evaluation failed for: %s", tt.expr)
	}
}
