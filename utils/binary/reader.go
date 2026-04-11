package binary

import (
	"encoding/binary"
	"io"
	"strconv"
	"unsafe"

	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
	"github.com/kernel-ai/koscore/utils/types"
)

type Reader struct {
	rdio io.Reader
	data []byte
	pos  int
}

func ParseReader(rdio io.Reader) *Reader { return &Reader{rdio: rdio} }
func NewReader(data []byte) *Reader      { return &Reader{data: data} }

func (r *Reader) Len() int {
	if r.rdio != nil {
		return -1
	}
	return len(r.data) - r.pos
}

func (r *Reader) ReadByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

// String means read all available data and return them as a string
//
// if r.reader got error, it will returns as err.Error()
func (r *Reader) String() string {
	if r.rdio != nil {
		data, err := io.ReadAll(r.rdio)
		if err != nil {
			return err.Error()
		}
		return utils.B2S(data)
	}
	s := string(r.data[r.pos:])
	r.pos = 0
	r.data = r.data[:0]
	return s
}

// ReadAll means read all available data and return them
//
// if r.reader got error, it will return nil
func (r *Reader) ReadAll() []byte {
	if r.rdio != nil {
		data, err := io.ReadAll(r.rdio)
		if err != nil {
			return nil
		}
		return data
	}
	s := r.data[r.pos:]
	r.pos = 0
	r.data = r.data[:0]
	buf := make([]byte, len(s))
	copy(buf, s)
	return s
}

func (r *Reader) ReadU8() (v uint8) {
	if r.rdio != nil {
		_, _ = r.rdio.Read(unsafe.Slice(&v, 1))
		return
	}
	v = r.data[r.pos]
	r.pos++
	return
}

func readint[T ~uint16 | ~uint32 | ~uint64](r *Reader) (v T) {
	sz := unsafe.Sizeof(v)
	buf := make([]byte, 8)
	if r.rdio != nil {
		_, _ = r.rdio.Read(buf[8-sz:])
	} else {
		copy(buf[8-sz:], r.data[r.pos:r.pos+int(sz)])
		r.pos += int(sz)
	}
	v = (T)(binary.BigEndian.Uint64(buf))
	return
}

func (r *Reader) ReadU16() (v uint16) { return readint[uint16](r) }
func (r *Reader) ReadU32() (v uint32) { return readint[uint32](r) }
func (r *Reader) ReadU64() (v uint64) { return readint[uint64](r) }

func (r *Reader) SkipBytes(length int) {
	if r.rdio != nil {
		_, _ = r.rdio.Read(make([]byte, length))
		return
	}
	r.pos += length
}

// ReadBytesNoCopy 不拷贝读取的数据, 用于读取后立即使用, 慎用
//
// 如需使用, 请确保 Reader 未被回收
func (r *Reader) ReadBytesNoCopy(length int) (v []byte) {
	if r.rdio != nil {
		return r.ReadBytes(length)
	}
	v = r.data[r.pos : r.pos+length]
	r.pos += length
	return
}

func (r *Reader) ReadBytes(length int) (v []byte) {
	// 返回一个全新的数组罢
	v = make([]byte, length)
	if r.rdio != nil {
		_, _ = r.rdio.Read(v)
	} else {
		copy(v, r.data[r.pos:r.pos+length])
		r.pos += length
	}
	return
}

func (r *Reader) ReadString(length int) string {
	return utils.B2S(r.ReadBytes(length))
}

func (r *Reader) SkipBytesWithLength(prefix string, withPerfix bool) {
	var length int
	switch prefix {
	case "u8":
		length = int(r.ReadU8())
	case "u16":
		length = int(r.ReadU16())
	case "u32":
		length = int(r.ReadU32())
	case "u64":
		length = int(r.ReadU64())
	default:
		panic("invaild prefix")
	}
	if withPerfix {
		plus, err := strconv.Atoi(prefix[1:])
		if err != nil {
			panic(err)
		}
		length -= plus / 8
	}
	r.SkipBytes(length)
}

func (r *Reader) ReadBytesWithLength(prefix string, withPerfix bool) []byte {
	var length int
	switch prefix {
	case "u8":
		length = int(r.ReadU8())
	case "u16":
		length = int(r.ReadU16())
	case "u32":
		length = int(r.ReadU32())
	case "u64":
		length = int(r.ReadU64())
	default:
		panic("invaild prefix")
	}
	if withPerfix {
		plus, err := strconv.Atoi(prefix[1:])
		if err != nil {
			panic(err)
		}
		length -= plus / 8
	}
	return r.ReadBytes(length)
}

func (r *Reader) ReadStringWithLength(prefix string, withPerfix bool) string {
	return utils.B2S(r.ReadBytesWithLength(prefix, withPerfix))
}

func (r *Reader) ReadTlv() (result types.Tlvs) {
	result = make(types.Tlvs)
	count := r.ReadU16()
	for i := 0; i < int(count); i++ {
		tag := r.ReadU16()
		result[tag] = r.ReadBytes(int(r.ReadU16()))
	}
	return
}

func (r *Reader) ReadI8() (v int8)   { return int8(r.ReadU8()) }
func (r *Reader) ReadI16() (v int16) { return int16(r.ReadU16()) }
func (r *Reader) ReadI32() (v int32) { return int32(r.ReadU32()) }
func (r *Reader) ReadI64() (v int64) { return int64(r.ReadU64()) }

func (r *Reader) ReadVarint() (int64, error)   { return binary.ReadVarint(r) }
func (r *Reader) ReadUvarint() (uint64, error) { return binary.ReadUvarint(r) }

// *****

func (r *Reader) ReadLength(flag prefix.Prefix) int {
	lengthCounted := (flag & prefix.WithPrefix) != 0 // != 0 is faster than > 0
	prefixLength := int(flag & 0b0111)
	size := uint64(prefixLength)
	switch prefixLength {
	case 1:
		size = uint64(r.ReadU8())
	case 2:
		size = uint64(r.ReadU16())
	case 4:
		size = uint64(r.ReadU32())
	case 8:
		size = r.ReadU64()
	}
	if lengthCounted {
		size -= uint64(prefixLength)
	}
	return int(size)
}

func (r *Reader) ReadLengthBytes(flag prefix.Prefix) []byte  { return r.ReadBytes(r.ReadLength(flag)) }
func (r *Reader) ReadLengthString(flag prefix.Prefix) string { return string(r.ReadLengthBytes(flag)) }

//func (b *Reader) WriteString      (str  string)                     *Builder { return b.WriteBytes([]byte(str)) }
