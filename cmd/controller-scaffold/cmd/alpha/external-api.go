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

package alpha

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"sigs.k8s.io/controller-tools/pkg/alpha/externalapi/pkg/scaffold/controller"
	"sigs.k8s.io/controller-tools/pkg/scaffold"
	// "sigs.k8s.io/controller-tools/pkg/scaffold/externalapi"
	"sigs.k8s.io/controller-tools/pkg/alpha/externalapi/pkg/scaffold/resource"
	"sigs.k8s.io/controller-tools/pkg/scaffold/input"
)

var r *resource.Resource
var resourceFlag, controllerFlag *flag.Flag
var doResource, doController, doMake bool

// ExtAPICmd represents the resource command
var ExtAPICmd = &cobra.Command{
	Use:   "external-api",
	Short: "Scaffold a Kubernetes API",
	Long: `Scaffold a Kubernetes API by creating a Resource definition and / or a Controller.

api will prompt the user for if it should scaffold the Resource and / or Controller.  To only
scaffold a Controller for an existing Resource, select "n" for Resource.  To only define
the schema for a Resource without writing a Controller, select "n" for Controller.

After the scaffold is written, api will run make on the project.
`,
	Example: `	# Create a frigates API with Group: ship, Version: v1beta1 and Kind: Frigate
	controller-scaffold external-api --group ship --version v1beta1 --kind Frigate

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
		DieIfNoProject()

		if !resourceFlag.Changed {
			fmt.Println("Create Resource under pkg/apis [y/n]?")
			doResource = yesno()
		}
		if !controllerFlag.Changed {
			fmt.Println("Create Controller under pkg/controller [y/n]?")
			doController = yesno()
		}

		fmt.Println("Writing scaffold for you to edit...")

		if doResource {
			fmt.Println(filepath.Join("pkg", "apis", r.Group, r.Version,
				fmt.Sprintf("%s_types.go", strings.ToLower(r.Kind))))
			fmt.Println(filepath.Join("pkg", "apis", r.Group, r.Version,
				fmt.Sprintf("%s_types_test.go", strings.ToLower(r.Kind))))

			err := (&scaffold.Scaffold{}).Execute(input.Options{},
				&resource.AddToScheme{Resource: r},
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		if doController {
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

		if doMake {
			fmt.Println("Running make...")
			cm := exec.Command("make") // #nosec
			cm.Stderr = os.Stderr
			cm.Stdout = os.Stdout
			if err := cm.Run(); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	ExtAPICmd.Flags().BoolVar(&doMake, "make", true,
		"if true, run make after generating files")
	ExtAPICmd.Flags().BoolVar(&doResource, "resource", true,
		"if set, generate the resource without prompting the user")
	resourceFlag = ExtAPICmd.Flag("resource")
	ExtAPICmd.Flags().BoolVar(&doController, "controller", true,
		"if set, generate the controller without prompting the user")
	controllerFlag = ExtAPICmd.Flag("controller")
	r = resourceForFlags(ExtAPICmd.Flags())
}

// DieIfNoProject checks to make sure the command is run from a directory containing a project file.
func DieIfNoProject() {
	if _, err := os.Stat("PROJECT"); os.IsNotExist(err) {
		log.Fatalf("Command must be run from a diretory containing %s", "PROJECT")
	}
}

// resourceForFlags registers flags for Resource fields and returns the Resource
func resourceForFlags(f *flag.FlagSet) *resource.Resource {
	r := &resource.Resource{}
	f.StringVar(&r.Kind, "kind", "", "resource Kind")
	f.StringVar(&r.Group, "group", "", "resource Group")
	f.StringVar(&r.Version, "version", "", "resource Version")
	f.StringVar(&r.ImportPath, "importPath", "", "import path of external type")
	f.BoolVar(&r.Namespaced, "namespaced", true, "true if the resource is namespaced")
	f.BoolVar(&r.CreateExampleReconcileBody, "example", true,
		"true if an example reconcile body should be written")
	return r
}
