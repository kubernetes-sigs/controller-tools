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

package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFeatureGates(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FeatureGateMap
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
			name:  "invalid format ignored",
			input: "alpha=true,invalid,beta=false",
			expected: FeatureGateMap{
				"alpha": true,
				"beta":  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFeatureGates(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateFeatureGateExpression(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "empty expression",
			expr:      "",
			expectErr: false,
		},
		{
			name:      "simple gate name",
			expr:      "alpha",
			expectErr: false,
		},
		{
			name:      "gate with hyphen",
			expr:      "alpha-feature",
			expectErr: false,
		},
		{
			name:      "gate with underscore",
			expr:      "alpha_feature",
			expectErr: false,
		},
		{
			name:      "OR expression",
			expr:      "alpha|beta",
			expectErr: false,
		},
		{
			name:      "AND expression",
			expr:      "alpha&beta",
			expectErr: false,
		},
		{
			name:      "mixed AND OR operators",
			expr:      "alpha&beta|gamma",
			expectErr: true,
			errMsg:    "cannot mix '&' and '|' operators",
		},
		{
			name:      "invalid character",
			expr:      "alpha@beta",
			expectErr: true,
			errMsg:    "invalid character '@'",
		},
		{
			name:      "invalid character space",
			expr:      "alpha beta",
			expectErr: true,
			errMsg:    "invalid character ' '",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFeatureGateExpression(tt.expr)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShouldIncludeWebhook(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		enabledGates FeatureGateMap
		expected     bool
	}{
		{
			name: "no feature gate always included",
			config: &Config{
				Name:        "test-webhook",
				FeatureGate: "",
			},
			enabledGates: FeatureGateMap{},
			expected:     true,
		},
		{
			name: "single gate enabled",
			config: &Config{
				Name:        "alpha-webhook",
				FeatureGate: "alpha",
			},
			enabledGates: FeatureGateMap{
				"alpha": true,
			},
			expected: true,
		},
		{
			name: "single gate disabled",
			config: &Config{
				Name:        "alpha-webhook",
				FeatureGate: "alpha",
			},
			enabledGates: FeatureGateMap{
				"alpha": false,
			},
			expected: false,
		},
		{
			name: "single gate not present",
			config: &Config{
				Name:        "alpha-webhook",
				FeatureGate: "alpha",
			},
			enabledGates: FeatureGateMap{
				"beta": true,
			},
			expected: false,
		},
		{
			name: "OR expression - first gate enabled",
			config: &Config{
				Name:        "alpha-beta-webhook",
				FeatureGate: "alpha|beta",
			},
			enabledGates: FeatureGateMap{
				"alpha": true,
				"beta":  false,
			},
			expected: true,
		},
		{
			name: "OR expression - second gate enabled",
			config: &Config{
				Name:        "alpha-beta-webhook",
				FeatureGate: "alpha|beta",
			},
			enabledGates: FeatureGateMap{
				"alpha": false,
				"beta":  true,
			},
			expected: true,
		},
		{
			name: "OR expression - both gates enabled",
			config: &Config{
				Name:        "alpha-beta-webhook",
				FeatureGate: "alpha|beta",
			},
			enabledGates: FeatureGateMap{
				"alpha": true,
				"beta":  true,
			},
			expected: true,
		},
		{
			name: "OR expression - no gates enabled",
			config: &Config{
				Name:        "alpha-beta-webhook",
				FeatureGate: "alpha|beta",
			},
			enabledGates: FeatureGateMap{
				"alpha": false,
				"beta":  false,
			},
			expected: false,
		},
		{
			name: "OR expression - gates not present",
			config: &Config{
				Name:        "alpha-beta-webhook",
				FeatureGate: "alpha|beta",
			},
			enabledGates: FeatureGateMap{
				"gamma": true,
			},
			expected: false,
		},
		{
			name: "AND expression - both gates enabled",
			config: &Config{
				Name:        "alpha-beta-webhook",
				FeatureGate: "alpha&beta",
			},
			enabledGates: FeatureGateMap{
				"alpha": true,
				"beta":  true,
			},
			expected: true,
		},
		{
			name: "AND expression - first gate disabled",
			config: &Config{
				Name:        "alpha-beta-webhook",
				FeatureGate: "alpha&beta",
			},
			enabledGates: FeatureGateMap{
				"alpha": false,
				"beta":  true,
			},
			expected: false,
		},
		{
			name: "AND expression - second gate disabled",
			config: &Config{
				Name:        "alpha-beta-webhook",
				FeatureGate: "alpha&beta",
			},
			enabledGates: FeatureGateMap{
				"alpha": true,
				"beta":  false,
			},
			expected: false,
		},
		{
			name: "AND expression - first gate not present",
			config: &Config{
				Name:        "alpha-beta-webhook",
				FeatureGate: "alpha&beta",
			},
			enabledGates: FeatureGateMap{
				"beta": true,
			},
			expected: false,
		},
		{
			name: "complex OR expression with three gates",
			config: &Config{
				Name:        "multi-webhook",
				FeatureGate: "alpha|beta|gamma",
			},
			enabledGates: FeatureGateMap{
				"alpha": false,
				"beta":  false,
				"gamma": true,
			},
			expected: true,
		},
		{
			name: "complex AND expression with three gates",
			config: &Config{
				Name:        "multi-webhook",
				FeatureGate: "alpha&beta&gamma",
			},
			enabledGates: FeatureGateMap{
				"alpha": true,
				"beta":  true,
				"gamma": false,
			},
			expected: false,
		},
		{
			name: "complex AND expression with all gates enabled",
			config: &Config{
				Name:        "multi-webhook",
				FeatureGate: "alpha&beta&gamma",
			},
			enabledGates: FeatureGateMap{
				"alpha": true,
				"beta":  true,
				"gamma": true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIncludeWebhook(tt.config, tt.enabledGates)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWebhookFeatureGateIntegration(t *testing.T) {
	// Test the full workflow: parse -> validate -> filter
	testCases := []struct {
		name           string
		featureGates   string
		webhookConfigs []Config
		expectedCount  int
		description    string
	}{
		{
			name:         "no feature gates - all webhooks included",
			featureGates: "",
			webhookConfigs: []Config{
				{Name: "webhook1", FeatureGate: ""},
				{Name: "webhook2", FeatureGate: "alpha"},
				{Name: "webhook3", FeatureGate: "beta"},
			},
			expectedCount: 1, // Only webhook1 (no feature gate) is included
			description:   "When no feature gates are enabled, only webhooks without feature gates are included",
		},
		{
			name:         "alpha enabled - alpha webhooks included",
			featureGates: "alpha=true",
			webhookConfigs: []Config{
				{Name: "webhook1", FeatureGate: ""},
				{Name: "webhook2", FeatureGate: "alpha"},
				{Name: "webhook3", FeatureGate: "beta"},
			},
			expectedCount: 2, // webhook1 (no gate) and webhook2 (alpha) are included
			description:   "When alpha is enabled, webhooks without gates and alpha webhooks are included",
		},
		{
			name:         "multiple gates with OR logic",
			featureGates: "alpha=true,beta=false",
			webhookConfigs: []Config{
				{Name: "webhook1", FeatureGate: ""},
				{Name: "webhook2", FeatureGate: "alpha"},
				{Name: "webhook3", FeatureGate: "beta"},
				{Name: "webhook4", FeatureGate: "alpha|beta"},
			},
			expectedCount: 3, // webhook1 (no gate), webhook2 (alpha), webhook4 (alpha|beta - alpha is true)
			description:   "OR logic includes webhook if any gate is enabled",
		},
		{
			name:         "multiple gates with AND logic",
			featureGates: "alpha=true,beta=true",
			webhookConfigs: []Config{
				{Name: "webhook1", FeatureGate: ""},
				{Name: "webhook2", FeatureGate: "alpha"},
				{Name: "webhook3", FeatureGate: "beta"},
				{Name: "webhook4", FeatureGate: "alpha&beta"},
			},
			expectedCount: 4, // All webhooks included
			description:   "AND logic includes webhook if all gates are enabled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse feature gates
			enabledGates := parseFeatureGates(tc.featureGates)

			// Filter webhooks
			includedCount := 0
			for _, config := range tc.webhookConfigs {
				if shouldIncludeWebhook(&config, enabledGates) {
					includedCount++
				}
			}

			assert.Equal(t, tc.expectedCount, includedCount, tc.description)
		})
	}
}

func TestWebhookFeatureGateValidationInGenerate(t *testing.T) {
	// Test validation errors that would occur in the Generate method
	testCases := []struct {
		name          string
		featureGate   string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid single gate",
			featureGate: "alpha",
			expectError: false,
		},
		{
			name:        "valid OR expression",
			featureGate: "alpha|beta",
			expectError: false,
		},
		{
			name:        "valid AND expression",
			featureGate: "alpha&beta",
			expectError: false,
		},
		{
			name:          "invalid mixed operators",
			featureGate:   "alpha&beta|gamma",
			expectError:   true,
			errorContains: "cannot mix '&' and '|' operators",
		},
		{
			name:          "invalid character",
			featureGate:   "alpha@beta",
			expectError:   true,
			errorContains: "invalid character '@'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{
				FeatureGate: tc.featureGate,
			}

			err := validateFeatureGateExpression(config.FeatureGate)
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
