package main

import (
	gogoplugin "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/gogo/protobuf/vanity/command"

	"istio.io/tools/cmd/protoc-gen-jsonshim/jsonshim"
)

func main() {
	request := command.Read()

	plugin := jsonshim.NewPlugin()

	response := command.GeneratePlugin(request, plugin, jsonshim.FileNameSuffix)

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
