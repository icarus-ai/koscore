package network

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/internal/network/conn.go

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type TCPClient struct {
	lock                 sync.RWMutex
	conn                 net.Conn
	Connected            bool
	plannedDisconnect    func(*TCPClient)
	unexpectedDisconnect func(*TCPClient, error)
}

var ErrConnectionClosed = errors.New("connection closed")

// 预料中的断开连接 | 如调用 Close() Connect()
func (t *TCPClient) PlannedDisconnect(f func(*TCPClient)) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.plannedDisconnect = f
}

// 未预料的断开连接
func (t *TCPClient) UnexpectedDisconnect(f func(*TCPClient, error)) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.unexpectedDisconnect = f
}

func (t *TCPClient) Connect(addr string) error {
	t.Close()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "dial tcp error")
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.conn = conn
	t.Connected = true
	return nil
}

func (t *TCPClient) Write(buf []byte) error {
	if conn := t.getConn(); conn != nil {
		_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_, err := conn.Write(buf)
		if err == nil {
			return nil
		}
		t.unexpectedClose(err)
	}
	return ErrConnectionClosed
}

func (t *TCPClient) ReadBytes(l int) ([]byte, error) {
	if conn := t.getConn(); conn != nil {
		buf := make([]byte, l)
		_, err := io.ReadFull(conn, buf)
		if err == nil {
			return buf, nil
		}
		// time.Sleep(time.Millisecond * 100) // 服务器会发送offline包后立即断开连接, 此时还没解析, 可能还是得加锁
		t.unexpectedClose(err)
	}
	return nil, ErrConnectionClosed
}

func (t *TCPClient) ReadInt32() (int32, error) {
	b, err := t.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}

func (t *TCPClient) Close() {
	t.close()
	t.invokePlannedDisconnect()
}

func (t *TCPClient) unexpectedClose(err error) {
	t.close()
	t.invokeUnexpectedDisconnect(err)
}

func (t *TCPClient) close() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.conn != nil {
		_ = t.conn.Close()
		t.conn = nil
	}
}

func (t *TCPClient) invokePlannedDisconnect() {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.plannedDisconnect != nil && t.Connected {
		go t.plannedDisconnect(t)
		t.Connected = false
	}
}

func (t *TCPClient) invokeUnexpectedDisconnect(err error) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.unexpectedDisconnect != nil && t.Connected {
		go t.unexpectedDisconnect(t, err)
		t.Connected = false
	}
}

func (t *TCPClient) getConn() net.Conn {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.conn
}
