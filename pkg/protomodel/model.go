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

package protomodel

import (
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// model represents a resolved in-memory version of all the input protos
type Model struct {
	AllFilesByName map[string]*FileDescriptor
	AllDescByName  map[string]CoreDesc
	Packages       []*PackageDescriptor
}

func NewModel(request *plugin.CodeGeneratorRequest, perFile bool) *Model {
	m := &Model{
		AllFilesByName: make(map[string]*FileDescriptor, len(request.ProtoFile)),
	}

	// organize files by package
	filesByPackage := map[string][]*descriptor.FileDescriptorProto{}
	for _, pf := range request.ProtoFile {
		pkg := packageName(pf)
		slice := filesByPackage[pkg]
		filesByPackage[pkg] = append(slice, pf)
	}

	// create all the package descriptors
	var allFiles []*FileDescriptor
	for pkg, files := range filesByPackage {
		p := newPackageDescriptor(pkg, files, perFile)
		m.Packages = append(m.Packages, p)

		for _, f := range p.Files {
			allFiles = append(allFiles, f)
			m.AllFilesByName[f.GetName()] = f
		}
	}

	// prepare a map of name to descriptor
	m.AllDescByName = createDescMap(allFiles)

	// resolve all type references to nice easily used pointers
	for _, f := range allFiles {
		resolveFieldTypes(f.Messages, m.AllDescByName)
		resolveMethodTypes(f.Services, m.AllDescByName)
		resolveDependencies(f, m.AllFilesByName)
	}

	return m
}

func packageName(f *descriptor.FileDescriptorProto) string {
	// Does the file have a package clause?
	if pkg := f.GetPackage(); pkg != "" {
		return pkg
	}

	// use the last path element of the name, with the last dotted suffix removed.

	// First, find the last element
	name := f.GetName()
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}

	// Now drop the suffix
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[0:i]
	}

	return name
}

// createDescMap builds a map from qualified names to descriptors.
// The key names for the map come from the input data, which puts a period at the beginning.
func createDescMap(files []*FileDescriptor) map[string]CoreDesc {
	descMap := make(map[string]CoreDesc)
	for _, f := range files {
		// The names in this loop are defined by the proto world, not us, so the
		// package name may be empty.  If so, the dotted package name of X will
		// be ".X"; otherwise it will be ".pkg.X".
		dottedPkg := "." + f.GetPackage()
		if dottedPkg != "." {
			dottedPkg += "."
		}

		for _, svc := range f.Services {
			descMap[dottedPkg+DottedName(svc)] = svc
		}

		recordEnums(f.Enums, descMap, dottedPkg)
		recordMessages(f.Messages, descMap, dottedPkg)
		recordServices(f.Services, descMap, dottedPkg)
		resolveFieldTypes(f.Messages, descMap)
	}

	return descMap
}

func recordMessages(messages []*MessageDescriptor, descMap map[string]CoreDesc, dottedPkg string) {
	for _, msg := range messages {
		descMap[dottedPkg+DottedName(msg)] = msg

		recordMessages(msg.Messages, descMap, dottedPkg)
		recordEnums(msg.Enums, descMap, dottedPkg)

		for _, f := range msg.Fields {
			descMap[dottedPkg+DottedName(f)] = f
		}
	}
}

func recordEnums(enums []*EnumDescriptor, descMap map[string]CoreDesc, dottedPkg string) {
	for _, e := range enums {
		descMap[dottedPkg+DottedName(e)] = e

		for _, v := range e.Values {
			descMap[dottedPkg+DottedName(v)] = v
		}
	}
}

func recordServices(services []*ServiceDescriptor, descMap map[string]CoreDesc, dottedPkg string) {
	for _, s := range services {
		descMap[dottedPkg+DottedName(s)] = s

		for _, m := range s.Methods {
			descMap[dottedPkg+DottedName(m)] = m
		}
	}
}

func resolveFieldTypes(messages []*MessageDescriptor, descMap map[string]CoreDesc) {
	for _, msg := range messages {
		for _, field := range msg.Fields {
			field.FieldType = descMap[field.GetTypeName()]
		}
		resolveFieldTypes(msg.Messages, descMap)
	}
}

func resolveMethodTypes(services []*ServiceDescriptor, descMap map[string]CoreDesc) {
	for _, svc := range services {
		for _, method := range svc.Methods {
			method.Input = descMap[method.GetInputType()].(*MessageDescriptor)
			method.Output = descMap[method.GetOutputType()].(*MessageDescriptor)
		}
	}
}

func resolveDependencies(file *FileDescriptor, filesByName map[string]*FileDescriptor) {
	for _, desc := range file.Dependency {
		dep := filesByName[desc]
		file.Dependencies = append(file.Dependencies, dep)
	}
}

// DottedName returns a dotted representation of the coreDesc's name
func DottedName(o CoreDesc) string {
	return strings.Join(o.QualifiedName(), ".")
}
