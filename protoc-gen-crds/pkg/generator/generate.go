// Copyright 2018 Istio Authors
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

package generator

import (
	"path"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"istio.io/tools/protoc-gen-crds/pkg/crd"
	"istio.io/tools/protoc-gen-crds/pkg/crd/emitter/golang"
	"istio.io/tools/protoc-gen-crds/pkg/pmodel"
)

// Generate implements a protoc plugin.
func Generate(request plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {

	m := pmodel.New(&request)

	b := newCRDBuilder(m)
	if err := b.build(); err != nil {
		return errorResponse(err), nil
	}

	response := &plugin.CodeGeneratorResponse{}
	crds := crd.ToString(b.definitions)
	addFile(response, "crd.yaml", crds)

	for _, defn := range b.definitions {

		code, err := golang.Emit(defn)
		if err != nil {
			return errorResponse(err), nil
		}
		addFile(response, getOutputPath(defn), code)
	}

	return response, nil
}

func addFile(r *plugin.CodeGeneratorResponse, name, content string) {
	f := &plugin.CodeGeneratorResponse_File{
		Name:    &name,
		Content: &content,
	}

	r.File = append(r.File, f)
}

func errorResponse(err error) *plugin.CodeGeneratorResponse {
	e := err.Error()
	return &plugin.CodeGeneratorResponse{
		Error: &e,
	}
}

func getOutputPath(d *crd.ResourceDefinition) string {
	return path.Join(d.Spec.Group, d.Spec.Version, d.Spec.Names.Kind+".go")
}
