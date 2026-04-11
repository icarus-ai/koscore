package client

import (
	"errors"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/oidb"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils/crypto"
)

// SendGroupSign 发送群聊打卡消息
func (m *QQClient) SendGroupSign(groupUin uint64) (*oidb.BotGroupClockInResult, error) {
	pkt, e := oidb.BuildGroupSignPacket(m.UIN(), groupUin, m.version.CurrentVersion)
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseGroupSignResp(pkt.Data)
}

// GetAtAllRemain 获取剩余@全员次数
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/group_msg.go#L68
func (m *QQClient) GetAtAllRemain(uin, groupUin uint64) (*oidb.AtAllRemainInfo, error) {
	pkt, err := oidb.BuildGetAtAllRemainRequest(uin, groupUin)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	return oidb.ParseGetAtAllRemainResponse(pkt.Data)
}

// CheckURLSafely 通过TX服务器检查URL安全性
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/security.go#L24
func (m *QQClient) CheckURLSafely(url string) (oidb.URLSecurityLevel, error) {
	pkt, err := oidb.BuildURLCheckRequest(m.UIN(), url)
	if err != nil {
		return oidb.URLSecurityLevelUnknown, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return oidb.URLSecurityLevelUnknown, err
	}
	return oidb.ParseURLCheckResponse(pkt.Data)
}

// ImageOcr 图片识别 有些域名的图可能无法识别，需要重新上传到tx服务器并获取图片下载链接
func (m *QQClient) ImageOcr(url string) (*oidb.OcrResponse, error) {
	if url == "" {
		return nil, errors.New("image url error")
	}
	pkt, e := oidb.BuildImageOcrRequestPacket(url)
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseImageOcrResp(pkt.Data)
}

// GetAiCharacters 获取AI语音角色列表
func (m *QQClient) GetAiCharacters(gin uint64, chatType entity.ChatType) (*entity.AiCharacterList, error) {
	if gin == 0 {
		gin = 42
	}
	pkt, e := oidb.BuildAiCharacterListService(gin, chatType)
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	result, e := oidb.ParseAiCharacterListService(pkt.Data)
	if e != nil {
		return nil, e
	}
	result.Type = chatType
	return result, nil
}

// SendGroupAiRecord 发送群AI语音
func (m *QQClient) SendGroupAiRecord(groupUin uint64, chatType entity.ChatType, voiceId, text string) (*message.VoiceElement, error) {
	pkt, e := oidb.BuildGroupAiRecordService(groupUin, voiceId, text, chatType, crypto.RandU32())
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseGroupAiRecordService(pkt.Data)
}
