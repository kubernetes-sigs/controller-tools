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

/*
Package featuregate provides a centralized implementation for feature gate functionality
across all controller-tools generators.

This package addresses the code duplication that existed in CRD, RBAC, and Webhook generators
by providing a unified API for:

- Parsing feature gate configurations from CLI parameters
- Validating feature gate expressions with strict validation
- Evaluating complex boolean expressions (AND, OR logic)
- Managing known feature gates with a registry

# Basic Usage

The simplest way to use this package is:

	gates, err := featuregate.ParseFeatureGates("alpha=true,beta=false")
	if err != nil {
		// handle error
	}
	evaluator := featuregate.NewFeatureGateEvaluator(gates)

	if evaluator.EvaluateExpression("alpha|beta") {
		// Include the feature
	}

# Expression Syntax

Feature gate expressions support the following formats:

- Empty string: "" (always evaluates to true - no gating)
- Single gate: "alpha" (true if alpha=true)
- OR logic: "alpha|beta" (true if either alpha=true OR beta=true)
- AND logic: "alpha&beta" (true if both alpha=true AND beta=true)

Multiple gates are supported:
- "alpha|beta|gamma" (true if any gate is enabled)
- "alpha&beta&gamma" (true if all gates are enabled)

Mixing AND and OR operators in the same expression is not allowed.

# Strict Validation

For new implementations requiring strict validation:

	registry := featuregate.NewRegistry([]string{"alpha", "beta"}, true)
	evaluator, err := registry.CreateEvaluator("alpha=true,beta=false")
	if err != nil {
		// Handle parsing error
	}

	err = registry.ValidateExpression("alpha|unknown")
	if err != nil {
		// Handle unknown gate error
	}

# Integration

This package provides functions that centralize the feature gate logic
previously duplicated across CRD, RBAC, and Webhook generators:

- ParseFeatureGates() replaces individual parseFeatureGates() functions
- ValidateFeatureGateExpression() replaces individual validateFeatureGateExpression() functions
- FeatureGateEvaluator.EvaluateExpression() replaces individual shouldInclude*() functions

The FeatureGateMap type is compatible with existing map[string]bool usage patterns.
*/
package featuregate
