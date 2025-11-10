/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "	return nil
}may not use this file except in compliance with the License.
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
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// ValidateFeatureGateExpression validates the syntax of a feature gate expression.
// Returns an error if the expression contains invalid characters or mixed operators.
//
// With strict validation, unknown feature gates will cause validation to fail.
// The knownGates parameter should contain all valid feature gate names.
func ValidateFeatureGateExpression(expr string, knownGates sets.Set[string], strict bool) error {
	if expr == "" {
		return nil
	}

	// Check for invalid characters (only allow alphanumeric, hyphens, underscores, &, |)
	for _, char := range expr {
		if !isValidCharacter(char) {
			return fmt.Errorf("invalid character '%c' in feature gate expression: %s", char, expr)
		}
	}

	// Validate parentheses are balanced
	if err := validateParentheses(expr); err != nil {
		return fmt.Errorf("invalid parentheses in feature gate expression '%s': %w", expr, err)
	}

	// Validate individual gate names if strict validation is enabled
	if strict && knownGates != nil && knownGates.Len() > 0 {
		gates := extractGateNames(expr)
		for _, gate := range gates {
			if gate == "" {
				return fmt.Errorf("empty gate name in expression: %s", expr)
			}
			if !knownGates.Has(gate) {
				return fmt.Errorf("unknown feature gate '%s' in expression: %s", gate, expr)
			}
		}
	}

	return nil
}

// isValidCharacter checks if a character is valid in a feature gate expression.
func isValidCharacter(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_' ||
		char == '&' || char == '|' ||
		char == '(' || char == ')'
}

// extractGateNames extracts individual gate names from a feature gate expression.
func extractGateNames(expr string) []string {
	// Remove parentheses and replace operators with a common delimiter
	normalized := strings.ReplaceAll(expr, "(", "")
	normalized = strings.ReplaceAll(normalized, ")", "")
	normalized = strings.ReplaceAll(normalized, "&", ",")
	normalized = strings.ReplaceAll(normalized, "|", ",")

	// Handle special case of empty parentheses
	if strings.TrimSpace(normalized) == "" && (strings.Contains(expr, "(") || strings.Contains(expr, ")")) {
		return []string{""}
	}

	// Split and trim
	parts := strings.Split(normalized, ",")
	var gates []string
	for _, part := range parts {
		gate := strings.TrimSpace(part)
		if gate != "" {
			gates = append(gates, gate)
		}
	}

	return gates
}

// validateParentheses checks if parentheses are properly balanced in the expression.
func validateParentheses(expr string) error {
	count := 0
	for _, char := range expr {
		switch char {
		case '(':
			count++
		case ')':
			count--
			if count < 0 {
				return fmt.Errorf("unmatched closing parenthesis")
			}
		}
	}
	if count != 0 {
		return fmt.Errorf("unmatched opening parenthesis")
	}
	return nil
}
