package pmodel

import (
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// Model represents the fully-read and indexed Proto Model
type Model struct {
	Request *plugin.CodeGeneratorRequest

	// MessageTypes, indexed by fully-qualified name
	MessageTypes map[string]*descriptor.DescriptorProto
}

func New(request *plugin.CodeGeneratorRequest) *Model {
	m := &Model{
		Request:      request,
		MessageTypes: make(map[string]*descriptor.DescriptorProto),
	}

	for _, pf := range request.ProtoFile {
		for _, msg := range pf.MessageType {
			name := asFQN(pf.Package, msg.Name)
			m.MessageTypes[name] = msg
		}
	}

	return m
}

func asFQN(pn *string, tn *string) string {
	result := "."
	if pn != nil {
		result += *pn + "."
	}
	result += *tn

	return result
}

func (m *Model) FindType(name string) *descriptor.DescriptorProto {
	return m.MessageTypes[name]
}
