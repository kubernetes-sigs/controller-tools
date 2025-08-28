/*
Copyright 2025.

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

package genall

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/controller-tools/pkg/featuregate"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// OutputToFeatureGateDirectories outputs each object to a directory based on enabled feature gates.
// Each combination of feature gates gets its own subdirectory under the base path.
//
// For example, with feature gates "alpha=true,beta=false,gamma=true", files would be output to:
// - basePath/alpha_gamma/ (for gates that are enabled)
// - basePath/no_gates/ (for objects without feature gate requirements)
// - basePath/alpha/ (for objects requiring only alpha)
// - basePath/alpha_beta/ (for objects requiring both alpha and beta when beta is disabled - these would be skipped)
//
// The directory name format is: enabled_gates_sorted_alphabetically
type OutputToFeatureGateDirectories struct {
	// BasePath is the base directory path where feature gate directories will be created
	BasePath string `marker:"basePath"`

	// FeatureGates is the current feature gate configuration string (e.g., "alpha=true,beta=false")
	FeatureGates string `marker:"featureGates"`

	// IncludeDisabled controls whether to create directories for disabled feature gate combinations
	IncludeDisabled bool `marker:"includeDisabled,optional"`
}

// Open creates a writer for the given package and filename, placing it in the appropriate
// feature gate directory based on the current feature gate configuration.
func (o OutputToFeatureGateDirectories) Open(pkg *loader.Package, itemPath string) (io.WriteCloser, error) {
	if o.BasePath == "" {
		return nil, fmt.Errorf("base path is required for feature gate directory output")
	}

	// Parse current feature gates
	featureGateMap, err := featuregate.ParseFeatureGates(o.FeatureGates, false)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feature gates: %w", err)
	}

	// Determine the appropriate directory based on enabled feature gates
	subDir := o.getFeatureGateDirectory(featureGateMap)
	fullPath := filepath.Join(o.BasePath, subDir)

	// Ensure the directory exists
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", fullPath, err)
	}

	// Create the file
	filePath := filepath.Join(fullPath, itemPath)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories for %s: %w", filePath, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", filePath, err)
	}

	return file, nil
}

// getFeatureGateDirectory determines the subdirectory name based on enabled feature gates
func (o OutputToFeatureGateDirectories) getFeatureGateDirectory(featureGates featuregate.FeatureGateMap) string {
	var enabledGates []string

	// Collect enabled gates
	for gate, enabled := range featureGates {
		if enabled {
			enabledGates = append(enabledGates, gate)
		}
	}

	if len(enabledGates) == 0 {
		return "no_gates"
	}

	// Sort gates for consistent directory names
	// Simple bubble sort for small slices
	for i := 0; i < len(enabledGates)-1; i++ {
		for j := 0; j < len(enabledGates)-i-1; j++ {
			if enabledGates[j] > enabledGates[j+1] {
				enabledGates[j], enabledGates[j+1] = enabledGates[j+1], enabledGates[j]
			}
		}
	}

	return strings.Join(enabledGates, "_")
}

// Help returns help information for the feature gate directory output rule
func (OutputToFeatureGateDirectories) Help() *markers.DefinitionHelp {
	return &markers.DefinitionHelp{
		Category: "output",
		DetailedHelp: markers.DetailedHelp{
			Summary: "output to feature gate-based directories",
			Details: `Outputs files to directories based on enabled feature gates.

Each combination of enabled feature gates gets its own subdirectory:
- "alpha_beta" for files when both alpha and beta gates are enabled
- "alpha" for files when only alpha gate is enabled  
- "no_gates" for files that don't depend on feature gates

This enables generating multiple CRD variants for different
feature gate configurations in a single run.`,
		},
		FieldHelp: map[string]markers.DetailedHelp{
			"basePath": {
				Summary: "the base directory path where feature gate directories will be created",
			},
			"featureGates": {
				Summary: "the current feature gate configuration string (e.g., \"alpha=true,beta=false\")",
			},
			"includeDisabled": {
				Summary: "whether to create directories for disabled feature gate combinations",
			},
		},
	}
}
