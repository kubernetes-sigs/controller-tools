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
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-tools/pkg/scaffold"
	"sigs.k8s.io/controller-tools/pkg/scaffold/controller"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
	"sigs.k8s.io/controller-tools/pkg/scaffold/resource"
)

var r *resource.Resource

// resourceCmd represents the resource command
var resourceCmd = &cobra.Command{
	Use:   "api",
	Short: "Scaffold a Kubernetes API",
	Long: `Scaffold a Kubernetes API by creating a Resource definition and / or a Controller.

When run, api will prompt the user for whether it should scaffold the Resource / Controller.  Once
the scaffold is written, api will run make to generate code, run go tests, and build the manager
command.
`,
	Example: `	# Create a frigates API with Group: ship, Version: v1beta1 and Kind: Frigate
	controller-tools api --group ship --version v1beta1 --kind Frigate

	# Run code generators, test and build.
	make

	# Install CRDs into the Kubernetes cluster
	kubectl apply -f config/crds

	# Run against the Kubernetes cluster
	./bin/manager
`,
	Run: func(cmd *cobra.Command, args []string) {
		project.DieIfNoProject()

		fmt.Println("Create Resource under pkg/apis [y/n]?")
		re := yesno()
		fmt.Println("Create Controller under pkg/controller [y/n]?")
		c := yesno()

		fmt.Println("Writing scaffold for you to edit...")

		if re {
			err := (&scaffold.Scaffold{}).Execute(scaffold.Options{},
				&resource.RegisterGo{Resource: r},
				&resource.Types{Resource: r},
				&resource.VersionSuiteTest{Resource: r},
				&resource.TypesTest{Resource: r},
				&resource.APIsDocGo{Resource: r},
				&resource.Group{Resource: r},
				&resource.AddResource{Resource: r},
				&resource.CRD{Resource: r},
				&resource.Role{Resource: r},
				&resource.RoleBinding{Resource: r},
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		if c {
			ctrl := &controller.Controller{Resource: r}
			in := &controller.AddController{Resource: r}
			err := (&scaffold.Scaffold{}).Execute(scaffold.Options{}, ctrl, in)
			if err != nil {
				log.Fatal(err)
			}
		}

		fmt.Println("Running make...")
		cm := exec.Command("make")
		cm.Stderr = os.Stderr
		cm.Stdout = os.Stdout
		if err := cm.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(resourceCmd)
	r = resource.ForFlags(resourceCmd.Flags())
}
