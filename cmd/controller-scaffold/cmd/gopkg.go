// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-tools/pkg/scaffold"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
)

var gopkglocal *project.GopkgToml

// gopkgCmd represents the gopkg command
var gopkgCmd = &cobra.Command{
	Use:   "gopkg",
	Short: "Write or update a managed Gopkg.toml with base Kubernetes transitive deps",
	Long: `Write or update a managed Gopkg.toml with base Kubernetes transitive deps.
Will fail if the Gopkg.toml exists and is unmanaged.  Will keep use changes
made above the delimiter.`,
	Example: `controller-scaffold gopkg`,
	Run: func(cmd *cobra.Command, args []string) {
		s := &scaffold.Scaffold{
			BoilerplateOptional: true,
			ProjectOptional:     true,
		}
		err := s.Execute(scaffold.Options{
			ProjectPath:     prj.Path(),
			BoilerplatePath: bp.Path(),
		}, gopkg, mrg, dkr)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(gopkgCmd)
	gopkglocal = project.GopkgTomlForFlags(projectCmd.Flags())
}
