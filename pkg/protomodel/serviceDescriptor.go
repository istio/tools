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
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type ServiceDescriptor struct {
	baseDesc
	*descriptor.ServiceDescriptorProto
	Methods []*MethodDescriptor // Methods, if any
}

type MethodDescriptor struct {
	baseDesc
	*descriptor.MethodDescriptorProto
	Input  *MessageDescriptor
	Output *MessageDescriptor
}

func newServiceDescriptor(desc *descriptor.ServiceDescriptorProto, file *FileDescriptor, path pathVector) *ServiceDescriptor {
	qualifiedName := []string{desc.GetName()}

	s := &ServiceDescriptor{
		ServiceDescriptorProto: desc,
		baseDesc:               newBaseDesc(file, path, qualifiedName),
	}

	for i, m := range desc.Method {
		nameCopy := make([]string, len(qualifiedName), len(qualifiedName)+1)
		copy(nameCopy, qualifiedName)
		nameCopy = append(nameCopy, m.GetName())

		md := &MethodDescriptor{
			MethodDescriptorProto: m,
			baseDesc:              newBaseDesc(file, path.append(serviceMethodPath, i), nameCopy),
		}
		s.Methods = append(s.Methods, md)
	}

	return s
}
