package objects

import (
	"encoding/binary"
	"fmt"
	"math"
)

type Object bits
type PtrTag uint8
type Addr32 uint32

type Ptr32 struct {
	Tag  PtrTag
	Addr Addr32
}

type Str32 struct {
	Len  uint8
	Addr Addr32
}

func (this Object) Is(mask bits) bool {
	if mask == MASK_F64 {
		return !math.IsNaN(math.Float64frombits(uint64(this)))
	}

	return bits(this)&mask == mask
}

const (
	STR_NULL  = "null"
	STR_STR32 = "Str32"
	STR_PTR32 = "Ptr32"
	STR_BYTES = "Bytes"
	STR_BOOL  = "Bool"
	STR_F64   = "F64"
)

func (this Object) String() string {
	return fmt.Sprintf("%X", bits(this))
}

func (this Object) Type() string {
	field := bits(this)
	if field == 0 {
		return STR_NULL
	}

	if field&MASK_PTR32 == MASK_PTR32 {
		return STR_PTR32
	}

	if field&MASK_STR32 == MASK_STR32 {
		return STR_STR32
	}

	if field&MASK_BYTES == MASK_BYTES {
		return STR_BYTES
	}

	if field&MASK_BOOL == MASK_BOOL {
		return STR_BOOL
	}

	if !math.IsNaN(math.Float64frombits(uint64(field))) {
		return STR_F64
	}

	panic("unrecognized type")
}

func (this Object) Truthy() bool {
	if this == PackedFalse || this == Null {
		return false
	}
	return true
}

func PackF64(f64 float64) Object {
	return Object(math.Float64bits(f64))
}

func UnpackF64(obj Object) float64 {
	f64 := math.Float64frombits(uint64(obj))
	if math.IsNaN(f64) {
		panic("not a float64")
	}

	return f64
}

func PackPtr32(ptr Ptr32) Object {
	var field bits
	(&field).PutU8(uint8(ptr.Tag))
	(&field).PutU32(uint32(ptr.Addr))
	return Object(field | MASK_PTR32)
}

func UnpackPtr32(obj Object) Ptr32 {
	var ptr Ptr32
	field := bits(obj)
	if field&MASK_PTR32 != MASK_PTR32 {
		panic("not a pointer")
	}

	ptr.Tag = PtrTag(field.GetU8())
	ptr.Addr = Addr32(field.GetU32())
	return ptr
}

func PackStr32(ptr Str32) Object {
	var field bits
	(&field).PutU8(ptr.Len)
	(&field).PutU32(uint32(ptr.Addr))
	return Object(field | MASK_STR32)
}

func UnpackStr32(obj Object) Str32 {
	var ptr Str32
	field := bits(obj)
	if field&MASK_STR32 != MASK_STR32 {
		panic("not a string")
	}

	ptr.Len = field.GetU8()
	ptr.Addr = Addr32(field.GetU32())
	return ptr
}

func PackBytes(arr [5]uint8) Object {
	return Object(bytes(arr).asbits(MASK_BYTES))
}

func UnpackBytes(obj Object) bytes {
	field := bits(obj)
	if field&MASK_BYTES != MASK_BYTES {
		panic("not an array")
	}

	return field.asbytes()
}

func PackBool(b bool) Object {
	if b {
		return PackedTrue
	}
	return PackedFalse
}

func UnpackBool(obj Object) bool {
	field := bits(obj)
	if field&MASK_BOOL != MASK_BOOL {
		panic("not a boolean")
	}
	return field.GetU8() == 1
}

const (
	ptrmask  bits = 1<<63 | 1<<49 | 1<<48
	strmask  bits = 1<<63 | 1<<49 | 0<<48
	array    bits = 1<<63 | 1<<49 | 0<<48
	boolmask bits = 0<<63 | 1<<49 | 1<<48

	QNAN     bits = 0x7FF8000000000001
	MASKBITS bits = (1 << 63) | (1 << 49) | (1 << 48)

	MASK_PTR32 bits = QNAN | ptrmask
	MASK_STR32 bits = QNAN | strmask
	MASK_BYTES bits = QNAN | array
	MASK_BOOL  bits = QNAN | boolmask
	MASK_F64   bits = ^QNAN
	MASK_NULL  bits = 0

	PackedTrue  = Object(0x7FFB000000000201)
	PackedFalse = Object(0x7FFB000000000001)
	Null        = Object(0)
)

var _ encoder = (*bits)(nil)
var _ encoder = (*bytes)(nil)

type bits uint64
type bytes [5]uint8

func (this *bits) PutU8(u8 uint8) {
	*this = *this | bits(u8)<<1
}

func (this bits) GetU8() uint8 {
	return uint8(this >> 1)
}

func (this *bits) PutU32(u32 uint32) {
	*this = *this | bits(u32)<<9
}

func (this bits) GetU32() uint32 {
	return uint32(this >> 9)
}

func (this bits) asbytes() bytes {
	var bytes bytes
	bytes.PutU8(this.GetU8())
	bytes.PutU32(this.GetU32())
	return bytes
}

func (this *bytes) PutU8(u8 uint8) {
	(*this)[0] = u8
}

func (this bytes) GetU8() uint8 {
	return this[0]
}

func (this *bytes) PutU32(u32 uint32) {
	binary.LittleEndian.PutUint32((*this)[1:], u32)
}

func (this bytes) GetU32() uint32 {
	return binary.LittleEndian.Uint32(this[1:])
}

func (this bytes) asbits(mask bits) bits {
	(&mask).PutU8(this.GetU8())
	(&mask).PutU32(this.GetU32())
	return mask
}

type encoder interface {
	PutU8(uint8)
	PutU32(uint32)
	GetU8() uint8
	GetU32() uint32
}
