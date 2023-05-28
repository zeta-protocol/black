// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: black/committee/v1beta1/proposal.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// CommitteeChangeProposal is a gov proposal for creating a new committee or modifying an existing one.
type CommitteeChangeProposal struct {
	Title        string     `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description  string     `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	NewCommittee *types.Any `protobuf:"bytes,3,opt,name=new_committee,json=newCommittee,proto3" json:"new_committee,omitempty"`
}

func (m *CommitteeChangeProposal) Reset()         { *m = CommitteeChangeProposal{} }
func (m *CommitteeChangeProposal) String() string { return proto.CompactTextString(m) }
func (*CommitteeChangeProposal) ProtoMessage()    {}
func (*CommitteeChangeProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_4886de4a6c720e57, []int{0}
}
func (m *CommitteeChangeProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CommitteeChangeProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CommitteeChangeProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CommitteeChangeProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommitteeChangeProposal.Merge(m, src)
}
func (m *CommitteeChangeProposal) XXX_Size() int {
	return m.Size()
}
func (m *CommitteeChangeProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_CommitteeChangeProposal.DiscardUnknown(m)
}

var xxx_messageInfo_CommitteeChangeProposal proto.InternalMessageInfo

// CommitteeDeleteProposal is a gov proposal for removing a committee.
type CommitteeDeleteProposal struct {
	Title       string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	CommitteeID uint64 `protobuf:"varint,3,opt,name=committee_id,json=committeeId,proto3" json:"committee_id,omitempty"`
}

func (m *CommitteeDeleteProposal) Reset()         { *m = CommitteeDeleteProposal{} }
func (m *CommitteeDeleteProposal) String() string { return proto.CompactTextString(m) }
func (*CommitteeDeleteProposal) ProtoMessage()    {}
func (*CommitteeDeleteProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_4886de4a6c720e57, []int{1}
}
func (m *CommitteeDeleteProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CommitteeDeleteProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CommitteeDeleteProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CommitteeDeleteProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommitteeDeleteProposal.Merge(m, src)
}
func (m *CommitteeDeleteProposal) XXX_Size() int {
	return m.Size()
}
func (m *CommitteeDeleteProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_CommitteeDeleteProposal.DiscardUnknown(m)
}

var xxx_messageInfo_CommitteeDeleteProposal proto.InternalMessageInfo

func init() {
	proto.RegisterType((*CommitteeChangeProposal)(nil), "black.committee.v1beta1.CommitteeChangeProposal")
	proto.RegisterType((*CommitteeDeleteProposal)(nil), "black.committee.v1beta1.CommitteeDeleteProposal")
}

func init() {
	proto.RegisterFile("black/committee/v1beta1/proposal.proto", fileDescriptor_4886de4a6c720e57)
}

var fileDescriptor_4886de4a6c720e57 = []byte{
	// 348 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x92, 0xbf, 0x4e, 0x02, 0x41,
	0x10, 0xc6, 0xef, 0xfc, 0x97, 0x70, 0x07, 0x31, 0xb9, 0x10, 0x05, 0x8a, 0x95, 0x90, 0x98, 0x90,
	0x18, 0x76, 0x03, 0x76, 0x76, 0x02, 0x85, 0x74, 0x86, 0xd2, 0x86, 0xec, 0xc1, 0xb8, 0x5c, 0x3c,
	0x76, 0x2e, 0xdc, 0x02, 0xf2, 0x16, 0xbe, 0x84, 0x6f, 0x40, 0xe7, 0x0b, 0x10, 0x2a, 0x4a, 0x2b,
	0xa3, 0xc7, 0x8b, 0x98, 0xfb, 0xc3, 0x86, 0xce, 0xc2, 0x6e, 0xbe, 0x6f, 0xbe, 0xcb, 0xfc, 0x6e,
	0x66, 0xad, 0xeb, 0x17, 0x3e, 0xe7, 0x6c, 0x88, 0x93, 0x89, 0xa7, 0x14, 0x00, 0x9b, 0x37, 0x5d,
	0x50, 0xbc, 0xc9, 0x82, 0x29, 0x06, 0x18, 0x72, 0x9f, 0x06, 0x53, 0x54, 0xe8, 0x5c, 0xc4, 0x31,
	0xaa, 0x63, 0x34, 0x8b, 0x55, 0xca, 0x43, 0x0c, 0x27, 0x18, 0x0e, 0x92, 0x14, 0x4b, 0x45, 0xfa,
	0x49, 0xa5, 0x28, 0x50, 0x60, 0xea, 0xc7, 0x55, 0xe6, 0x96, 0x05, 0xa2, 0xf0, 0x81, 0x25, 0xca,
	0x9d, 0x3d, 0x33, 0x2e, 0x97, 0x69, 0xab, 0xf6, 0x61, 0x5a, 0x97, 0x9d, 0xfd, 0x84, 0xce, 0x98,
	0x4b, 0x01, 0x8f, 0x19, 0x85, 0x53, 0xb4, 0x4e, 0x95, 0xa7, 0x7c, 0x28, 0x99, 0x55, 0xb3, 0x9e,
	0xeb, 0xa7, 0xc2, 0xa9, 0x5a, 0xf6, 0x08, 0xc2, 0xe1, 0xd4, 0x0b, 0x94, 0x87, 0xb2, 0x74, 0x94,
	0xf4, 0x0e, 0x2d, 0xe7, 0xc1, 0x2a, 0x48, 0x58, 0x0c, 0x34, 0x78, 0xe9, 0xb8, 0x6a, 0xd6, 0xed,
	0x56, 0x91, 0xa6, 0x18, 0x74, 0x8f, 0x41, 0xef, 0xe5, 0xb2, 0x5d, 0xd8, 0xac, 0x1a, 0x39, 0x4d,
	0xd0, 0xcf, 0x4b, 0x58, 0x68, 0x75, 0x47, 0x36, 0xab, 0x46, 0x25, 0xfb, 0x41, 0x81, 0xf3, 0xfd,
	0x06, 0x68, 0x07, 0xa5, 0x02, 0xa9, 0x6a, 0xef, 0x87, 0xf4, 0x5d, 0xf0, 0x41, 0xfd, 0x9f, 0xbe,
	0x65, 0xe5, 0x35, 0xf9, 0xc0, 0x1b, 0x25, 0xf0, 0x27, 0xed, 0xf3, 0xe8, 0xeb, 0xca, 0xd6, 0xa3,
	0x7a, 0xdd, 0xbe, 0xad, 0x43, 0xbd, 0xd1, 0x5f, 0x9c, 0xed, 0xde, 0xfa, 0x87, 0x18, 0xeb, 0x88,
	0x98, 0xdb, 0x88, 0x98, 0xdf, 0x11, 0x31, 0xdf, 0x76, 0xc4, 0xd8, 0xee, 0x88, 0xf1, 0xb9, 0x23,
	0xc6, 0xd3, 0x8d, 0xf0, 0xd4, 0x78, 0xe6, 0xc6, 0x97, 0x66, 0xf1, 0xc9, 0x1b, 0x3e, 0x77, 0xc3,
	0xa4, 0x62, 0xaf, 0x07, 0xaf, 0x44, 0x2d, 0x03, 0x08, 0xdd, 0xb3, 0x64, 0x7b, 0xb7, 0xbf, 0x01,
	0x00, 0x00, 0xff, 0xff, 0x58, 0x7f, 0x45, 0x9a, 0x44, 0x02, 0x00, 0x00,
}

func (m *CommitteeChangeProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CommitteeChangeProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CommitteeChangeProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.NewCommittee != nil {
		{
			size, err := m.NewCommittee.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintProposal(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *CommitteeDeleteProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CommitteeDeleteProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CommitteeDeleteProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.CommitteeID != 0 {
		i = encodeVarintProposal(dAtA, i, uint64(m.CommitteeID))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintProposal(dAtA []byte, offset int, v uint64) int {
	offset -= sovProposal(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *CommitteeChangeProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	if m.NewCommittee != nil {
		l = m.NewCommittee.Size()
		n += 1 + l + sovProposal(uint64(l))
	}
	return n
}

func (m *CommitteeDeleteProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	if m.CommitteeID != 0 {
		n += 1 + sovProposal(uint64(m.CommitteeID))
	}
	return n
}

func sovProposal(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozProposal(x uint64) (n int) {
	return sovProposal(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *CommitteeChangeProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
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
			return fmt.Errorf("proto: CommitteeChangeProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CommitteeChangeProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NewCommittee", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.NewCommittee == nil {
				m.NewCommittee = &types.Any{}
			}
			if err := m.NewCommittee.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProposal
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *CommitteeDeleteProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
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
			return fmt.Errorf("proto: CommitteeDeleteProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CommitteeDeleteProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommitteeID", wireType)
			}
			m.CommitteeID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CommitteeID |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProposal
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipProposal(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProposal
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
					return 0, ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowProposal
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
				return 0, ErrInvalidLengthProposal
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupProposal
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthProposal
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthProposal        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProposal          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupProposal = fmt.Errorf("proto: unexpected end of group")
)
