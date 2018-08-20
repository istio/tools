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

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"istio.io/tools/isotope/convert/pkg/graph"
	"istio.io/tools/isotope/convert/pkg/kubernetes"
)

// kubernetesCmd represents the kubernetes command
var kubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Convert service graph YAML to manifests for performance testing",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		inPath := args[0]
		outPath := args[1]

		serviceNodeSelectorStr := args[2]
		serviceNodeSelector, err := extractNodeSelector(
			serviceNodeSelectorStr)
		exitIfError(err)

		clientNodeSelectorStr := args[3]
		clientNodeSelector, err := extractNodeSelector(clientNodeSelectorStr)
		exitIfError(err)

		serviceImage, err := cmd.PersistentFlags().GetString("service-image")
		exitIfError(err)

		serviceMaxIdleConnectionsPerHost, err :=
			cmd.PersistentFlags().GetInt("service-max-idle-connections-per-host")
		exitIfError(err)

		clientImage, err := cmd.PersistentFlags().GetString("client-image")
		exitIfError(err)

		yamlContents, err := ioutil.ReadFile(inPath)
		exitIfError(err)

		var serviceGraph graph.ServiceGraph
		exitIfError(yaml.Unmarshal(yamlContents, &serviceGraph))

		manifests, err := kubernetes.ServiceGraphToKubernetesManifests(
			serviceGraph, serviceNodeSelector, serviceImage,
			serviceMaxIdleConnectionsPerHost, clientNodeSelector, clientImage)
		exitIfError(err)

		exitIfError(writeManifest(outPath, manifests))
	},
}

func init() {
	rootCmd.AddCommand(kubernetesCmd)
	kubernetesCmd.PersistentFlags().String(
		"service-image", "", "the image to deploy for all services in the graph")
	kubernetesCmd.PersistentFlags().Int(
		"service-max-idle-connections-per-host", 0,
		"maximum number of connections to keep open per host on each service")
	kubernetesCmd.PersistentFlags().String(
		"client-image", "", "the image to use for the load testing client job")
}

func writeManifest(path string, manifest []byte) error {
	// 0644 allows owner to read/write and others to read.
	return ioutil.WriteFile(path, manifest, 0644)
}

func splitByEquals(s string) (k string, v string, err error) {
	parts := strings.Split(s, "=")
	if len(parts) != 2 {
		err = fmt.Errorf("%s is not a valid node selector", s)
		return
	}
	k = parts[0]
	v = parts[1]
	return
}

func extractNodeSelector(s string) (map[string]string, error) {
	nodeSelector := make(map[string]string, 1)
	k, v, err := splitByEquals(s)
	if err != nil {
		return nodeSelector, err
	}
	nodeSelector[k] = v
	return nodeSelector, nil
}
