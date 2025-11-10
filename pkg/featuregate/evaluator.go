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
	"strings"
)

const (
	boolTrueStr  = "true"
	boolFalseStr = "false"
)

// evaluateAndExpression evaluates an AND expression where all gates must be enabled.
// Format: "gate1&gate2&gate3" - returns true only if all gates are enabled.
// Also handles boolean values from parenthetical evaluation.
func (fge *FeatureGateEvaluator) evaluateAndExpression(expr string) bool {
	gates := strings.Split(expr, "&")
	for _, gate := range gates {
		gate = strings.TrimSpace(gate)
		// Handle boolean values from parenthetical evaluation
		if gate == boolTrueStr {
			continue // true AND anything = anything, so continue
		}
		if gate == boolFalseStr {
			return false // false AND anything = false
		}
		// Regular gate evaluation
		if !fge.gates.IsEnabled(gate) {
			return false
		}
	}
	return true
}

// evaluateOrExpression evaluates an OR expression where any gate can be enabled.
// Format: "gate1|gate2|gate3" - returns true if any gate is enabled.
// Also handles boolean values from parenthetical evaluation.
func (fge *FeatureGateEvaluator) evaluateOrExpression(expr string) bool {
	gates := strings.Split(expr, "|")
	for _, gate := range gates {
		gate = strings.TrimSpace(gate)
		// Handle boolean values from parenthetical evaluation
		if gate == boolTrueStr {
			return true // true OR anything = true
		}
		if gate == boolFalseStr {
			continue // false OR anything = anything, so continue
		}
		// Regular gate evaluation
		if fge.gates.IsEnabled(gate) {
			return true
		}
	}
	return false
}

// evaluateSimpleExpression evaluates feature gate expressions with support for parentheses,
// OR operations (lower precedence), and AND operations (higher precedence).
func (fge *FeatureGateEvaluator) evaluateSimpleExpression(expr string) bool {
	// Remove all spaces for easier parsing
	expr = strings.ReplaceAll(expr, " ", "")

	// Handle parentheses by evaluating them first (highest precedence)
	for strings.Contains(expr, "(") {
		// Find the innermost parentheses
		start := -1
		for i, char := range expr {
			if char == '(' {
				start = i
			} else if char == ')' && start != -1 {
				// Evaluate the expression inside the parentheses
				inner := expr[start+1 : i]
				result := fge.evaluateSimpleExpression(inner)

				// Replace the parenthetical expression with its result
				replacement := boolTrueStr
				if !result {
					replacement = boolFalseStr
				}
				expr = expr[:start] + replacement + expr[i+1:]
				break
			}
		}
	}

	// Handle special boolean values from parenthetical evaluation
	if expr == boolTrueStr {
		return true
	}
	if expr == boolFalseStr {
		return false
	}

	// Handle OR operations (lower precedence)
	if strings.Contains(expr, "|") {
		parts := strings.Split(expr, "|")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if fge.evaluateAndPart(part) {
				return true
			}
		}
		return false
	}

	// No OR operators, evaluate as AND expression or single gate
	return fge.evaluateAndPart(expr)
}

// evaluateAndPart evaluates a part that may contain AND operations or be a single gate.
func (fge *FeatureGateEvaluator) evaluateAndPart(expr string) bool {
	// Handle special boolean values
	if expr == boolTrueStr {
		return true
	}
	if expr == boolFalseStr {
		return false
	}

	// Handle AND operations
	if strings.Contains(expr, "&") {
		return fge.evaluateAndExpression(expr)
	}

	// Single gate
	return fge.gates.IsEnabled(strings.TrimSpace(expr))
}
