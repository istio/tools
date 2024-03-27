// Copyright Istio Authors
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

package main

import (
	"fmt"
	"strings"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"

	"istio.io/tools/cmd/protoc-gen-crd/pkg/protocgen"
	"istio.io/tools/pkg/protomodel"
)

const (
	extendedChannelFileName = "kubernetes/extended.gen.yaml"
	// Only GA features are included in the stable channel
	stableChannelFileName = "kubernetes/stable.gen.yaml"
)

// Breaks the comma-separated list of key=value pairs
// in the parameter string into an easy to use map.
func extractParams(parameter string) map[string]string {
	m := make(map[string]string)
	for _, p := range strings.Split(parameter, ",") {
		if p == "" {
			continue
		}

		if i := strings.Index(p, "="); i < 0 {
			m[p] = ""
		} else {
			m[p[0:i]] = p[i+1:]
		}
	}

	return m
}

func generate(request *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	includeDescription := true
	enumAsIntOrString := false
	type genMetadata struct {
		shouldGen       bool
		includeExtended bool
		fds             []*protomodel.FileDescriptor
	}

	p := extractParams(request.GetParameter())
	for k, v := range p {
		if k == "include_description" {
			switch strings.ToLower(v) {
			case "true":
				includeDescription = true
			case "false":
				includeDescription = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for include_description", v)
			}
		} else if k == "enum_as_int_or_string" {
			switch strings.ToLower(v) {
			case "true":
				enumAsIntOrString = true
			case "false":
				enumAsIntOrString = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for enum_as_int_or_string", v)
			}
		} else {
			return nil, fmt.Errorf("unknown argument '%s' specified", k)
		}
	}

	m := protomodel.NewModel(request, false)
	channelOutput := map[string]*genMetadata{
		extendedChannelFileName: {
			shouldGen:       true,
			includeExtended: true,
			fds:             make([]*protomodel.FileDescriptor, 0),
		},
		stableChannelFileName: {
			shouldGen:       true,
			includeExtended: false,
			fds:             make([]*protomodel.FileDescriptor, 0),
		},
	}

	for _, fileName := range request.FileToGenerate {
		fd := m.AllFilesByName[fileName]
		if fd == nil {
			return nil, fmt.Errorf("unable to find %s", request.FileToGenerate)
		}

		channelOutput[extendedChannelFileName].fds = append(channelOutput[extendedChannelFileName].fds, fd)
		// We'll later remove the files from the stable channel that are experimental
		channelOutput[stableChannelFileName].fds = append(channelOutput[stableChannelFileName].fds, fd)
	}

	descriptionConfiguration := &DescriptionConfiguration{
		IncludeDescriptionInSchema: includeDescription,
	}

	response := plugin.CodeGeneratorResponse{}
	for outputFileName, meta := range channelOutput {
		meta := meta
		g := newOpenAPIGenerator(
			m,
			descriptionConfiguration,
			enumAsIntOrString,
			meta.includeExtended,
		)
		filesToGen := map[*protomodel.FileDescriptor]bool{}
		for _, fd := range meta.fds {
			// All files will be generated in each channel, as shouldGen is true for all
			filesToGen[fd] = meta.shouldGen
		}
		rf := g.generateSingleFileOutput(filesToGen, outputFileName, meta.includeExtended)
		response.File = append(response.File, &rf)
	}

	return &response, nil
}

func main() {
	protocgen.Generate(generate)
}
