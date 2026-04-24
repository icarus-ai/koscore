package client

import (
	pkt_msg "github.com/kernel-ai/koscore/client/packets/message"

	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils/exception"
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
		msgs = append(msgs, message.ParseGroupMessage(m.session.Info.Uin, msg))
	}
	return msgs, nil
}

// 获取私聊历史消息
func (m *QQClient) GetPrivateMessages(peer_uin uint64, timestamp, count uint32) ([]*message.PrivateMessage, error) {
	uid, err := m.GetUid(peer_uin)
	if err != nil {
		return nil, err
	}
	pkt, err := m.sendOidbPacketAndWait(pkt_msg.BuildGetRoamMessagePacket(uid, timestamp, count))
	if err != nil {
		return nil, err
	}
	rsp, err := pkt_msg.ParseGetRoamMessagePacket(pkt.Data)
	if err != nil {
		return nil, err
	}
	var msgs []*message.PrivateMessage
	for _, msg := range rsp {
		msgs = append(msgs, message.ParsePrivateMessage(m.session.Info.Uin, msg))
	}
	return msgs, nil
}

func (m *QQClient) GetC2CMessages(peer_uin, start_seq, end_seq uint64) ([]*message.TempMessage, error) {
	uid, err := m.GetUid(peer_uin)
	if err != nil {
		return nil, err
	}
	pkt, err := m.sendOidbPacketAndWait(pkt_msg.BuildGetC2CMessagePacket(uid, start_seq, end_seq))
	if err != nil {
		return nil, err
	}
	rsp, err := pkt_msg.ParseGetC2CMessagePacket(pkt.Data)
	if err != nil {
		return nil, err
	}
	var msgs []*message.TempMessage
	for _, msg := range rsp {
		msgs = append(msgs, message.ParseTempMessage(m.session.Info.Uin, msg))
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
		return exception.ErrEmptyRsp
	}
	return nil
}

// 撤回私聊消息
func (m *QQClient) RecallFriendMessage(uin, seq, random, client_seq uint64, timestamp uint32) error {
	uid, err := m.GetUid(uin)
	if err != nil {
		return err
	}
	_, err = m.sendOidbPacketAndWait(pkt_msg.BuildC2CRecallMessagePacket(uid, seq, random, client_seq, timestamp))
	return err // sbtx不报错
}
