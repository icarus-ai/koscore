package client

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/message/message_type"
	"github.com/kernel-ai/koscore/client/packets/oidb"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/websso"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
)

// 获取Rkey
func (m *QQClient) FetchRkey() (entity.RKeyMap, error) {
	pkt, e := oidb.BuildFetchRKeyPacket()
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseFetchRKeyPacket(pkt.Data)
}

// 设置在线状态
func (m *QQClient) SetOnlineStatus(status operation.SetStatus) error {
	data, _ := proto.Marshal(&status)
	pkt, err := m.sendOidbPacketAndWait(message_type.AttributeSetStatus.NewSsoPacket(m.Session().GetAndIncreaseSequence(), data))
	if err != nil {
		return err
	}
	rsp, err := proto.Unmarshal[operation.SetStatusResponse](pkt.Data)
	if err != nil {
		return err
	}
	if rsp.Message != "set status success" {
		return fmt.Errorf("set status failed: %s", rsp.Message)
	}
	return nil
}

// 获取单向好友列表
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/web.go#L23
func (m *QQClient) GetUnidirectionalFriendList() ([]*entity.User, error) {
	rsp, err := m.webSsoRequest("ti.qq.com", "OidbSvc.0xe17_0", fmt.Sprintf(`{"uint64_uin":%v,"uint64_top":0,"uint32_req_num":99,"bytes_cookies":""}`, m.Uin()))
	if err != nil {
		return nil, err
	}
	return websso.ParseUnidirectionalFriendsPacket(utils.S2B(rsp))
}

// 删除单向好友
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/web.go#L62
func (m *QQClient) DeleteUnidirectionalFriend(uin uint64) error {
	rsp, err := m.webSsoRequest("ti.qq.com", "OidbSvc.0x5d4_0", fmt.Sprintf(`{"uin_list":[%v]}`, uin))
	if err != nil {
		return err
	}
	webRsp := &struct {
		ErrorCode int32 `json:"ErrorCode"`
	}{}
	if err = json.Unmarshal(utils.S2B(rsp), webRsp); err != nil {
		return errors.Wrap(err, "unmarshal json error")
	}
	if webRsp.ErrorCode != 0 {
		return fmt.Errorf("web sso request error: %v", webRsp.ErrorCode)
	}
	return nil
}

// 获取对应群的群成员信息
func (m *QQClient) FetchGroupMember(groupUin, memberUin uint64) (*entity.GroupMember, error) {
	pkt, err := oidb.BuildFetchGroupMemberPacket(groupUin, m.GetUid(memberUin, groupUin))
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	return oidb.ParseFetchGroupMemberPacket(pkt.Data)
}

// 发送群聊打卡消息
func (m *QQClient) SendGroupSign(groupUin uint64) (*oidb.BotGroupClockInResult, error) {
	pkt, e := oidb.BuildGroupSignPacket(m.Uin(), groupUin, m.version.CurrentVersion)
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseGroupSignResp(pkt.Data)
}

// 获取剩余@全员次数
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

// 通过TX服务器检查URL安全性
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/security.go#L24
func (m *QQClient) CheckURLSafely(url string) (oidb.URLSecurityLevel, error) {
	pkt, err := oidb.BuildURLCheckRequest(m.Uin(), url)
	if err != nil {
		return oidb.URLSecurityLevelUnknown, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return oidb.URLSecurityLevelUnknown, err
	}
	return oidb.ParseURLCheckResponse(pkt.Data)
}

// 图片识别 有些域名的图可能无法识别，需要重新上传到tx服务器并获取图片下载链接
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

// 获取AI语音角色列表
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

// 发送群AI语音
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

// 处理好友请求
func (m *QQClient) SetFriendRequest(accept bool, target_uid string) error {
	pkt, err := oidb.BuildSetFriendRequestPacket(accept, target_uid)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.ParseSetFriendRequestPacket(pkt.Data)
}
