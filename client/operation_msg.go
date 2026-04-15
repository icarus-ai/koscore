package client

import (
	"errors"

	pkt_msg "github.com/kernel-ai/koscore/client/packets/message"

	"github.com/kernel-ai/koscore/message"
)

// 获取群聊历史消息
func (m *QQClient) GetGroupMessages(gin, start_seq, end_seq uint64) ([]*message.GroupMessage, error) {
	pkt, err := m.sendOidbPacketAndWait(pkt_msg.BuildGetGroupMessagePacket(gin, start_seq, end_seq))
	if err != nil {
		return nil, err
	}
	rsp, err := pkt_msg.ParseGetGroupMessagePacket(pkt.Data)
	if err != nil {
		return nil, err
	}
	var msgs []*message.GroupMessage
	for _, msg := range rsp {
		msgs = append(msgs, message.ParseGroupMessage(m.Uin(), msg))
	}
	return msgs, nil
}

// 获取私聊历史消息
func (m *QQClient) GetPrivateMessages(peer_uin uint64, timestamp, count uint32) ([]*message.PrivateMessage, error) {
	pkt, err := m.sendOidbPacketAndWait(pkt_msg.BuildGetRoamMessagePacket(m.GetUid(peer_uin), timestamp, count))
	if err != nil {
		return nil, err
	}
	rsp, err := pkt_msg.ParseGetRoamMessagePacket(pkt.Data)
	if err != nil {
		return nil, err
	}
	var msgs []*message.PrivateMessage
	for _, msg := range rsp {
		msgs = append(msgs, message.ParsePrivateMessage(m.Uin(), msg))
	}
	return msgs, nil
}

func (m *QQClient) GetC2CMessages(peer_uin, start_seq, end_seq uint64) ([]*message.TempMessage, error) {
	pkt, err := m.sendOidbPacketAndWait(pkt_msg.BuildGetC2CMessagePacket(m.GetUid(peer_uin), start_seq, end_seq))
	if err != nil {
		return nil, err
	}
	rsp, err := pkt_msg.ParseGetC2CMessagePacket(pkt.Data)
	if err != nil {
		return nil, err
	}
	var msgs []*message.TempMessage
	for _, msg := range rsp {
		msgs = append(msgs, message.ParseTempMessage(m.Uin(), msg))
	}
	return msgs, nil
}

// 撤回群聊消息
func (m *QQClient) RecallGroupMessage(gin, seq uint64) error {
	pkt, err := m.sendOidbPacketAndWait(pkt_msg.BuildGroupRecallMessagePacket(gin, seq))
	if err != nil {
		return err
	}
	if len(pkt.Data) == 0 {
		return errors.New("empty response data")
	}
	return nil
}

// 撤回私聊消息
func (m *QQClient) RecallFriendMessage(uin, seq, random, client_seq uint64, timestamp uint32) error {
	_, err := m.sendOidbPacketAndWait(pkt_msg.BuildC2CRecallMessagePacket(m.GetUid(uin), seq, random, client_seq, timestamp))
	return err // sbtx不报错
}
