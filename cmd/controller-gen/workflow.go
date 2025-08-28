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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-tools/pkg/featuregate"
	"sigs.k8s.io/controller-tools/pkg/genall"
)

// NewWorkflowCommand provides advanced CLI workflows for multi-gate, multi-output generation
func NewWorkflowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Advanced workflows for multi-gate, multi-output generation",
		Long: `Advanced workflows for generating multiple CRD variants with different feature gate combinations.

This enables developers to:
- Generate CRDs for multiple feature gate combinations in a single command
- Create progressive rollout scenarios
- Test feature gate matrix combinations

Examples:
  # Generate CRDs for all feature gate combinations
  controller-gen workflow matrix --gates=alpha,beta --base-path=./output

  # Generate CRDs for progressive rollout
  controller-gen workflow progressive --gates=alpha,beta,gamma --base-path=./output`,
	}

	cmd.AddCommand(NewMatrixCommand())
	cmd.AddCommand(NewProgressiveCommand())

	return cmd
}

// createWorkflowCommand creates a workflow command with common flags and validation
func createWorkflowCommand(use, short, long string, runFunc func([]string, string, []string) error) *cobra.Command {
	var gates []string
	var basePath string
	var paths []string

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(gates) == 0 {
				return fmt.Errorf("at least one feature gate must be specified")
			}
			if basePath == "" {
				return fmt.Errorf("base path must be specified")
			}
			if len(paths) == 0 {
				paths = []string{"./..."}
			}

			return runFunc(gates, basePath, paths)
		},
	}

	cmd.Flags().StringSliceVar(&gates, "gates", nil, "Feature gates to generate combinations for")
	cmd.Flags().StringVar(&basePath, "base-path", "", "Base output directory")
	cmd.Flags().StringSliceVar(&paths, "paths", []string{"./..."}, "Go package paths to process")

	_ = cmd.MarkFlagRequired("gates")
	_ = cmd.MarkFlagRequired("base-path")

	return cmd
}

// NewMatrixCommand generates CRDs for all possible feature gate combinations
func NewMatrixCommand() *cobra.Command {
	return createWorkflowCommand(
		"matrix",
		"Generate CRDs for all possible feature gate combinations",
		`Generate CRDs for all possible combinations of the specified feature gates.

This will create directories for each combination:
- no_gates/ (all gates disabled)
- alpha/ (only alpha enabled)  
- beta/ (only beta enabled)
- alpha_beta/ (both alpha and beta enabled)
- etc.

This is useful for testing and packaging different feature variants.`,
		generateMatrix,
	)
}

// NewProgressiveCommand generates CRDs for progressive feature rollout
func NewProgressiveCommand() *cobra.Command {
	return createWorkflowCommand(
		"progressive",
		"Generate CRDs for progressive feature rollout",
		`Generate CRDs for progressive feature rollout scenarios:

- stage_0/ (stable features only)
- stage_1/ (stable + first gate)
- stage_2/ (stable + first two gates)  
- stage_N/ (stable + all gates)

This enables gradual feature introduction in production environments.`,
		generateProgressive,
	)
}

// generateMatrix generates all possible feature gate combinations
func generateMatrix(gates []string, basePath string, paths []string) error {
	// Generate all 2^n combinations
	n := len(gates)
	totalCombinations := 1 << n

	fmt.Printf("Generating %d feature gate combinations for gates: %v\n", totalCombinations, gates)

	for i := 0; i < totalCombinations; i++ {
		var enabledGates []string
		var gateSettings []string

		for j, gate := range gates {
			if (i>>j)&1 == 1 {
				enabledGates = append(enabledGates, gate)
				gateSettings = append(gateSettings, fmt.Sprintf("%s=true", gate))
			} else {
				gateSettings = append(gateSettings, fmt.Sprintf("%s=false", gate))
			}
		}

		combination := strings.Join(gateSettings, ",")
		outputDir := getOutputDirectory(enabledGates)

		err := runGenerationForCombination(combination, filepath.Join(basePath, outputDir), paths)
		if err != nil {
			return fmt.Errorf("failed to generate for combination %s: %w", combination, err)
		}
	}

	fmt.Printf("Successfully generated all %d combinations in %s\n", totalCombinations, basePath)
	return nil
}

// generateProgressive generates progressive rollout configurations
func generateProgressive(gates []string, basePath string, paths []string) error {
	stages := len(gates) + 1

	fmt.Printf("Generating %d progressive stages for gates: %v\n", stages, gates)

	// Stage 0: all gates disabled
	stage0Settings := make([]string, len(gates))
	for i, gate := range gates {
		stage0Settings[i] = fmt.Sprintf("%s=false", gate)
	}

	err := runGenerationForCombination(
		strings.Join(stage0Settings, ","),
		filepath.Join(basePath, "stage_0"),
		paths,
	)
	if err != nil {
		return fmt.Errorf("failed to generate stage 0: %w", err)
	}

	// Progressive stages: enable one more gate at each stage
	for i := 1; i <= len(gates); i++ {
		var stageSettings []string
		for j, gate := range gates {
			if j < i {
				stageSettings = append(stageSettings, fmt.Sprintf("%s=true", gate))
			} else {
				stageSettings = append(stageSettings, fmt.Sprintf("%s=false", gate))
			}
		}

		err := runGenerationForCombination(
			strings.Join(stageSettings, ","),
			filepath.Join(basePath, fmt.Sprintf("stage_%d", i)),
			paths,
		)
		if err != nil {
			return fmt.Errorf("failed to generate stage %d: %w", i, err)
		}
	}

	fmt.Printf("Successfully generated all %d progressive stages in %s\n", stages, basePath)
	return nil
}

// getOutputDirectory determines the output directory name based on enabled gates
func getOutputDirectory(enabledGates []string) string {
	if len(enabledGates) == 0 {
		return "no_gates"
	}

	// Sort gates for consistent naming
	for i := 0; i < len(enabledGates)-1; i++ {
		for j := 0; j < len(enabledGates)-i-1; j++ {
			if enabledGates[j] > enabledGates[j+1] {
				enabledGates[j], enabledGates[j+1] = enabledGates[j+1], enabledGates[j]
			}
		}
	}

	return strings.Join(enabledGates, "_")
}

// runGenerationForCombination runs controller-gen for a specific feature gate combination
func runGenerationForCombination(featureGates string, outputPath string, paths []string) error {
	// Create output directory
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputPath, err)
	}

	fmt.Printf("  Generating: %s -> %s\n", featureGates, outputPath)

	// Parse feature gates to validate them
	_, err := featuregate.ParseFeatureGates(featureGates, false)
	if err != nil {
		return fmt.Errorf("invalid feature gates %s: %w", featureGates, err)
	}

	// Create a runtime configuration for this combination
	args := []string{
		"crd",
		fmt.Sprintf("crd:featureGates=%s", featureGates),
		fmt.Sprintf("output:crd:dir=%s", outputPath),
	}

	// Add paths
	for _, path := range paths {
		args = append(args, fmt.Sprintf("paths=%s", path))
	}

	// Use the existing optionsRegistry and run generation
	rt, err := genall.FromOptions(optionsRegistry, args)
	if err != nil {
		return fmt.Errorf("failed to create runtime for combination %s: %w", featureGates, err)
	}

	if len(rt.Generators) == 0 {
		return fmt.Errorf("no generators specified for combination %s", featureGates)
	}

	if hadErrs := rt.Run(); hadErrs {
		return fmt.Errorf("generation failed for combination %s", featureGates)
	}

	return nil
}
