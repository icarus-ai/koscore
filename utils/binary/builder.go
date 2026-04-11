package binary

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"net"
	"runtime"
	"strconv"
	"sync"

	"github.com/fumiama/gofastTEA"

	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
)

// from https://github.com/Mrs4s/MiraiGo/blob/master/binary/pool.go

var bufferPool = sync.Pool{New: func() any { return new(Builder) }}

// 从池中取出一个 Builder
func SelectBuilder(key []byte) *Builder {
	// 因为 bufferPool 定义有 New 函数
	// 所以 bufferPool.Get() 永不为 nil
	// 不用判空
	return bufferPool.Get().(*Builder).init(key)
}

// 将 Builder 放回池中
func (b *Builder) PutBuilder() {
	// See https://golang.org/issue/23199
	const maxSize = 32 * 1024
	if b.hasput {
		return
	}
	b.hasput = true
	if b.buffer.Cap() < maxSize { // 对于大Buffer直接丢弃
		b.buffer.Reset()
		bufferPool.Put(b)
	}
}

func (b *Builder) init(key []byte) *Builder {
	b.key = tea.NewTeaCipher(key)
	b.usetea = len(key) == 16
	b.hasput = false
	return b
}

type Builder struct {
	buffer bytes.Buffer
	key    tea.TEA
	usetea bool
	hasput bool
	hasset bool
	io.Writer
	io.ReaderFrom
}

// from https://github.com/Mrs4s/MiraiGo/blob/master/binary/writer.go

func NewWriterF(f func(writer *Builder)) []byte {
	w := SelectBuilder(nil)
	f(w)
	return w.ToBytes()
}

func ToBytes(i any) []byte {
	return NewWriterF(func(w *Builder) {
		switch t := i.(type) {
		case int16:
			w.WriteU16(uint16(t))
		case int32:
			w.WriteU32(uint32(t))
		}
	})
}

// with finalizer of itself.
//
// Be sure to use all data before builder is GCed.
func NewBuilder(key ...byte) *Builder {
	b := SelectBuilder(key)
	if !b.hasset {
		b.hasset = true
		runtime.SetFinalizer(b, func(b any) { b.(*Builder).PutBuilder() })
	}
	return b
}

func (b *Builder) Len() int { return b.buffer.Len() }

// GC 不安全, 确保在 Builder 被回收前使用
//func (b *Builder) ToReader()  io.Reader    { return &b.buffer }
//func (b *Builder) Buffer  () *bytes.Buffer { return &b.buffer }

// return data with tea encryption if key is set
//
// GC 安全, 返回的数据在 Builder 被销毁之后仍能被正确读取,
// 但是只能调用一次, 调用后 Builder 即失效
func (b *Builder) ToBytes() []byte {
	defer b.PutBuilder()
	if b.usetea {
		return b.key.Encrypt(b.buffer.Bytes())
	}
	buf := make([]byte, b.Len())
	copy(buf, b.buffer.Bytes())
	return buf
}

// Pack TLV with tea encryption if key is set
//
// GC 安全, 返回的数据在 Builder 被销毁之后仍能被正确读取,
// 但是只能调用一次, 调用后 Builder 即失效
func (b *Builder) Pack(typ uint16) []byte {
	defer b.PutBuilder()
	n, buf := 0, make([]byte, b.Len()+2+2+16)
	if b.usetea {
		n = b.key.EncryptTo(b.buffer.Bytes(), buf[2+2:])
	} else {
		n = copy(buf[2+2:], b.buffer.Bytes())
	}
	binary.BigEndian.PutUint16(buf[0:2], typ)         // type
	binary.BigEndian.PutUint16(buf[2:2+2], uint16(n)) // length
	return buf[:n+2+2]
}

func (b *Builder) WriteBool(v bool) *Builder {
	if v {
		b.WriteU8('1')
	} else {
		b.WriteU8('0')
	}
	return b
}

// WritePacketBytes prefix must not be empty
func (b *Builder) WritePacketBytes(v []byte, prefix string, withPrefix bool) *Builder {
	n := len(v)
	if withPrefix {
		plus, e := strconv.Atoi(prefix[1:])
		if e != nil {
			panic(e)
		}
		n += plus / 8
	}
	switch prefix {
	case "u8":
		b.WriteU8(uint8(n))
	case "u16":
		b.WriteU16(uint16(n))
	case "u32":
		b.WriteU32(uint32(n))
	case "u64":
		b.WriteU64(uint64(n))
	default:
		panic("invaild prefix")
	}
	b.WriteBytes(v)
	return b
}

func (b *Builder) WritePacketString(s, prefix string, withPrefix bool) *Builder {
	return b.WritePacketBytes(utils.S2B(s), prefix, withPrefix)
}

// Write for impl. io.Writer
func (b *Builder) Write(p []byte) (n int, err error) { return b.buffer.Write(p) }

func (b *Builder) EncryptAndWrite(key []byte, data []byte) *Builder {
	_, _ = b.Write(tea.NewTeaCipher(key).Encrypt(data))
	return b
}

// ReadFrom for impl. io.ReaderFrom
func (b *Builder) ReadFrom(r io.Reader) (n int64, err error) { return io.Copy(&b.buffer, r) }

func (b *Builder) WriteLenBytes(v []byte) *Builder {
	b.WriteU16(uint16(len(v)))
	b.WriteBytes(v)
	return b
}

func (b *Builder) WriteBytes(v []byte) *Builder {
	_, _ = b.Write(v)
	return b
}

func (b *Builder) WriteLenString(v string) *Builder { return b.WriteLenBytes(utils.S2B(v)) }

func (b *Builder) WriteStruct(data ...any) *Builder {
	for _, data := range data {
		_ = binary.Write(&b.buffer, binary.BigEndian, data)
	}
	return b
}

func (b *Builder) WriteU8(v uint8) *Builder {
	b.buffer.WriteByte(v)
	return b
}

func (b *Builder) WriteU16(v uint16) *Builder {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v)
	b.buffer.Write(buf)
	return b
}

func (b *Builder) WriteU32(v uint32) *Builder {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)
	b.buffer.Write(buf)
	return b
}

func (b *Builder) WriteU64(v uint64) *Builder {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, v)
	b.buffer.Write(buf)
	return b
}

func (b *Builder) WriteI8(v int8) *Builder        { return b.WriteU8(byte(v)) }
func (b *Builder) WriteI16(v int16) *Builder      { return b.WriteU16(uint16(v)) }
func (b *Builder) WriteI32(v int32) *Builder      { return b.WriteU32(uint32(v)) }
func (b *Builder) WriteI64(v int64) *Builder      { return b.WriteU64(uint64(v)) }
func (b *Builder) WriteFloat(v float32) *Builder  { return b.WriteU32(math.Float32bits(v)) }
func (b *Builder) WriteDouble(v float64) *Builder { return b.WriteU64(math.Float64bits(v)) }

func (b *Builder) WriteTLV(tlvs ...[]byte) *Builder {
	b.WriteU16(uint16(len(tlvs)))
	_, _ = io.Copy(b, (*net.Buffers)(&tlvs))
	return b
}

// *****

//func (b *Builder) WriteString(str  string) *Builder { return b.WriteBytes([]byte(str)) }

func (b *Builder) WriteLength(size int, flag prefix.Prefix, addition ...int) *Builder {
	lengthCounted := (flag & prefix.WithPrefix) != 0 // != 0 is faster than > 0
	prefixLength := int(flag & 0b0111)
	if lengthCounted {
		size += prefixLength
		if len(addition) > 0 {
			size += addition[0]
		}
	}
	switch prefixLength {
	case 1:
		b.WriteU8(uint8(size))
	case 2:
		b.WriteU16(uint16(size))
	case 4:
		b.WriteU32(uint32(size))
	case 8:
		b.WriteU64(uint64(size))
	}
	return b
}

func (b *Builder) WriteLengthBytes(data []byte, flag prefix.Prefix) *Builder {
	return b.WriteLength(len(data), flag).WriteBytes(data)
}
func (b *Builder) WriteLengthString(str string, flag prefix.Prefix) *Builder {
	return b.WriteLengthBytes([]byte(str), flag)
}

func (b *Builder) WriteLenBarrier(byt *Builder, flag prefix.Prefix, includePrefix bool, addition ...int) *Builder {
	defer byt.PutBuilder()
	if includePrefix {
		b.WriteLength(byt.buffer.Len(), flag|prefix.WithPrefix, addition...)
	} else {
		b.WriteLength(byt.buffer.Len(), flag|prefix.LengthOnly, addition...)
	}
	b.buffer.Write(byt.buffer.Bytes())
	return b
}
