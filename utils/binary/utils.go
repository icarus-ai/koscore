package binary

// from https://github.com/Mrs4s/MiraiGo/blob/master/binary/utils.go

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"net"
	"sync"
)

var Empty []byte

type zlibWriter struct {
	w   *zlib.Writer
	buf *bytes.Buffer
}

var zlibPool = sync.Pool{
	New: func() any {
		buf := new(bytes.Buffer)
		return &zlibWriter{w: zlib.NewWriter(buf), buf: buf}
	}}

func acquireZlibWriter() *zlibWriter {
	ret := zlibPool.Get().(*zlibWriter)
	ret.buf.Reset()
	ret.w.Reset(ret.buf)
	return ret
}

func (w *zlibWriter) Release() {
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16
	if w.buf.Cap() < maxSize {
		w.buf.Reset()
		zlibPool.Put(w)
	}
}

// ***** *****

type GzipWriter struct {
	w   *gzip.Writer
	buf *bytes.Buffer
}

var gzipPool = sync.Pool{
	New: func() any {
		buf := new(bytes.Buffer)
		return &GzipWriter{w: gzip.NewWriter(buf), buf: buf}
	}}

func acquireGzipWriter() *GzipWriter {
	ret := gzipPool.Get().(*GzipWriter)
	ret.buf.Reset()
	ret.w.Reset(ret.buf)
	return ret
}

func (w *GzipWriter) Release() {
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16
	if w.buf.Cap() < maxSize {
		w.buf.Reset()
		gzipPool.Put(w)
	}
}

//func (w *GzipWriter) Write(p []byte) (int, error) { return w.w.Write(p) }
//func (w *GzipWriter) Close() error { return w.w.Close() }
//func (w *GzipWriter) Bytes() []byte { return w.buf.Bytes() }

// ***** *****

func ZlibCompress(data []byte) []byte {
	ctx := acquireZlibWriter()
	_, _ = ctx.w.Write(data)
	_ = ctx.w.Close()
	ret := make([]byte, len(ctx.buf.Bytes()))
	copy(ret, ctx.buf.Bytes())
	ctx.Release()
	return ret
}

func GZipCompress(data []byte) []byte {
	ctx := acquireGzipWriter()
	_, _ = ctx.w.Write(data)
	_ = ctx.w.Close()
	ret := make([]byte, len(ctx.buf.Bytes()))
	copy(ret, ctx.buf.Bytes())
	ctx.Release()
	return ret
}

func ZlibUncompress(src []byte) []byte {
	b := bytes.NewReader(src)
	r, _ := zlib.NewReader(b)
	defer r.Close()
	var out bytes.Buffer
	_, _ = out.ReadFrom(r)
	return out.Bytes()
}

func GZipUncompress(src []byte) []byte {
	b := bytes.NewReader(src)
	r, _ := gzip.NewReader(b)
	defer r.Close()
	var out bytes.Buffer
	_, _ = out.ReadFrom(r)
	return out.Bytes()
}

func UInt32ToIPV4Address(i uint32) string {
	ip := net.IP{0, 0, 0, 0}
	binary.LittleEndian.PutUint32(ip, i)
	return ip.String()
}
