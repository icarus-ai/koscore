package client

import (
	"github.com/kernel-ai/koscore/client/event"
	"github.com/kernel-ai/koscore/client/packets/message/message_type"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/notify"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/client/packets/system"
	"github.com/kernel-ai/koscore/client/packets/system/system_type"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/proto"

	pb_msg "github.com/kernel-ai/koscore/client/packets/pb/v2/message"
)

func (m *QQClient) message_handle_parse_packet(pkt *sso_type.SsoPacket) bool {
	switch pkt.Command {
	case message_type.AttributeMsgPush.Command:
		_ = m.message_handle_parse_push_message(pkt.Data)
	case system_type.AttributeKickNt.Command:
		ev := &event.Disconnected{}
		if rsp, e := system.ParseKickPacket(pkt); e == nil {
			ev.Message = rsp.TipsTitle.Unwrap() + " " + rsp.TipsInfo.Unwrap()
		} else {
			ev.Message = e.Error()
		}
		m.Events.Disconnected.dispatch(m, ev)
	case system_type.AttributePushParams.Command:
		pb, e := system.ParsePushParamsPacket(pkt)
		if e != nil {
			m.LOGD("parse push params packet: %v", e)
		} else {
			for _, v := range pb.OnlineDevices {
				m.LOGD("device name: %s", v.DeviceName.Unwrap())
			}
		}
	case system_type.AttributeInfoSyncPush.Command, system_type.AttributeHeartbeat.Command:
	default:
		//m.LOGD("PacketContext::DispatchPacket: sso: %v %s %d %d %X", e, sso.Command, sso.Sequence, len(sso.Data), sso.Data)
		if fn, ok := m.decoders[pkt.Command]; ok {
			fn(pkt.Data)
			return true
		}
		m.LOGD("message_handle_parse_packet: %d %s", pkt.Sequence, pkt.Command)
		if fn, ok := m.sso_context.handlers.LoadAndDelete(pkt.Sequence); ok {
			fn(pkt, nil)
		}
		return false
	}
	return true
}

func (m *QQClient) message_handle_parse_push_message(data []byte) error {
	msg_push, e := proto.Unmarshal[pb_msg.MsgPush](data)
	if e != nil {
		return e
	}

	common_msg := msg_push.CommonMessage
	msg_type := common_msg.ContentHead.Type.Unwrap()

	switch msg_type {
	case message_type.GROUP_MESSAGE:
		msg := message.ParseGroupMessage(m.session.Info.Uin, common_msg)
		m.PreprocessGroupMessageEvent(msg)
		if msg.Sender.Uin == m.session.Info.Uin {
			m.Events.SelfGroupMessage.dispatch(m, msg)
		} else {
			m.Events.GroupMessage.dispatch(m, msg)
		}
		return nil
	case message_type.PRIVATE_MESSAGE: // 166 for private msg, 208 for private record, 529 for private file
		msg := message.ParsePrivateMessage(m.session.Info.Uin, common_msg)
		m.PreprocessPrivateMessageEvent(msg)
		if msg.Sender.Uin == m.session.Info.Uin {
			m.Events.SelfPrivateMessage.dispatch(m, msg)
		} else {
			m.Events.PrivateMessage.dispatch(m, msg)
		}
		return nil
	case message_type.TEMP_MESSAGE:
		msg := message.ParseTempMessage(m.session.Info.Uin, common_msg)
		m.Events.TempMessage.dispatch(m, msg)
		return nil
	}

	if common_msg.MessageBody != nil && len(common_msg.MessageBody.MsgContent) > 0 {
		msg_content := common_msg.MessageBody.MsgContent
		sub_type := common_msg.ContentHead.SubType.Unwrap()
		switch msg_type {
		case message_type.GROUP_MEMBER_INCREASE_NOTICE:
			pb, e := proto.Unmarshal[notify.GroupChange](msg_content)
			if e != nil {
				return e
			}
			ev := event.ParseMemberIncreaseEvent(pb)
			_ = m.ResolveUin(ev)
			if ev.UserUin == m.session.Info.Uin { // bot 进群
				_ = m.RefreshAllGroupsInfo()
				m.Events.GroupJoin.dispatch(m, ev)
			} else {
				_ = m.RefreshGroupMemberCache(ev.GroupUin, ev.UserUin)
				m.Events.GroupMemberJoin.dispatch(m, ev)
			}
		case message_type.GROUP_MEMBER_DECREASE_NOTICE:
			pb, e := proto.Unmarshal[notify.GroupChange](msg_content)
			if e != nil {
				return e
			}
			switch pb.Type.Unwrap() {
			case 3, 131:
				// 3   KickSelf bot自身被踢出，Operator字段会是一个protobuf
				// 131 Kick
				op, e := proto.Unmarshal[notify.OperatorInfo](pb.Operator)
				if e != nil {
					return e
				}
				pb.Operator = utils.S2B(op.Operator.Uid.Unwrap())
				ev := event.ParseMemberDecreaseEvent(pb)
				_ = m.ResolveUin(ev)
				if ev.UserUin == m.session.Info.Uin {
					m.Events.GroupLeave.dispatch(m, ev)
				} else {
					m.Events.GroupMemberLeave.dispatch(m, ev)
				}
			case 130: // Exit
			}
		case message_type.GROUP_ADMIN_CHANGED_NOTICE:
			pb, e := proto.Unmarshal[notify.GroupAdmin](msg_content)
			if e != nil {
				return e
			}
			ev := event.ParseGroupMemberPermissionChanged(pb)
			_ = m.ResolveUin(ev)
			_ = m.RefreshGroupMemberCache(ev.GroupUin, ev.UserUin)
			m.Events.GroupMemberPermissionChanged.dispatch(m, ev)
		case message_type.GROUP_JOIN_NOTICE:
			pb, e := proto.Unmarshal[notify.GroupJoin](msg_content)
			if e != nil {
				return e
			}
			ev := event.ParseRequestJoinNotice(pb)
			_ = m.ResolveUin(ev)
			if user, _ := m.FetchStrangerUid(ev.UserUid); user != nil {
				ev.UserUin, ev.TargetNick = user.Uin, user.Nickname
			}

			commonRequests, reqErr := m.FetchGroupNotice(0, 20, false, ev.GroupUin)
			filteredRequests, freqErr := m.FetchGroupNotice(0, 20, true, ev.GroupUin)
			if reqErr == nil && freqErr == nil {
				for _, request := range append(commonRequests.JoinRequests, filteredRequests.JoinRequests...) {
					if request.TargetUid == ev.UserUid && !request.Checked {
						ev.RequestSeq = request.Sequence
						break
					}
				}
			}

			m.Events.GroupMemberJoinRequest.dispatch(m, ev)
		case message_type.GROUP_INVITE_NOTICE: // group invite notice
			pb, e := proto.Unmarshal[notify.GroupInvite0X57](msg_content)
			if e != nil {
				return e
			}
			ev := event.ParseInviteNotice(pb)
			_ = m.ResolveUin(ev)
			if group, err := m.FetchGroupExtra(ev.GroupUin, true); err == nil {
				ev.GroupName = group.GroupName.Unwrap()
			}
			if user, _ := m.FetchStrangerUid(ev.InvitorUid); user != nil {
				ev.InvitorUin, ev.InvitorNick = user.Uin, user.Nickname
			}

			commonRequests, reqErr := m.FetchGroupNotice(0, 20, false, ev.GroupUin)
			filteredRequests, freqErr := m.FetchGroupNotice(0, 20, true, ev.GroupUin)
			if reqErr == nil && freqErr == nil {
				for _, request := range append(commonRequests.InvitedRequests, filteredRequests.InvitedRequests...) {
					if !request.Checked {
						ev.RequestSeq = request.Sequence
						break
					}
				}
			}

			m.Events.GroupInvited.dispatch(m, ev)
		case message_type.Event0x20D: // group request invitation notice // GroupInviteProcessor
			// another in `RichTextMsgProcessor` for private send invitation card.
			pb, e := proto.Unmarshal[notify.Event0X20D](msg_content)
			if e != nil {
				return e
			}
			if pb.SubType.Unwrap() != 87 {
				return nil
			} // GroupInviteNotification
			body, e := proto.Unmarshal[notify.GroupInvite](pb.Body)
			if e != nil {
				return e
			}
			ev := event.ParseRequestInvitationNotice(body)
			_ = m.ResolveUin(ev)
			if user, _ := m.FetchStrangerUid(ev.UserUid); user != nil {
				ev.UserUin, ev.TargetNick = user.Uin, user.Nickname
			}
			m.Events.GroupMemberJoinRequest.dispatch(m, ev)
		case message_type.EVENT_FRIEND:
			return m.decodeOlPushServicePacket_group_notify_msg_0x210(sub_type, common_msg)
		case message_type.EVENT_GROUP:
			return m.decodeOlPushServicePacket_group_notify_msg_0x2DC(sub_type, common_msg)
		}
	}
	return nil
}

func (m *QQClient) PreprocessGroupMessageEvent(msg *message.GroupMessage) {
	if friend := m.GetCachedFriendInfo(msg.Sender.Uin); friend != nil {
		msg.Sender.IsFriend = true
		msg.Sender.Nickname = friend.Nickname
	} else if member := m.GetCachedMemberInfo(msg.GroupUin, msg.Sender.Uin); member != nil {
		msg.Sender.Nickname = member.Nickname
	}
	for _, elem := range msg.Elements {
		switch e := elem.(type) {
		case *message.ImageElement:
			if e.URL == "" {
				e.URL, _ = m.GetGroupImageURL(msg.GroupUin, e.MsgInfo.MsgInfoBody[0].Index)
			}
		case *message.VoiceElement:
			if url, err := m.GetGroupRecordURL(msg.GroupUin, e.Node); err == nil {
				e.URL = url
			}
		case *message.ShortVideoElement:
			if url, err := m.GetGroupVideoURL(msg.GroupUin, e.Node); err == nil {
				e.URL = url
			}
		case *message.FileElement:
			if url, err := m.GetGroupFileURL(msg.GroupUin, e.FileId); err == nil {
				e.FileURL = url
			}
		case *message.ForwardMessage:
			if e.Nodes == nil {
				if forward, err := m.FetchForwardMsg(e.ResId, true); err == nil {
					e.Nodes = forward.Nodes
				}
			}
		}
	}
}

func (m *QQClient) PreprocessPrivateMessageEvent(msg *message.PrivateMessage) {
	if friend := m.GetCachedFriendInfo(msg.Sender.Uin); friend != nil {
		msg.Sender.Nickname = friend.Nickname
	}
	for _, elem := range msg.Elements {
		switch e := elem.(type) {
		case *message.ImageElement:
			if e.URL == "" {
				e.URL, _ = m.GetPrivateImageURL(e.MsgInfo.MsgInfoBody[0].Index)
			}
		case *message.VoiceElement:
			if url, err := m.GetPrivateRecordURL(e.Node); err == nil {
				e.URL = url
			}
		case *message.ShortVideoElement:
			if url, err := m.GetPrivateVideoURL(e.Node); err == nil {
				e.URL = url
			}
		case *message.FileElement:
			if url, err := m.GetPrivateFileURL(e.FileUuid, e.FileHash); err == nil {
				e.FileURL = url
			}
		case *message.ForwardMessage:
			if e.Nodes == nil {
				if forward, err := m.FetchForwardMsg(e.ResId, false); err == nil {
					e.Nodes = forward.Nodes
				}
			}
		}
	}
}

func (m *QQClient) ResolveUin(g event.Iuid2uin) error {
	g.ResolveUin(func(uid string, groupUin ...uint64) uint64 { return m.GetUin(uid, groupUin...) })
	return nil
}
