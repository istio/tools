// Code generated by protoc-gen-go. DO NOT EDIT.
// source: kubernetes/resource/options.proto

package resource // import "istio.io/tools/kubernetes/resource"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Scope int32

const (
	Scope_NAMESPACED Scope = 0
	Scope_CLUSTER    Scope = 1
	Scope_ALL        Scope = 2
)

var Scope_name = map[int32]string{
	0: "NAMESPACED",
	1: "CLUSTER",
	2: "ALL",
}
var Scope_value = map[string]int32{
	"NAMESPACED": 0,
	"CLUSTER":    1,
	"ALL":        2,
}

func (x Scope) Enum() *Scope {
	p := new(Scope)
	*p = x
	return p
}
func (x Scope) String() string {
	return proto.EnumName(Scope_name, int32(x))
}
func (x *Scope) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(Scope_value, data, "Scope")
	if err != nil {
		return err
	}
	*x = Scope(value)
	return nil
}
func (Scope) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_options_46995d5f99e089a8, []int{0}
}

var E_Spec = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MessageOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         50000,
	Name:          "kubernetes.resource.spec",
	Tag:           "varint,50000,opt,name=spec",
	Filename:      "kubernetes/resource/options.proto",
}

var E_Scope = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MessageOptions)(nil),
	ExtensionType: (*Scope)(nil),
	Field:         50001,
	Name:          "kubernetes.resource.scope",
	Tag:           "varint,50001,opt,name=scope,enum=kubernetes.resource.Scope",
	Filename:      "kubernetes/resource/options.proto",
}

var E_Group = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MessageOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         50002,
	Name:          "kubernetes.resource.group",
	Tag:           "bytes,50002,opt,name=group",
	Filename:      "kubernetes/resource/options.proto",
}

var E_Version = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MessageOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         50003,
	Name:          "kubernetes.resource.version",
	Tag:           "bytes,50003,opt,name=version",
	Filename:      "kubernetes/resource/options.proto",
}

var E_Kind = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MessageOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         50004,
	Name:          "kubernetes.resource.kind",
	Tag:           "bytes,50004,opt,name=kind",
	Filename:      "kubernetes/resource/options.proto",
}

var E_Singular = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MessageOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         50005,
	Name:          "kubernetes.resource.singular",
	Tag:           "bytes,50005,opt,name=singular",
	Filename:      "kubernetes/resource/options.proto",
}

var E_Plural = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MessageOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         50006,
	Name:          "kubernetes.resource.plural",
	Tag:           "bytes,50006,opt,name=plural",
	Filename:      "kubernetes/resource/options.proto",
}

func init() {
	proto.RegisterEnum("kubernetes.resource.Scope", Scope_name, Scope_value)
	proto.RegisterExtension(E_Spec)
	proto.RegisterExtension(E_Scope)
	proto.RegisterExtension(E_Group)
	proto.RegisterExtension(E_Version)
	proto.RegisterExtension(E_Kind)
	proto.RegisterExtension(E_Singular)
	proto.RegisterExtension(E_Plural)
}

func init() {
	proto.RegisterFile("kubernetes/resource/options.proto", fileDescriptor_options_46995d5f99e089a8)
}

var fileDescriptor_options_46995d5f99e089a8 = []byte{
	// 300 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0xd2, 0x4f, 0x4b, 0xf3, 0x40,
	0x10, 0xc7, 0xf1, 0xa7, 0x4f, 0xad, 0xad, 0x23, 0x94, 0x12, 0x2f, 0xd2, 0x8b, 0x55, 0x3c, 0x88,
	0xe0, 0x06, 0x04, 0x11, 0x23, 0x1e, 0x6a, 0xed, 0xad, 0x55, 0x49, 0xf4, 0xe2, 0xad, 0x4d, 0xc7,
	0xb0, 0x34, 0x64, 0x96, 0x9d, 0x5d, 0x5f, 0x82, 0xaf, 0xcf, 0xff, 0xaf, 0x47, 0xdc, 0x24, 0x7a,
	0x29, 0xec, 0xfd, 0xfb, 0xf9, 0x65, 0x12, 0x02, 0xbb, 0x4b, 0x3b, 0x47, 0x5d, 0xa0, 0x41, 0x0e,
	0x35, 0x32, 0x59, 0x9d, 0x62, 0x48, 0xca, 0x48, 0x2a, 0x58, 0x28, 0x4d, 0x86, 0x82, 0xad, 0xbf,
	0x44, 0xd4, 0x49, 0x7f, 0x90, 0x11, 0x65, 0x39, 0x86, 0x2e, 0x99, 0xdb, 0xc7, 0x70, 0x81, 0x9c,
	0x6a, 0xa9, 0x0c, 0xe9, 0x92, 0x1d, 0x1e, 0x41, 0x2b, 0x49, 0x49, 0x61, 0xd0, 0x05, 0xb8, 0x1e,
	0x4e, 0xc7, 0xc9, 0xed, 0x70, 0x34, 0xbe, 0xea, 0xfd, 0x0b, 0x36, 0xa1, 0x3d, 0x9a, 0xdc, 0x27,
	0x77, 0xe3, 0xb8, 0xd7, 0x08, 0xda, 0xd0, 0x1c, 0x4e, 0x26, 0xbd, 0xff, 0xd1, 0x09, 0xac, 0xb1,
	0xc2, 0x34, 0xd8, 0x11, 0xe5, 0xb2, 0xa8, 0x97, 0xc5, 0x14, 0x99, 0x67, 0x19, 0xde, 0x94, 0x47,
	0x6d, 0xbf, 0x3c, 0x37, 0x07, 0x8d, 0x83, 0x4e, 0xec, 0xf2, 0x28, 0x81, 0x16, 0xbb, 0xa7, 0x78,
	0xdd, 0xab, 0x73, 0xdd, 0xe3, 0xbe, 0x58, 0xf1, 0x3e, 0xc2, 0x9d, 0x1a, 0x97, 0x5b, 0xd1, 0x29,
	0xb4, 0x32, 0x4d, 0x56, 0xf9, 0x47, 0xdf, 0xdc, 0xe8, 0x46, 0x5c, 0xf6, 0xd1, 0x39, 0xb4, 0x9f,
	0x50, 0xb3, 0xa4, 0xc2, 0x4f, 0xdf, 0x2b, 0x5a, 0x8b, 0x9f, 0x2f, 0xb0, 0x94, 0xc5, 0xc2, 0x2f,
	0x3f, 0x2a, 0xe9, 0xf2, 0xe8, 0x02, 0x3a, 0x2c, 0x8b, 0xcc, 0xe6, 0x33, 0xed, 0xa7, 0x9f, 0x15,
	0xfd, 0x25, 0xd1, 0x19, 0xac, 0xab, 0xdc, 0xea, 0x59, 0xee, 0xc7, 0x5f, 0x15, 0xae, 0xc0, 0xe5,
	0xfe, 0xc3, 0x9e, 0x64, 0x23, 0x49, 0x48, 0x0a, 0x0d, 0x51, 0xce, 0xe1, 0x8a, 0x9f, 0xe9, 0x3b,
	0x00, 0x00, 0xff, 0xff, 0x20, 0xe4, 0x29, 0x12, 0x62, 0x02, 0x00, 0x00,
}
