package utils

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/statistics.go

import (
	"bytes"
	"strconv"
	"sync/atomic"
)

type (
	stat_pkt_t struct {
		Recv atomic.Uint64 // Received
		Sent atomic.Uint64
	}
	Statistics struct {
		Packet struct {
			stat_pkt_t
			Lost atomic.Uint64
			Size stat_pkt_t
		}
		Message struct {
			stat_pkt_t
			LastTime atomic.Int64
		}
		DisconnectCount atomic.Int64
		//LostCount     atomic.Uint32
		HeartBeat struct {
			Count atomic.Int32 // 心跳包协程数
			Sso   atomic.Uint64
			Alive atomic.Uint64
		}
	}
)

// encodes the wrapped statistics into JSON.
func (m *Statistics) MarshalJSON() ([]byte, error) {
	var w bytes.Buffer
	w.Grow(256)
	w.WriteString(`{"packet_received":`)
	w.WriteString(strconv.FormatUint(m.Packet.Recv.Load(), 10))
	w.WriteString(`,"packet_sent":`)
	w.WriteString(strconv.FormatUint(m.Packet.Sent.Load(), 10))
	//w.WriteString(`,"packet_lost":`)
	//w.WriteString(strconv.FormatUint(m.Packet.Lost.Load(), 10))
	w.WriteString(`,"message_received":`)
	w.WriteString(strconv.FormatUint(m.Message.Recv.Load(), 10))
	w.WriteString(`,"message_sent":`)
	w.WriteString(strconv.FormatUint(m.Message.Sent.Load(), 10))
	w.WriteString(`,"last_message_time":`)
	w.WriteString(strconv.FormatInt(m.Message.LastTime.Load(), 10))
	w.WriteString(`,"disconnect_times":`)
	w.WriteString(strconv.FormatUint(uint64(m.DisconnectCount.Load()), 10))
	//w.WriteString(`,"lost_times":`)
	//w.WriteString(strconv.FormatUint(uint64(m.LostCount.Load()), 10))
	w.WriteByte('}')
	return w.Bytes(), nil
}
