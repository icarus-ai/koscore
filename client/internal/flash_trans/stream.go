package flash_trans

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/kernel-ai/koscore/utils/crypto"
)

type Stream struct {
	R    io.ReadSeeker
	Size int64
	Sha1 []byte
	Name string
}

func NewStreamFile(path string) *Stream {
	fs, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	s := NewStreamIO(fs)
	stat, _ := fs.Stat()
	s.Name = stat.Name()
	return s
}

func NewStreamBytes(d []byte) *Stream {
	return NewStreamIO(bytes.NewReader(d))
}

func NewStreamIO(v io.ReadSeeker) *Stream {
	_, sha1, size := crypto.ComputeMd5AndSha1AndLength(v)
	return &Stream{R: v, Size: int64(size), Sha1: sha1, Name: fmt.Sprintf("%02x", sha1[:5])}
}

func (m *Stream) Close()                               {} //m.R.Close() }
func (m *Stream) Seek(idx int64, _ int) (int64, error) { return m.R.Seek(idx, io.SeekStart) }
func (m *Stream) ToStart()                             { _, _ = m.R.Seek(0, io.SeekStart) }
func (m *Stream) Read(idx int64, data []byte) (int, error) {
	_, _ = m.R.Seek(idx, io.SeekStart)
	return m.R.Read(data)
}
