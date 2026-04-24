package client

import (
	"errors"

	pkt_msg "github.com/kernel-ai/koscore/client/packets/message"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/oidb"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils/types"
)

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
func (m *QQClient) FetchCookies(domains []string) (types.MapSS, error) {
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

// 获取对应群的所有群成员信息，使用token可以获取下一页的群成员信息
func (m *QQClient) FetchGroupMembers(group_uin uint64, token []byte) ([]*entity.GroupMember, []byte, error) {
	pkt, err := oidb.BuildFetchGroupMembersPacket(group_uin, token)
	if err != nil {
		return nil, nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, nil, err
	}
	return oidb.ParseFetchGroupMembersPacket(pkt.Data)
}

// 获取群信息 => FetchGroupInfo strange 是否陌生群聊
func (m *QQClient) FetchGroupExtra(group_uin uint64, strange bool) (*operation.FetchGroupExtraResponseInfoResult, error) {
	pkt, err := oidb.BuildFetchGroupExtraPacket(group_uin, strange)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	return oidb.ParseFetchGroupExtraPacket(pkt.Data)
}

// 获取加群请求信息 => GetGroupSystemMessages
func (m *QQClient) FetchGroupNotice(start, count uint64, is_filtered bool, gin ...uint64) (*entity.GroupSystemMessages, error) {
	pkt, err := oidb.BuildFetchGroupNotificationsPacket(count, start, is_filtered)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	ret, err := oidb.ParseFetchGroupNotificationsPacket(&m.cache, is_filtered, pkt.Data, gin...)
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
func (m *QQClient) SetGroupRequest(is_filtered bool, operate entity.GroupRequestOperate, sequence uint64, typ uint32, group_uin uint64, msg string) error {
	pkt, err := oidb.BuildSetGroupRequestPacket(is_filtered, operate, sequence, typ, group_uin, msg)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 上传合并转发消息 group_uin should be the group number where the uploader is located or 0 (c2c)
func (m *QQClient) UploadForwardMsg(forward *message.ForwardMessage, group_uin uint64) (*message.ForwardMessage, error) {
	pkt := pkt_msg.BuildMultiMsgUploadPacket(m.session.Info.Uid, group_uin, m.BuildFakeMessage(forward.Nodes), m.version)
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
func (m *QQClient) FetchForwardMsg(resid string, is_group bool) (msg *message.ForwardMessage, err error) {
	if resid == "" {
		return msg, errors.New("empty resid")
	}
	pkt, err := m.sendOidbPacketAndWait(pkt_msg.BuildMultiMsgDownloadPcket(m.session.Info.Uid, resid, is_group, m.version))
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
			ms := message.ParseGroupMessage(m.session.Info.Uin, b)
			m.PreprocessGroupMessageEvent(ms)
			forward.Nodes[idx].GroupId = ms.GroupUin
			forward.Nodes[idx].SenderName = ms.Sender.CardName
			forward.Nodes[idx].Message = ms.Elements
		} else {
			ms := message.ParsePrivateMessage(m.session.Info.Uin, b)
			m.PreprocessPrivateMessageEvent(ms)
			forward.Nodes[idx].SenderName = ms.Sender.Nickname
			forward.Nodes[idx].Message = ms.Elements
		}
	}

	return forward, nil
}

// 设置群消息表态
func (m *QQClient) SetGroupReaction(group_uin, sequence uint64, code string, is_Add bool) error {
	pkt, err := oidb.BuildGroupReactionPacket(group_uin, sequence, code, is_Add)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 群全员禁言
func (m *QQClient) SetGroupGlobalMute(group_uin uint64, is_mute bool) error {
	pkt, err := oidb.BuildSetGroupGlobalMutePacket(group_uin, is_mute)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.ParseSetGroupGlobalMutePacket(pkt.Data)
}

// 设置群聊名称
func (m *QQClient) SetGroupName(group_uin uint64, name string) error {
	pkt, err := oidb.BuildSetGroupNamePacket(group_uin, name)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 设置群成员昵称
func (m *QQClient) SetGroupMemberName(group_uin, uin uint64, name string) error {
	uid, err := m.GetUid(uin, group_uin)
	if err != nil {
		return err
	}
	pkt, err := oidb.BuildSetGroupMemberNamePacket(group_uin, uid, name)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	if err = oidb.CheckError(pkt.Data); err != nil {
		return err
	}
	if g := m.GetCachedMemberInfo(uin, group_uin); g != nil {
		g.MemberCard = name
		m.cache.RefreshGroupMember(group_uin, g)
	}
	return nil
}

// 设置群成员专属头衔
func (m *QQClient) SetGroupMemberSpecialTitle(group_uin, uin uint64, title string) error {
	uid, err := m.GetUid(uin, group_uin)
	if err != nil {
		return err
	}
	pkt, err := oidb.BuildSetGroupMemberSpecialTitlePacket(group_uin, uid, title)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 退出群聊
func (m *QQClient) SetGroupLeave(group_uin uint64) error {
	pkt, err := oidb.BuildSetGroupLeavePacket(group_uin)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 戳一戳群友
func (m *QQClient) GroupPoke(group_uin, uin uint64) error {
	pkt, err := oidb.BuildNudgePacket(group_uin, uin)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 戳一戳好友
func (m *QQClient) FriendPoke(uin uint64) error {
	pkt, err := oidb.BuildNudgePacket(0, uin)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}
