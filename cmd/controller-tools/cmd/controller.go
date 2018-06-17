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

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-tools/pkg/scaffold"
	"sigs.k8s.io/controller-tools/pkg/scaffold/controller"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
)

var ctrl *controller.Controller
var in *controller.AddController

// controllerCmd represents the controller command
var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Scaffold a Controller for some Kubernetes Resource",
	Long: `Scaffold a Controller for some Kubernetes Resource.

Resource may be either a core Resource or a custom Resource.
Core Resource must specify --resource-package k8s.io/api

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Example: `	# Create a Controller for the Frigate API
	controller-tools scaffold controller --group ship --version v1beta1 --kind Frigate

	# Create a Controller for Deployments
	controller-tools scaffold controller --group apps --version v1 --kind Deployment
`,
	Run: func(cmd *cobra.Command, args []string) {
		project.DieIfNoProject()

		s := &scaffold.Scaffold{}
		err := s.Execute(scaffold.Options{}, ctrl, in)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s Edit %s and to define your API implementation in Reconcile.\n",
			actionRequired, ctrl.Path())
	},
}

func init() {
	scaffoldCmd.AddCommand(controllerCmd)

	ctrl = controller.ForFlags(controllerCmd.Flags())
	in = &controller.AddController{Resource: ctrl.Resource}
}
