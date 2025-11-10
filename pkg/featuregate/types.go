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

// FeatureGateMap represents enabled feature gates as a map for efficient lookup.
// Key is the gate name, value indicates if the gate is enabled.
type FeatureGateMap map[string]bool

// IsEnabled checks if a feature gate is enabled.
// Returns true only if the gate exists and is explicitly set to true.
// Gates not present in the map are considered disabled.
func (fg FeatureGateMap) IsEnabled(gateName string) bool {
	enabled, exists := fg[gateName]
	return exists && enabled
}

// FeatureGateEvaluator provides methods for parsing and evaluating feature gate expressions.
type FeatureGateEvaluator struct {
	gates FeatureGateMap
}

// NewFeatureGateEvaluator creates a new FeatureGateEvaluator with the given gates.
func NewFeatureGateEvaluator(gates FeatureGateMap) *FeatureGateEvaluator {
	return &FeatureGateEvaluator{gates: gates}
}

// EvaluateExpression evaluates a feature gate expression and returns whether it should be included.
// Supports the following formats:
// - Empty string: always returns true (no gating)
// - Single gate: "alpha" - returns true if alpha=true
// - OR logic: "alpha|beta" - returns true if either alpha=true OR beta=true
// - AND logic: "alpha&beta" - returns true if both alpha=true AND beta=true
// - Complex precedence: "(alpha&beta)|gamma" - returns true if (alpha AND beta) OR gamma
func (fge *FeatureGateEvaluator) EvaluateExpression(expr string) bool {
	if expr == "" {
		// No feature gate specified, always include
		return true
	}

	// Use the unified expression evaluator for all cases
	return fge.evaluateSimpleExpression(expr)
}
