package network

// from https://github.com/Mrs4s/MiraiGo/blob/master/binary/reader.go

import (
	"encoding/binary"
	"io"
	"net"
)

type NetReader struct{ conn net.Conn }

func NewNetReader(conn net.Conn) *NetReader {
	return &NetReader{conn: conn}
}

func (r *NetReader) ReadByte() (byte, error) {
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

func (r *NetReader) ReadBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	_, e := io.ReadFull(r.conn, b)
	return b, e
}

func (r *NetReader) ReadInt32() (int32, error) {
	b := make([]byte, 4)
	_, e := r.conn.Read(b)
	if e != nil {
		return 0, e
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}
