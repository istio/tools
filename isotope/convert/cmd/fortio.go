// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
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
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"istio.io/tools/isotope/convert/pkg/kubernetes"
)

// kubernetesCmd represents the kubernetes command
var fortioCmd = &cobra.Command{
	Use:   "fortio [application.yaml]",
	Short: "Generate fortio client manifests for performance testing",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inPath := args[0]

		clientNodeSelectorStr, err := cmd.PersistentFlags().GetString("client-node-selector")
		exitIfError(err)
		clientNodeSelector, err := extractNodeSelector(clientNodeSelectorStr)
		exitIfError(err)

		clientImage, err := cmd.PersistentFlags().GetString("client-image")
		exitIfError(err)

		files, err := ioutil.ReadDir(inPath)
		exitIfError(err)

		manifests, err := kubernetes.GenerateFortioManifests(
			clientNodeSelector, clientImage)
		exitIfError(err)

		var result strings.Builder
		result.Write(manifests)
		result.WriteString("\n---\n")
		for _, file := range files {
			yamlContents, err := ioutil.ReadFile(inPath + "/" + file.Name())
			exitIfError(err)
			result.Write(yamlContents)
			result.WriteString("\n---\n")
		}

		fmt.Println(result.String())
	},
}

func init() {
	rootCmd.AddCommand(fortioCmd)
	fortioCmd.PersistentFlags().String(
		"client-image", "", "the image to use for the load testing client job")
	fortioCmd.PersistentFlags().String(
		"client-node-selector", "", "the node selector for client workloads")
}
