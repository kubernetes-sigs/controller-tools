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
	"k8s.io/apimachinery/pkg/util/sets"
)

// Registry maintains a registry of known feature gates and provides
// centralized validation and evaluation capabilities.
type Registry struct {
	knownGates sets.Set[string]
	strict     bool
}

// NewRegistry creates a new feature gate registry.
func NewRegistry(knownGates []string, strict bool) *Registry {
	gateSet := sets.New[string]()
	for _, gate := range knownGates {
		gateSet.Insert(gate)
	}

	return &Registry{
		knownGates: gateSet,
		strict:     strict,
	}
}

// ParseAndValidate parses feature gates and validates expressions in one step.
func (r *Registry) ParseAndValidate(featureGatesStr string, expression string) (FeatureGateMap, error) {
	// Parse the feature gates
	gates, err := ParseFeatureGates(featureGatesStr)
	if err != nil {
		return nil, err
	}

	// Validate the expression
	err = ValidateFeatureGateExpression(expression, r.knownGates, r.strict)
	if err != nil {
		return nil, err
	}

	return gates, nil
}

// CreateEvaluator creates a new FeatureGateEvaluator with the parsed gates.
func (r *Registry) CreateEvaluator(featureGatesStr string) (*FeatureGateEvaluator, error) {
	gates, err := ParseFeatureGates(featureGatesStr)
	if err != nil {
		return nil, err
	}

	return NewFeatureGateEvaluator(gates), nil
}

// ValidateExpression validates a feature gate expression using the registry's settings.
func (r *Registry) ValidateExpression(expr string) error {
	return ValidateFeatureGateExpression(expr, r.knownGates, r.strict)
}

// AddKnownGates adds multiple gates to the known gates set.
func (r *Registry) AddKnownGates(gates ...string) {
	for _, gate := range gates {
		r.knownGates.Insert(gate)
	}
}

// IsKnownGate checks if a gate is in the known gates set.
func (r *Registry) IsKnownGate(gate string) bool {
	return r.knownGates.Has(gate)
}

// GetKnownGates returns a copy of the known gates set.
func (r *Registry) GetKnownGates() sets.Set[string] {
	return r.knownGates.Clone()
}
