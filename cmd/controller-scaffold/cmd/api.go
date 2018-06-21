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

	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-tools/pkg/scaffold"
	"sigs.k8s.io/controller-tools/pkg/scaffold/controller"
	"sigs.k8s.io/controller-tools/pkg/scaffold/input"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
	"sigs.k8s.io/controller-tools/pkg/scaffold/resource"
)

var r *resource.Resource

// APICmd represents the resource command
var APICmd = &cobra.Command{
	Use:   "api",
	Short: "Scaffold a Kubernetes API",
	Long: `Scaffold a Kubernetes API by creating a Resource definition and / or a Controller.

api will prompt the user for if it should scaffold the Resource and / or Controller.  To only
scaffold a Controller for an existing Resource, select "n" for Resource.  To only define
the schema for a Resource without writing a Controller, select "n" for Controller.

After the scaffold is written, api will run make on the project.
`,
	Example: `	# Create a frigates API with Group: ship, Version: v1beta1 and Kind: Frigate
	controller-scaffold api --group ship --version v1beta1 --kind Frigate

	# Edit the API Scheme
	nano pkg/apis/ship/v1beta1/frigate_types.go

	# Edit the Controller
	nano pkg/controller/frigate/frigate_controller.go

	# Edit the Controller Test
	nano pkg/controller/frigate/frigate_controller_test.go

	# Install CRDs into the Kubernetes cluster using kubectl apply
	make install

	# Regenerate code and run against the Kubernetes cluster configured by ~/.kube/config
	make run
`,
	Run: func(cmd *cobra.Command, args []string) {
		project.DieIfNoProject()

		fmt.Println("Create Resource under pkg/apis [y/n]?")
		re := yesno()
		fmt.Println("Create Controller under pkg/controller [y/n]?")
		c := yesno()

		fmt.Println("Writing scaffold for you to edit...")

		if re {
			fmt.Println(filepath.Join("pkg", "apis", r.Group, r.Version,
				fmt.Sprintf("%s_types.go", strings.ToLower(r.Kind))))
			fmt.Println(filepath.Join("pkg", "apis", r.Group, r.Version,
				fmt.Sprintf("%s_types_test.go", strings.ToLower(r.Kind))))

			err := (&scaffold.Scaffold{}).Execute(input.Options{},
				&resource.Register{Resource: r},
				&resource.Types{Resource: r},
				&resource.VersionSuiteTest{Resource: r},
				&resource.TypesTest{Resource: r},
				&resource.Doc{Resource: r},
				&resource.Group{Resource: r},
				&resource.AddToScheme{Resource: r},
				&resource.CRD{Resource: r},
				&resource.Role{Resource: r},
				&resource.RoleBinding{Resource: r},
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		if c {
			fmt.Println(filepath.Join("pkg", "controller", strings.ToLower(r.Kind),
				fmt.Sprintf("%s_controller.go", strings.ToLower(r.Kind))))
			fmt.Println(filepath.Join("pkg", "apis", strings.ToLower(r.Kind),
				fmt.Sprintf("%s_controller_test.go", strings.ToLower(r.Kind))))

			err := (&scaffold.Scaffold{}).Execute(input.Options{},
				&controller.Controller{Resource: r},
				&controller.AddController{Resource: r},
				&controller.Test{Resource: r},
				&controller.SuiteTest{Resource: r},
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		fmt.Println("Running make...")
		cm := exec.Command("make") // #nosec
		cm.Stderr = os.Stderr
		cm.Stdout = os.Stdout
		if err := cm.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(APICmd)
	r = resource.ForFlags(APICmd.Flags())
}
