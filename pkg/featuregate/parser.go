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
	"fmt"
	"strings"
)

// ParseFeatureGates parses a comma-separated feature gate string into a FeatureGateMap.
// Format: "gate1=true,gate2=false,gate3=true"
//
// With strict validation enabled, this function will return an error for:
// - Invalid format (missing = or wrong number of parts)
// - Invalid values (anything other than "true" or "false")
//
// Returns a FeatureGateMap and an error if parsing fails with strict validation.
func ParseFeatureGates(featureGates string, strict bool) (FeatureGateMap, error) {
	gates := make(FeatureGateMap)
	if featureGates == "" {
		return gates, nil
	}

	pairs := strings.Split(featureGates, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), "=")
		if len(parts) != 2 {
			if strict {
				return nil, fmt.Errorf("invalid feature gate format: %s (expected format: gate1=true,gate2=false)", pair)
			}
			// In non-strict mode, skip invalid entries
			continue
		}

		gateName := strings.TrimSpace(parts[0])
		gateValue := strings.TrimSpace(parts[1])

		switch gateValue {
		case "true":
			gates[gateName] = true
		case "false":
			gates[gateName] = false
		default:
			if strict {
				return nil, fmt.Errorf("invalid feature gate value for %s: %s (must be 'true' or 'false')", gateName, gateValue)
			}
			// In non-strict mode, treat invalid values as false
			gates[gateName] = false
		}
	}

	return gates, nil
}
