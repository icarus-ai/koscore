package binary

// from https://github.com/Mrs4s/MiraiGo/blob/master/binary/reader.go

import (
	"encoding/binary"
	"io"
	"net"
)

type NetworkReader struct{ conn net.Conn }

func NewNetworkReader(conn net.Conn) *NetworkReader {
	return &NetworkReader{conn: conn}
}

func (r *NetworkReader) ReadByte() (byte, error) {
	b := make([]byte, 1)
	n, e := r.conn.Read(b)
	if e != nil {
		return 0, e
	}
	if n != 1 {
		return r.ReadByte()
	}
	return b[0], nil
}

func (r *NetworkReader) ReadBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	_, e := io.ReadFull(r.conn, b)
	return b, e
}

func (r *NetworkReader) ReadInt32() (int32, error) {
	b := make([]byte, 4)
	_, e := r.conn.Read(b)
	if e != nil {
		return 0, e
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}
