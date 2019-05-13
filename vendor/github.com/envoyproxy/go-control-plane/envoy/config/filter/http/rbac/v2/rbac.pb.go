// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: envoy/config/filter/http/rbac/v2/rbac.proto

package v2

import (
	fmt "fmt"
	io "io"
	math "math"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/lyft/protoc-gen-validate/validate"

	v2alpha "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2alpha"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// RBAC filter config.
type RBAC struct {
	// Specify the RBAC rules to be applied globally.
	// If absent, no enforcing RBAC policy will be applied.
	Rules *v2alpha.RBAC `protobuf:"bytes,1,opt,name=rules,proto3" json:"rules,omitempty"`
	// Shadow rules are not enforced by the filter (i.e., returning a 403)
	// but will emit stats and logs and can be used for rule testing.
	// If absent, no shadow RBAC policy will be applied.
	ShadowRules          *v2alpha.RBAC `protobuf:"bytes,2,opt,name=shadow_rules,json=shadowRules,proto3" json:"shadow_rules,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *RBAC) Reset()         { *m = RBAC{} }
func (m *RBAC) String() string { return proto.CompactTextString(m) }
func (*RBAC) ProtoMessage()    {}
func (*RBAC) Descriptor() ([]byte, []int) {
	return fileDescriptor_15d628c6558085a7, []int{0}
}
func (m *RBAC) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RBAC) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RBAC.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RBAC) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RBAC.Merge(m, src)
}
func (m *RBAC) XXX_Size() int {
	return m.Size()
}
func (m *RBAC) XXX_DiscardUnknown() {
	xxx_messageInfo_RBAC.DiscardUnknown(m)
}

var xxx_messageInfo_RBAC proto.InternalMessageInfo

func (m *RBAC) GetRules() *v2alpha.RBAC {
	if m != nil {
		return m.Rules
	}
	return nil
}

func (m *RBAC) GetShadowRules() *v2alpha.RBAC {
	if m != nil {
		return m.ShadowRules
	}
	return nil
}

type RBACPerRoute struct {
	// Override the global configuration of the filter with this new config.
	// If absent, the global RBAC policy will be disabled for this route.
	Rbac                 *RBAC    `protobuf:"bytes,2,opt,name=rbac,proto3" json:"rbac,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RBACPerRoute) Reset()         { *m = RBACPerRoute{} }
func (m *RBACPerRoute) String() string { return proto.CompactTextString(m) }
func (*RBACPerRoute) ProtoMessage()    {}
func (*RBACPerRoute) Descriptor() ([]byte, []int) {
	return fileDescriptor_15d628c6558085a7, []int{1}
}
func (m *RBACPerRoute) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RBACPerRoute) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RBACPerRoute.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RBACPerRoute) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RBACPerRoute.Merge(m, src)
}
func (m *RBACPerRoute) XXX_Size() int {
	return m.Size()
}
func (m *RBACPerRoute) XXX_DiscardUnknown() {
	xxx_messageInfo_RBACPerRoute.DiscardUnknown(m)
}

var xxx_messageInfo_RBACPerRoute proto.InternalMessageInfo

func (m *RBACPerRoute) GetRbac() *RBAC {
	if m != nil {
		return m.Rbac
	}
	return nil
}

func init() {
	proto.RegisterType((*RBAC)(nil), "envoy.config.filter.http.rbac.v2.RBAC")
	proto.RegisterType((*RBACPerRoute)(nil), "envoy.config.filter.http.rbac.v2.RBACPerRoute")
}

func init() {
	proto.RegisterFile("envoy/config/filter/http/rbac/v2/rbac.proto", fileDescriptor_15d628c6558085a7)
}

var fileDescriptor_15d628c6558085a7 = []byte{
	// 279 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0x41, 0x4b, 0xc3, 0x30,
	0x14, 0xc7, 0x49, 0xa9, 0x32, 0xb3, 0x1d, 0xa4, 0x08, 0xca, 0x0e, 0x75, 0x0c, 0x11, 0x41, 0x48,
	0xa0, 0xe2, 0xc5, 0x9b, 0xf5, 0x26, 0x08, 0x25, 0xc7, 0x5d, 0x24, 0x6d, 0xb3, 0x36, 0x10, 0xf6,
	0x4a, 0x9a, 0x45, 0x77, 0xf4, 0xdb, 0x79, 0xf4, 0x23, 0x48, 0x3f, 0x89, 0x24, 0xa9, 0xe0, 0x4e,
	0xf3, 0xd4, 0xc7, 0xeb, 0xef, 0xff, 0xcb, 0xe3, 0x8f, 0x6f, 0xc5, 0xc6, 0xc2, 0x8e, 0x56, 0xb0,
	0x59, 0xcb, 0x86, 0xae, 0xa5, 0x32, 0x42, 0xd3, 0xd6, 0x98, 0x8e, 0xea, 0x92, 0x57, 0xd4, 0x66,
	0xfe, 0x4b, 0x3a, 0x0d, 0x06, 0x92, 0x85, 0x87, 0x49, 0x80, 0x49, 0x80, 0x89, 0x83, 0x89, 0x87,
	0x6c, 0x36, 0xbf, 0xda, 0xd3, 0x8d, 0x0a, 0xae, 0xba, 0x96, 0xff, 0xf1, 0xcc, 0xcf, 0x2d, 0x57,
	0xb2, 0xe6, 0x46, 0xd0, 0xdf, 0x61, 0xfc, 0x71, 0xd6, 0x40, 0x03, 0x7e, 0xa4, 0x6e, 0x0a, 0xdb,
	0xe5, 0x07, 0xc2, 0x31, 0xcb, 0x1f, 0x9f, 0x92, 0x7b, 0x7c, 0xa4, 0xb7, 0x4a, 0xf4, 0x17, 0x68,
	0x81, 0x6e, 0xa6, 0xd9, 0x25, 0xd9, 0xbb, 0x67, 0xbc, 0xc1, 0xbf, 0x46, 0x1c, 0xcf, 0x02, 0x9d,
	0xe4, 0x78, 0xd6, 0xb7, 0xbc, 0x86, 0xb7, 0xd7, 0x90, 0x8e, 0xfe, 0x97, 0x9e, 0x86, 0x10, 0x73,
	0x99, 0xe5, 0x0a, 0xcf, 0xdc, 0xb2, 0x10, 0x9a, 0xc1, 0xd6, 0x88, 0xe4, 0x01, 0xc7, 0x2e, 0x31,
	0xba, 0xae, 0xc9, 0xa1, 0x66, 0x82, 0xd2, 0x67, 0x9e, 0xe3, 0x09, 0x3a, 0x8d, 0xd8, 0xa4, 0x96,
	0x3d, 0x2f, 0x95, 0xa8, 0xf3, 0x97, 0xcf, 0x21, 0x45, 0x5f, 0x43, 0x8a, 0xbe, 0x87, 0x14, 0x61,
	0x22, 0x21, 0xd8, 0x3a, 0x0d, 0xef, 0xbb, 0x83, 0xe2, 0xfc, 0x84, 0x95, 0xbc, 0x2a, 0x5c, 0x51,
	0x05, 0x5a, 0x45, 0x36, 0x2b, 0x8f, 0x7d, 0x6b, 0x77, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xc2,
	0x20, 0xd6, 0x2f, 0xdb, 0x01, 0x00, 0x00,
}

func (m *RBAC) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RBAC) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Rules != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintRbac(dAtA, i, uint64(m.Rules.Size()))
		n1, err := m.Rules.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if m.ShadowRules != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintRbac(dAtA, i, uint64(m.ShadowRules.Size()))
		n2, err := m.ShadowRules.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *RBACPerRoute) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RBACPerRoute) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Rbac != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintRbac(dAtA, i, uint64(m.Rbac.Size()))
		n3, err := m.Rbac.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n3
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeVarintRbac(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *RBAC) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Rules != nil {
		l = m.Rules.Size()
		n += 1 + l + sovRbac(uint64(l))
	}
	if m.ShadowRules != nil {
		l = m.ShadowRules.Size()
		n += 1 + l + sovRbac(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *RBACPerRoute) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Rbac != nil {
		l = m.Rbac.Size()
		n += 1 + l + sovRbac(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovRbac(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozRbac(x uint64) (n int) {
	return sovRbac(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *RBAC) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRbac
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RBAC: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RBAC: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Rules", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRbac
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthRbac
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRbac
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Rules == nil {
				m.Rules = &v2alpha.RBAC{}
			}
			if err := m.Rules.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ShadowRules", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRbac
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthRbac
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRbac
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ShadowRules == nil {
				m.ShadowRules = &v2alpha.RBAC{}
			}
			if err := m.ShadowRules.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRbac(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRbac
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthRbac
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *RBACPerRoute) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRbac
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RBACPerRoute: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RBACPerRoute: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Rbac", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRbac
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthRbac
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRbac
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Rbac == nil {
				m.Rbac = &RBAC{}
			}
			if err := m.Rbac.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRbac(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRbac
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthRbac
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipRbac(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowRbac
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowRbac
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowRbac
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthRbac
			}
			iNdEx += length
			if iNdEx < 0 {
				return 0, ErrInvalidLengthRbac
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowRbac
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipRbac(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
				if iNdEx < 0 {
					return 0, ErrInvalidLengthRbac
				}
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthRbac = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowRbac   = fmt.Errorf("proto: integer overflow")
)
