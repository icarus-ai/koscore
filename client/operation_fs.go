package client

import "github.com/kernel-ai/koscore/client/packets/message"

// 移动群文件
func (m *QQClient) MoveGroupFile(groupUin uint64, file_id string, parent_dir, target_dir string) error {
	pkt, e := message.BuildGroupFSMovePacket(groupUin, file_id, parent_dir, target_dir)
	if e != nil {
		return e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return e
	}
	return message.ParseGroupFSMovePacket(pkt.Data)
}

// 删除群文件
func (m *QQClient) DeleteGroupFile(groupUin uint64, file_id string) error {
	pkt, e := message.BuildGroupFSDeletePacket(groupUin, file_id)
	if e != nil {
		return e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return e
	}
	return message.ParseGroupFSDeletePacket(pkt.Data)
}
