package client

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/RomiChan/syncx"
	"github.com/pkg/errors"

	"github.com/kernel-ai/koscore/client/event"
	"github.com/kernel-ai/koscore/client/internal"
	"github.com/kernel-ai/koscore/client/internal/network"
	"github.com/kernel-ai/koscore/client/packets/structs"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

type handlerCallFn func(sso *sso_type.SsoPacket, err error)
type handlerMessage struct {
	Packet *sso_type.SsoPacket
	Error  error
}

type PacketContext struct {
	*QQClient
	sock     *internal.SocketContext
	handlers syncx.Map[uint32, handlerCallFn]
}

func (m *PacketContext) dispatchPacket(payload []byte) {
	defer func() {
		if pan := recover(); pan != nil {
			m.DUMP(payload, "panic on decoder error: %v %v", pan, debug.Stack())
			m.LOGE("panic on decoder error: %v %v", pan, debug.Stack())
		}
	}()

	//m.LOGD("PacketContext::DispatchPacket: raw %d %X", len(data), data)
	sso, e := structs.ParseSsoPacket(m.session, payload)
	if e != nil {
		m.LOGW("dispatch: parse sso packet: %v", e)
		if fn, ok := m.handlers.LoadAndDelete(sso.Sequence); ok {
			fn(sso, e)
		}
		return
	}
	if sso.RetCode == 0 {
		m.message_handle_parse_packet(sso)
		return
	}

	// does not need decoder
	// -10001正常缓存过期 -10003登录失效？
	switch sso.RetCode {
	case -10001, -10008:
		if fn, ok := m.handlers.LoadAndDelete(sso.Sequence); ok {
			fn(sso, event.ErrSessionExpired)
		}
	case -10003:
		if fn, ok := m.handlers.LoadAndDelete(sso.Sequence); ok {
			fn(sso, event.ErrAuthenticationFailed)
		}
	default:
		m.LOGE("parse incoming packet error: %s (%d)", sso.Extra, sso.RetCode)
		m.sock.Disconnect()
		go m.Events.Disconnected.dispatch(m.QQClient, &event.Disconnected{Message: fmt.Sprintf("%s (%d)", sso.Extra, sso.RetCode)})
	}
}

func (m *PacketContext) SendPacket(pkt *sso_type.SsoPacket) (e error) {
	if _, pkt.Data, e = m.uniPacket(pkt); e != nil {
		return e
	}
	return m.sock.Send(pkt.Data)
}

func (m *PacketContext) SendPacketAndWait(pkt *sso_type.SsoPacket) (*sso_type.SsoPacket, error) {
	seq, data, e := m.uniPacket(pkt)
	if e != nil {
		return nil, e
	}

	ch := make(chan handlerMessage, 1)
	m.handlers.Store(seq, func(sso *sso_type.SsoPacket, err error) { ch <- handlerMessage{Packet: sso, Error: err} })

	if e = m.sock.Send(data); e != nil {
		m.handlers.Delete(seq)
		return nil, e
	}

	retry := 0
	for {
		select {
		case rsp := <-ch:
			return rsp.Packet, rsp.Error
		case <-time.After(time.Second * 15):
			retry++
			if retry < 2 {
				_ = m.sock.Send(data)
				continue
			}
			m.handlers.Delete(seq)
			return nil, errors.New("Packet timed out " + pkt.Command)
		}
	}
}

func NewPacketContext(wt_ctx *QQClient) *PacketContext {
	ctx := &PacketContext{QQClient: wt_ctx}
	ctx.sock = internal.NewSocketContext(
		ctx.dispatchPacket,
		ctx.quickReconnect,
		ctx.plannedDisconnect,
		ctx.unexpectedDisconnect,
	)
	return ctx
}

// 中断连接 不释放资源
func (m *PacketContext) Disconnect()     { m.sock.Disconnect() }
func (m *PacketContext) Connect() error  { return m.sock.Connect() }
func (m *PacketContext) IsConnect() bool { return m.sock.IsConnect() }

// 计划中断线事件
func (m *PacketContext) plannedDisconnect(_ *network.TCPClient) {
	m.LOGD("planned disconnect.")
	m.stat.DisconnectTimes.Add(1)
	m.Online.Store(false)
}

// 非预期断线事件
func (m *PacketContext) unexpectedDisconnect(_ *network.TCPClient, e error) {
	m.LOGE("unexpected disconnect: %v", e)
	m.stat.DisconnectTimes.Add(1)
	m.Online.Store(false)
	if err := m.sock.Connect(); err != nil {
		m.LOGE("connect server error: %v", err)
		m.Events.Disconnected.dispatch(m.QQClient, &event.Disconnected{Message: "connection dropped by server."})
		return
	}
	if e := m.register(); e != nil {
		m.LOGE("register client failed: %v", e)
		m.sock.Disconnect()
		m.Events.Disconnected.dispatch(m.QQClient, &event.Disconnected{Message: "register error"})
		return
	}
}

// 快速重连
func (m *PacketContext) quickReconnect() {
	m.sock.Disconnect()
	time.Sleep(time.Millisecond * 200)
	if err := m.sock.Connect(); err != nil {
		m.LOGE("connect server error: %v", err)
		m.Events.Disconnected.dispatch(m.QQClient, &event.Disconnected{Message: "quick reconnect failed"})
		return
	}
	if err := m.register(); err != nil {
		m.LOGE("register client failed: %v", err)
		m.sock.Disconnect()
		m.Events.Disconnected.dispatch(m.QQClient, &event.Disconnected{Message: "register error"})
		return
	}
}
