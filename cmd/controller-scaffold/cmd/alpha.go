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

package cmd

import (
	// "fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-tools/cmd/controller-scaffold/cmd/alpha"
)

// AlphaCmd represents the alpha command
var AlphaCmd = &cobra.Command{
	Use:   "alpha",
	Short: "Experimental scaffold commmands",
	Long: `Commands that are not ready for prime time. For example:

	controller-scaffold alpha list
`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("alpha called")
	// },
}

func init() {
	rootCmd.AddCommand(AlphaCmd)
	AlphaCmd.AddCommand(alpha.ListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// AlphaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// AlphaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
