// Copyright 2019 Istio Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	gogoplugin "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/gogo/protobuf/vanity/command"

	"istio.io/tools/cmd/protoc-gen-deepcopy/deepcopy"
)

func main() {
	request := command.Read()

	plugin := deepcopy.NewPlugin()

	response := command.GeneratePlugin(request, plugin, deepcopy.FileNameSuffix)

	filterResponse(response, plugin.FilesWritten())

	command.Write(response)
}

func filterResponse(response *gogoplugin.CodeGeneratorResponse, written map[string]interface{}) {
	files := response.GetFile()
	filtered := make([]*gogoplugin.CodeGeneratorResponse_File, 0, len(files))
	for _, file := range files {
		if _, ok := written[file.GetName()]; ok {
			filtered = append(filtered, file)
		}
	}
	response.File = filtered
}
