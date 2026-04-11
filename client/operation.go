package client

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	pkt_msg "github.com/kernel-ai/koscore/client/packets/message"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/oidb"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/websso"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils"
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

// 获取ClientKey
func (m *QQClient) FetchClientKey() (*oidb.FetchClientKeyRep, error) {
	pkt, e := oidb.BuildFetchClientKeyPacket()
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseFetchClientKeyPacket(pkt.Data)
}

// 获取cookies
func (m *QQClient) FetchCookies(domains []string) (*oidb.FetchCookiesRsp, error) {
	pkt, e := oidb.BuildFetchCookiesPacket(domains)
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseFetchCookiesPacket(pkt.Data)
}

// 获取用户信息
func (m *QQClient) FetchStrangerUin(uin uint64) (*entity.User, error) {
	pkt, err := oidb.BuildFetchStrangerPacket(uin, 2)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	return oidb.ParseFetchStrangerPacket(pkt.Data)
}

// 获取用户信息
func (m *QQClient) FetchStrangerUid(uid string) (*entity.User, error) {
	pkt, err := oidb.BuildFetchStrangerPacket(uid, 2)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	return oidb.ParseFetchStrangerPacket(pkt.Data)
}

// 获取好友列表信息，使用token可以获取下一页的群成员信息
func (m *QQClient) FetchFriends(token []byte) (*oidb.FetchFriendsRsp, error) {
	pkt, err := oidb.BuildFetchFriendsPacket(token)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	return oidb.ParseFetchFriendsPacket(pkt.Data)
}

// 获取单向好友列表
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/web.go#L23
func (m *QQClient) GetUnidirectionalFriendList() ([]*entity.User, error) {
	rsp, err := m.webSsoRequest("ti.qq.com", "OidbSvc.0xe17_0", fmt.Sprintf(`{"uint64_uin":%v,"uint64_top":0,"uint32_req_num":99,"bytes_cookies":""}`, m.UIN()))
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

// 获取所有已加入的群的信息
func (m *QQClient) FetchGroups() ([]*entity.Group, error) {
	pkt, err := oidb.BuildFetchGroupsPacket()
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	return oidb.ParseFetchGroupsPacket(pkt.Data)
}

// OLD_CODE
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

// 获取对应群的所有群成员信息，使用token可以获取下一页的群成员信息
func (m *QQClient) FetchGroupMembers(groupUin uint64, token []byte) ([]*entity.GroupMember, []byte, error) {
	pkt, err := oidb.BuildFetchGroupMembersPacket(groupUin, token)
	if err != nil {
		return nil, nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, nil, err
	}
	return oidb.ParseFetchGroupMembersPacket(pkt.Data)
}

// 获取群信息 => FetchGroupInfo strange 是否陌生群聊
func (m *QQClient) FetchGroupExtra(groupUin uint64, strange bool) (*operation.FetchGroupExtraResponseInfoResult, error) {
	pkt, err := oidb.BuildFetchGroupExtraPacket(groupUin, strange)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	return oidb.ParseFetchGroupExtraPacket(pkt.Data)
}

// 获取加群请求信息 => GetGroupSystemMessages
func (m *QQClient) FetchGroupNotice(start, count uint64, isfiltered bool, gin ...uint64) (*entity.GroupSystemMessages, error) {
	pkt, err := oidb.BuildFetchGroupNotificationsPacket(count, start, isfiltered)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	ret, err := oidb.ParseFetchGroupNotificationsPacket(&m.cache, isfiltered, pkt.Data, gin...)
	if err != nil {
		return nil, err
	}

	for _, req := range ret.InvitedRequests {
		if g, err := m.FetchGroupExtra(req.GroupUin, true); err == nil {
			req.GroupName = g.GroupName.Unwrap()
		}
		if u, err := m.FetchStrangerUid(req.InvitorUid); err == nil {
			req.InvitorNick = u.Nickname
		}
	}
	for _, req := range ret.JoinRequests {
		if g, err := m.FetchGroupExtra(req.GroupUin, false); err == nil {
			req.GroupName = g.GroupName.Unwrap()
		}
		if u, err := m.FetchStrangerUid(req.TargetUid); err == nil {
			req.TargetNick = u.Nickname
		}
	}
	return ret, nil
}

// 处理加群请求
func (m *QQClient) SetGroupRequest(isFiltered bool, operate entity.GroupRequestOperate, sequence uint64, typ uint32, groupUin uint64, message string) error {
	pkt, err := oidb.BuildSetGroupRequestPcket(isFiltered, operate, sequence, typ, groupUin, message)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.ParseSetGroupRequestPaacket(pkt.Data)
}

// OLD_CODE
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

// 上传合并转发消息 groupUin should be the group number where the uploader is located or 0 (c2c)
func (m *QQClient) UploadForwardMsg(forward *message.ForwardMessage, groupUin uint64) (*message.ForwardMessage, error) {
	pkt := pkt_msg.BuildMultiMsgUploadPacket(m.Uid(), groupUin, m.BuildFakeMessage(forward.Nodes), m.version)
	pkt, err := m.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	pasted, err := pkt_msg.ParseMultiMsgUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}
	forward.ResId = pasted.ResId.Unwrap()
	return forward, nil
}

// 获取合并转发消息
func (m *QQClient) FetchForwardMsg(resid string, isgroup bool) (msg *message.ForwardMessage, err error) {
	if resid == "" {
		return msg, errors.New("empty resid")
	}
	pkt, err := m.sendOidbPacketAndWait(pkt_msg.BuildMultiMsgDownloadPcket(m.Uid(), resid, isgroup, m.version))
	if err != nil {
		return nil, err
	}
	rsp, e := pkt_msg.ParseMultiMsgDownloadPacket(pkt.Data)
	if e != nil {
		return nil, e
	}

	forward := &message.ForwardMessage{ResId: resid}
	forward.Nodes = make([]*message.ForwardNode, len(rsp))

	for idx, b := range rsp {
		forward.Nodes[idx] = &message.ForwardNode{
			SenderId: uint64(b.RoutingHead.FromUin.Unwrap()),
			Time:     uint32(b.ContentHead.Time.Unwrap()),
		}
		if forward.IsGroup = b.RoutingHead.Group != nil; forward.IsGroup {
			ms := message.ParseGroupMessage(m.UIN(), b)
			m.PreprocessGroupMessageEvent(ms)
			forward.Nodes[idx].GroupId = ms.GroupUin
			forward.Nodes[idx].SenderName = ms.Sender.CardName
			forward.Nodes[idx].Message = ms.Elements
		} else {
			ms := message.ParsePrivateMessage(m.UIN(), b)
			m.PreprocessPrivateMessageEvent(ms)
			forward.Nodes[idx].SenderName = ms.Sender.Nickname
			forward.Nodes[idx].Message = ms.Elements
		}
	}

	return forward, nil
}
