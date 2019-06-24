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

type EnumDescriptor struct {
	baseDesc
	*descriptor.EnumDescriptorProto
	Values []*EnumValueDescriptor // The values of this enum
}

type EnumValueDescriptor struct {
	baseDesc
	*descriptor.EnumValueDescriptorProto
}

func newEnumDescriptor(desc *descriptor.EnumDescriptorProto, parent *MessageDescriptor, file *FileDescriptor, path pathVector) *EnumDescriptor {
	var qualifiedName []string
	if parent == nil {
		qualifiedName = []string{desc.GetName()}
	} else {
		qualifiedName = make([]string, len(parent.QualifiedName()), len(parent.QualifiedName())+1)
		copy(qualifiedName, parent.QualifiedName())
		qualifiedName = append(qualifiedName, desc.GetName())
	}

	e := &EnumDescriptor{
		EnumDescriptorProto: desc,
		baseDesc:            newBaseDesc(file, path, qualifiedName),
	}

	e.Values = make([]*EnumValueDescriptor, 0, len(desc.Value))
	for i, ev := range desc.Value {
		nameCopy := make([]string, len(qualifiedName), len(qualifiedName)+1)
		copy(nameCopy, qualifiedName)
		nameCopy = append(nameCopy, ev.GetName())

		evd := &EnumValueDescriptor{
			EnumValueDescriptorProto: ev,
			baseDesc:                 newBaseDesc(file, path.append(enumValuePath, i), nameCopy),
		}
		e.Values = append(e.Values, evd)
	}

	return e
}
