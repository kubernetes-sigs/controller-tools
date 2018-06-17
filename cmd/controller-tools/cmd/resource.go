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

	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-tools/pkg/scaffold"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
	"sigs.k8s.io/controller-tools/pkg/scaffold/resource"
)

var r *resource.Resource

// resourceCmd represents the resource command
var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Scaffold a Kubernetes API",
	Long:  `Scaffold a Kubernetes API.`,
	Example: `	# Create a frigates resource with Group: ship, Version: v1beta1 and Kind: Frigate
	controller-tools scaffold resource --group ship --version v1beta1 --kind Frigate

	# Generate the deepcopy code required by the Resource.  Re-run this anytime the Resource is changed.
	go generate ./pkg/...
`,
	Run: func(cmd *cobra.Command, args []string) {
		project.DieIfNoProject()

		t := &resource.Types{Resource: r}
		err := (&scaffold.Scaffold{}).Execute(scaffold.Options{},
			&resource.RegisterGo{Resource: r},
			t,
			&resource.TypesTest{Resource: r},
			&resource.APIsDocGo{Resource: r},
			&resource.Group{Resource: r},
			&resource.AddResource{Resource: r},
		)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Run `go generate ./pkg/apis/...` [y/n]?")
		if yesno() {
			c := exec.Command("go", "generate", "./pkg/apis/...")
			c.Stderr = os.Stderr
			c.Stdout = os.Stdout
			fmt.Println(strings.Join(c.Args, " "))
			if err := c.Run(); err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("Skipping `go generate ./pkg/apis/......`.  Deepcopy will not be generated.")
		}

		fmt.Println("Note: You must run `go generate ./pkg/...` anytime you change files under `pkg/apis/`")
		fmt.Printf("%s Edit %s to define your API Scheme in the %sSpec' and '%sStatus'\n",
			actionRequired, t.Path(), t.Kind, t.Kind)
	},
}

func init() {
	scaffoldCmd.AddCommand(resourceCmd)
	r = resource.ForFlags(resourceCmd.Flags())
}
