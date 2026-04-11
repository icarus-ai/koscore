package client

import (
	"github.com/kernel-ai/koscore/client/packets/message"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
)

// 获取私聊图片下载url
func (m *QQClient) GetPrivateImageURL(node *oidb.IndexNode) (string, error) {
	pkt, e := message.BuildPrivateImageDownloadPacket(m.Uid(), node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return message.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取群聊图片下载url
func (m *QQClient) GetGroupImageURL(groupUin uint64, node *oidb.IndexNode) (string, error) {
	pkt, e := message.BuildGroupImageDownloadPacket(groupUin, node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return message.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取私聊语音下载url
func (m *QQClient) GetPrivateRecordURL(node *oidb.IndexNode) (string, error) {
	pkt, e := message.BuildPrivateRecordDownloadPacket(m.Uid(), node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return message.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取群聊语音下载url
func (m *QQClient) GetGroupRecordURL(groupUin uint64, node *oidb.IndexNode) (string, error) {
	pkt, e := message.BuildGroupRecordDownloadPacket(groupUin, node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return message.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取私聊视频下载链接
func (m *QQClient) GetPrivateVideoURL(node *oidb.IndexNode) (string, error) {
	pkt, e := message.BuildPrivateVideoDownloadPacket(m.Uid(), node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return message.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取群聊视频下载链接
func (m *QQClient) GetGroupVideoURL(groupUin uint64, node *oidb.IndexNode) (string, error) {
	pkt, e := message.BuildGroupVideoDownloadPacket(groupUin, node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return message.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取私聊文件下载链接
func (m *QQClient) GetPrivateFileURL(file_uuid string, file_hash string) (string, error) {
	pkt, e := message.BuildPrivateFSDownloadPacket(m.Uid(), file_uuid, file_hash)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return message.ParsePrivateFSDownloadPacket(pkt.Data)
}

// 获取群文件下载链接
func (m *QQClient) GetGroupFileURL(groupUin uint64, file_id string) (string, error) {
	pkt, e := message.BuildGroupFSDownloadPacket(groupUin, file_id)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return message.ParseGroupFSDownloadPacket(pkt.Data)
}

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
