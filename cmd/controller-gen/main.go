/*
Copyright 2018 The Kubernetes Authors.

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
	"log"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-tools/pkg/generate/rbac"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "controller-gen",
		Short: "A reference implementation generation tool for Kubernetes APIs.",
		Long:  `A reference implementation generation tool for Kubernetes APIs.`,
		Example: `	# Generate RBAC manifests for a project
	controller-gen rbac
	
	# Generate CRD manifests for a project
	controller-gen crd 

	# Run all the generators for a given project
	controller-gen all
`,
	}

	// add RBAC manifests generator as subcommand
	rootCmd.AddCommand(
		newRBACCmd(),
		newAllSubCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newRBACCmd() *cobra.Command {
	o := &rbac.ManifestOptions{}
	o.SetDefaults()

	cmd := &cobra.Command{
		Use:   "rbac",
		Short: "Generates RBAC manifests",
		Long: `Generate RBAC manifests from the RBAC annotations in Go source files.
Usage:
# rbac generate [--name manager] [--input-dir input_dir] [--output-dir output_dir]
`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rbac.Generate(o); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("RBAC manifests generated under '%s' directory\n", o.OutputDir)
		},
	}

	f := cmd.Flags()
	f.StringVar(&o.Name, "name", o.Name, "Name to be used as prefix in identifier for manifests")
	f.StringVar(&o.InputDir, "input-dir", o.InputDir, "input directory pointing to Go source files")
	f.StringVar(&o.OutputDir, "output-dir", o.OutputDir, "output directory where generated manifests will be saved.")

	return cmd
}

func newAllSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "runs all generators for a project",
		Long: `Run all available generators for a given project
Usage:
# controller-gen all
`,
		Run: func(cmd *cobra.Command, args []string) {
			// load default project for populating generator options
			// TODO: uncomment the following while enabling CRD
			// projectFile, err := loadProjectFile()
			// if err != nil {
			// 	log.Fatal(err)
			// }

			// RBAC generation
			rbacOptions := &rbac.ManifestOptions{}
			rbacOptions.SetDefaults()
			if err := rbac.Generate(rbacOptions); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("RBAC manifests generated under '%s' directory\n", rbacOptions.OutputDir)

			// Plug-in CRD generation
		},
	}
	return cmd
}

// func loadProjectFile() (*input.ProjectFile, error) {
// 	// Load the PROJECT file for default options
// 	if _, err := os.Stat("PROJECT"); os.IsNotExist(err) {
// 		return nil, fmt.Errorf("%s file missing", "PROJECT")
// 	}
// 	content, err := ioutil.ReadFile("PROJECT")
// 	if err != nil {
// 		return nil, fmt.Errorf("Error reading '%s' file %v", "PROJECT", err)
// 	}
// 	projectFile := input.ProjectFile{}
// 	err = yaml.Unmarshal(content, &projectFile)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error loading '%s' file %v", "PROJECT", err)
// 	}
// 	return &projectFile, nil
// }
