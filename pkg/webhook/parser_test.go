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

package webhook

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseLabelSelector(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *metav1.LabelSelector
		expectError bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:  "matchLabels with single label",
			input: "matchLabels~key1=value1",
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"key1": "value1",
				},
			},
		},
		{
			name:  "matchLabels with multiple labels",
			input: "matchLabels~key1=value1.key2=value2",
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		{
			name:  "matchLabels with hyphenated keys",
			input: "matchLabels~webhook-enabled=true",
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"webhook-enabled": "true",
				},
			},
		},
		{
			name:  "matchLabels with multiple hyphenated labels",
			input: "matchLabels~managed-by=myoperator.team=platform",
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"managed-by": "myoperator",
					"team":       "platform",
				},
			},
		},
		{
			name:  "matchExpressions with single expression",
			input: "matchExpressions~key=environment.operator=In.values=dev|staging|prod",
			expected: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "environment",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"dev", "staging", "prod"},
					},
				},
			},
		},
		{
			name:  "matchExpressions with NotIn operator",
			input: "matchExpressions~key=environment.operator=NotIn.values=production",
			expected: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "environment",
						Operator: metav1.LabelSelectorOpNotIn,
						Values:   []string{"production"},
					},
				},
			},
		},
		{
			name:  "matchExpressions with Exists operator",
			input: "matchExpressions~key=app.operator=Exists",
			expected: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "app",
						Operator: metav1.LabelSelectorOpExists,
					},
				},
			},
		},
		{
			name:  "combined matchLabels and matchExpressions",
			input: "matchLabels~managed-by=controller&matchExpressions~key=tier.operator=In.values=frontend|backend",
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"managed-by": "controller",
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "tier",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"frontend", "backend"},
					},
				},
			},
		},
		{
			name:  "combined with multiple labels and expression",
			input: "matchLabels~app=myapp.version=v1&matchExpressions~key=env.operator=In.values=dev|test",
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "myapp",
					"version": "v1",
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "env",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"dev", "test"},
					},
				},
			},
		},
		{
			name:        "invalid format - no selector type",
			input:       "key=value",
			expectError: true,
		},
		{
			name:        "invalid matchLabels - missing value",
			input:       "matchLabels~key1",
			expectError: true,
		},
		{
			name:        "invalid matchExpressions - missing key",
			input:       "matchExpressions~operator=In.values=a|b",
			expectError: true,
		},
		{
			name:        "invalid matchExpressions - missing operator",
			input:       "matchExpressions~key=env.values=dev",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLabelSelector(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expected == nil && result != nil {
				t.Errorf("expected nil result but got: %+v", result)
				return
			}

			if tt.expected == nil {
				return
			}

			// Compare MatchLabels
			if len(tt.expected.MatchLabels) != len(result.MatchLabels) {
				t.Errorf("MatchLabels count mismatch: expected %d, got %d", len(tt.expected.MatchLabels), len(result.MatchLabels))
			}
			for k, v := range tt.expected.MatchLabels {
				if result.MatchLabels[k] != v {
					t.Errorf("MatchLabels[%s]: expected %s, got %s", k, v, result.MatchLabels[k])
				}
			}

			// Compare MatchExpressions
			if len(tt.expected.MatchExpressions) != len(result.MatchExpressions) {
				t.Errorf("MatchExpressions count mismatch: expected %d, got %d", len(tt.expected.MatchExpressions), len(result.MatchExpressions))
			}
			for i, expectedExpr := range tt.expected.MatchExpressions {
				if i >= len(result.MatchExpressions) {
					break
				}
				resultExpr := result.MatchExpressions[i]

				if expectedExpr.Key != resultExpr.Key {
					t.Errorf("MatchExpressions[%d].Key: expected %s, got %s", i, expectedExpr.Key, resultExpr.Key)
				}
				if expectedExpr.Operator != resultExpr.Operator {
					t.Errorf("MatchExpressions[%d].Operator: expected %s, got %s", i, expectedExpr.Operator, resultExpr.Operator)
				}
				if len(expectedExpr.Values) != len(resultExpr.Values) {
					t.Errorf("MatchExpressions[%d].Values count: expected %d, got %d", i, len(expectedExpr.Values), len(resultExpr.Values))
				}
				for j, expectedVal := range expectedExpr.Values {
					if j >= len(resultExpr.Values) {
						break
					}
					if expectedVal != resultExpr.Values[j] {
						t.Errorf("MatchExpressions[%d].Values[%d]: expected %s, got %s", i, j, expectedVal, resultExpr.Values[j])
					}
				}
			}
		})
	}
}
