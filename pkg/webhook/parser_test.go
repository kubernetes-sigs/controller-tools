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

// TestParseLabelSelector verifies that JSON label selector strings are parsed correctly into
// LabelSelector values. Most inputs use the form used with backticks in markers (no escaped quotes).
// One case uses escaped quotes to ensure both forms produce the same result.
func TestParseLabelSelector(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *metav1.LabelSelector
		expectError bool
	}{
		{name: "empty", input: "", expected: nil},
		{
			name:  "matchLabels without escapes (backtick form)",
			input: `{"matchLabels":{"key":"value"}}`,
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{"key": "value"},
			},
		},
		{
			name:  "matchLabels with escaped quotes (same result as backtick form)",
			input: "{\"matchLabels\":{\"key\":\"value\"}}",
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{"key": "value"},
			},
		},
		{
			name:  "matchExpressions without escapes",
			input: `{"matchExpressions":[{"key":"env","operator":"In","values":["dev","prod"]}]}`,
			expected: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "env", Operator: metav1.LabelSelectorOpIn, Values: []string{"dev", "prod"}},
				},
			},
		},
		{
			name:  "matchExpressions hyphenated key, doc example without escapes",
			input: `{"matchExpressions":[{"key":"app-type","operator":"In","values":["web","api","worker"]}]}`,
			expected: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "app-type", Operator: metav1.LabelSelectorOpIn, Values: []string{"web", "api", "worker"}},
				},
			},
		},
		{
			name:  "matchLabels and matchExpressions without escapes",
			input: `{"matchLabels":{"managed-by":"controller"},"matchExpressions":[{"key":"tier","operator":"In","values":["frontend","backend"]}]}`,
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{"managed-by": "controller"},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "tier", Operator: metav1.LabelSelectorOpIn, Values: []string{"frontend", "backend"}},
				},
			},
		},
		{name: "invalid JSON", input: "{invalid", expectError: true},
		{name: "empty object", input: "{}", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLabelSelector(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}
			if len(tt.expected.MatchLabels) != len(result.MatchLabels) {
				t.Errorf("MatchLabels: expected %d, got %d", len(tt.expected.MatchLabels), len(result.MatchLabels))
			}
			for k, v := range tt.expected.MatchLabels {
				if result.MatchLabels[k] != v {
					t.Errorf("MatchLabels[%s]: expected %q, got %q", k, v, result.MatchLabels[k])
				}
			}
			if len(tt.expected.MatchExpressions) != len(result.MatchExpressions) {
				t.Errorf("MatchExpressions: expected %d, got %d", len(tt.expected.MatchExpressions), len(result.MatchExpressions))
			}
			for i, exp := range tt.expected.MatchExpressions {
				if i >= len(result.MatchExpressions) {
					break
				}
				res := result.MatchExpressions[i]
				if exp.Key != res.Key || exp.Operator != res.Operator {
					t.Errorf("MatchExpressions[%d]: expected %+v, got %+v", i, exp, res)
				}
				if len(exp.Values) != len(res.Values) {
					t.Errorf("MatchExpressions[%d].Values: expected %v, got %v", i, exp.Values, res.Values)
				}
				for j, v := range exp.Values {
					if j < len(res.Values) && res.Values[j] != v {
						t.Errorf("MatchExpressions[%d].Values[%d]: expected %q, got %q", i, j, v, res.Values[j])
					}
				}
			}
		})
	}
}
