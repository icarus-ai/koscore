package client

import (
	"regexp"

	"github.com/kernel-ai/koscore/client/event"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/notify"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
	"github.com/kernel-ai/koscore/utils/proto"
)

// commit 53bc9c04123967aa745e216d00c14755e61b969c
// miraigo 旧代码 可能有bug

func (m *QQClient) decodeOlPushServicePacket_group_notify_msg_0x210(sub_type int32, pkg *message.CommonMessage) error {
	msg_content := pkg.MessageBody.MsgContent
	switch sub_type {
	case 35: // friend request notice
		pb, err := proto.Unmarshal[notify.FriendRequest](msg_content)
		if err != nil {
			return err
		}
		if pb.Info == nil {
			break
		}
		ev := event.ParseFriendRequestNotice(pb.Info)
		user, _ := m.FetchStrangerUid(ev.SourceUid)
		if user != nil {
			ev.SourceUin = user.Uin
			ev.SourceNick = user.Nickname
		}
		//_ = m.ResolveUin(ev)
		m.Events.NewFriendRequest.dispatch(m, ev)

	case 138: // friend recall
		pb, err := proto.Unmarshal[notify.FriendRecall](msg_content)
		if err != nil {
			return err
		}
		ev := event.ParseFriendRecallEvent(pb.Info)
		_ = m.ResolveUin(ev)
		m.Events.FriendRecall.dispatch(m, ev)

	case 39: // friend rename
		pb, err := proto.Unmarshal[notify.FriendRenameMsg](msg_content)
		if err != nil {
			return err
		}
		if pb.Body.Field2 == 20 { // friend name update
			ev := event.ParseFriendRenameEvent(pb)
			_ = m.ResolveUin(ev)
			m.Events.Rename.dispatch(m, ev)
		} // 40 grp name

	case 29:
		pb, err := proto.Unmarshal[notify.SelfRenameMsg](msg_content)
		if err != nil {
			return err
		}
		m.Events.Rename.dispatch(m, event.ParseSelfRenameEvent(pb, m.session.Info))

	case 290: // greyTip
		pb, err := proto.Unmarshal[notify.GeneralGrayTipInfo](msg_content)
		if err != nil {
			return err
		}
		m.gray_tip_processor(0, pkg, nil, pb)

	case 226: // 好友验证消息，申请，同意都有
	case 179: // new friend 主动加好友且对方同意
		pb, err := proto.Unmarshal[notify.NewFriend](msg_content)
		if err != nil {
			return err
		}
		ev := event.ParseNewFriendEvent(pb)
		_ = m.ResolveUin(ev)
		m.Events.NewFriend.dispatch(m, ev)

	case 38: // group member notice
	//case 212: // group kick notice
	case 321: // friend recall poke
		pb, err := proto.Unmarshal[notify.FriendRecallPokeInfo](msg_content)
		if err != nil {
			return err
		}
		poke_recall := &event.FriendPokeRecallEvent{
			PeerUid: pb.PeerUid.Unwrap(),
		}
		poke_recall.OperatorUid = pb.OperatorUid.Unwrap()
		poke_recall.TipsSeqId = pb.TipsSeqId.Unwrap()
		_ = m.ResolveUin(poke_recall)
		m.Events.FriendNotify.dispatch(m, poke_recall)

	case 0x15D: // 上下线
		if fn, ok := m.decoders[event.DECODE_CMD_0210_015D]; ok {
			fn(msg_content)
		}

	default:
		m.LOGD("unknown sub_type %d of type 0x210, proto data: %x", sub_type, msg_content)
	}
	return nil
}

func (m *QQClient) decodeOlPushServicePacket_group_notify_msg_0x2DC(sub_type int32, pkg *message.CommonMessage) error {
	msg_content := pkg.MessageBody.MsgContent
	switch sub_type {
	case 12: // mute
		pb, err := proto.Unmarshal[notify.GroupMute](msg_content)
		if err != nil {
			return err
		}
		ev := event.ParseGroupMuteEvent(pb)
		_ = m.ResolveUin(ev)
		m.Events.GroupMute.dispatch(m, ev)
		return nil
	case 0x10, 0x11, 0x14, 0x15: // group notify msg
		reader := binary.NewReader(msg_content)
		group_uin := uint64(reader.ReadU32()) // group uin
		reader.SkipBytes(1)                   // unknown byte

		pb_data := reader.ReadLengthBytes(prefix.Int16 | prefix.LengthOnly)
		pb, err := proto.Unmarshal[notify.NotifyMessageBody](pb_data)
		if err != nil {
			return err
		}

		if pb.Recall != nil {
			group_uin = uint64(pb.GroupUin.Unwrap())
			operator_uid := pb.Recall.OperatorUid.Unwrap()
			for _, rm := range pb.Recall.RecallMessages {
				if rm.Type.Unwrap() == 2 {
					continue
				}
				ev := event.ParseGroupRecallEvent(group_uin, operator_uid, rm)
				_ = m.ResolveUin(ev)
				m.Events.GroupRecall.dispatch(m, ev)
			}
			return nil
		}

		if pb.GeneralGrayTip != nil {
			m.gray_tip_processor(group_uin, pkg, pb, pb.GeneralGrayTip)
			return nil
		}

		// ***** ? HAS_OLD_CODE BIGIN *****
		if len(pb.RedTips) != 0 {
			tips, err := proto.Unmarshal[notify.RedGrayTipsInfo](pb.RedTips)
			if err != nil {
				return nil
			}
			if tips.LuckyFlag == 1 { // 运气王提示
				m.Events.GroupNotify.dispatch(m, &event.GroupRedBagLuckyKingNotifyEvent{
					GroupUin:  group_uin,
					Sender:    tips.SenderUin,
					LuckyKing: tips.LuckyUin,
				})
			}
			return nil
		}

		if len(pb.EventParam) != 0 { // msgGrayTipProcessor 用于处理群内 aio notify tips
			tips, err := proto.Unmarshal[notify.AIOGrayTipsInfo](pb.EventParam)
			if err != nil {
				return nil
			}
			if len(tips.Content) > 0 {
				switch pb.SubType.Unwrap() {
				case 6: // GroupMemberSpecialTitle
					m.Events.MemberSpecialTitleUpdated.dispatch(m, event.ParseGroupMemberSpecialTitleUpdatedEvent(tips, group_uin))
				case 12: // group name update
					ev := event.ParseGroupNameUpdatedEvent(group_uin, pb.OperatorUid.Unwrap(), tips.Content)
					_ = m.ResolveUin(ev)
					m.Events.GroupNameUpdated.dispatch(m, ev)
				}
			}
		}
		// ***** HAS_OLD_CODE END *****

		if pb.EssenceMessage != nil {
			// pb.SubType == 27 // essence
			ev := event.ParseGroupDigestEvent(pb.EssenceMessage)
			//_ = m.ResolveUin(ev)
			m.Events.GroupDigest.dispatch(m, ev)
			return nil
		}

		if pb.GroupRecallNudge != nil {
			// pb.SubType == 32 // group recall poke
			poke_recall := &event.GroupPokeRecallEvent{
				GroupEvent: event.GroupEvent{
					GroupUin: uint64(pb.GroupRecallNudge.GroupUin.Unwrap()),
					UserUid:  pb.GroupRecallNudge.OperatorUid.Unwrap(),
				},
				PokeRecallEventBase: event.PokeRecallEventBase{
					TipsSeqId: pb.GroupRecallNudge.TipsSeqId.Unwrap(),
				},
			}
			_ = m.ResolveUin(poke_recall)
			m.Events.GroupNotify.dispatch(m, poke_recall)
			return nil
		}

		if pb.Reaction != nil {
			// pb.SubType == 13 // group reaction
			ev := event.ParseGroupReactionEvent(pb)
			_ = m.ResolveUin(ev)
			m.Events.GroupReaction.dispatch(m, ev)
			return nil
		}
	}

	m.LOGD("unknown sub_type %d of type 0x2DC, proto data: %x", sub_type, msg_content)
	return nil
}

// 提取出来专门用于处理群内 notify tips
func (m *QQClient) gray_tip_processor(group_uin uint64, pkg *message.CommonMessage, notify_msg *notify.NotifyMessageBody, graytip *notify.GeneralGrayTipInfo) {
	//fmt.Printf("notify gray tip: busi_type: %d templ_id: %d\n", graytip.BusiType.Unwrap(), graytip.TemplId.Unwrap())

	if graytip.BusiType.Unwrap() == 12 && graytip.BusiId.Unwrap() == 1061 {
		var sender, receiver uint64
		suffix, action := "", "戳了戳"
		for _, data := range graytip.MsgTemplParam {
			switch data.Name.Unwrap() {
			case "uin_str1":
				sender = utils.S_NUM[uint64](data.Value.Unwrap())
			case "uin_str2":
				receiver = utils.S_NUM[uint64](data.Value.Unwrap())
			case "suffix_str":
				suffix = data.Value.Unwrap()
			case "action_str", "alt_str1":
				action = data.Value.Unwrap()
				//case "action_img_url":
			}
		}
		if sender != 0 {
			if receiver == 0 {
				receiver = m.session.Info.Uin
			}
			poke_event := event.PokeEventBase{
				Action:    action,
				Sender:    sender,
				Receiver:  receiver,
				Suffix:    suffix,
				MsgTime:   pkg.ContentHead.Time.Unwrap(),
				TipsSeqId: graytip.TipsSeqId.Unwrap(),
			}
			// ??? 未测试
			if notify_msg != nil {
				poke_event.TipsSeqId = notify_msg.TipsSeqId.Unwrap()
				poke_event.MsgSeq = notify_msg.MsgSequence.Unwrap()
			} else if graytip.MsgInfo != nil {
				poke_event.MsgSeq = graytip.MsgInfo.Sequence.Unwrap()
			}
			if group_uin == 0 {
				poke := &event.FriendPokeEvent{
					PokeEventBase: poke_event,
					PeerUin:       uint64(pkg.RoutingHead.FromUin.Unwrap()),
				}
				//_ = m.ResolveUin(ev)
				m.Events.FriendNotify.dispatch(m, poke)
			} else {
				poke := &event.GroupPokeEvent{
					PokeEventBase: poke_event,
					GroupEvent: event.GroupEvent{
						GroupUin: group_uin,
						UserUin:  sender,
					},
				}
				//_ = m.ResolveUin(ev)
				m.Events.GroupNotify.dispatch(m, poke)
			}
		}
		return
	}

	if graytip.TemplId.Unwrap() == 10036 || graytip.TemplId.Unwrap() == 10038 { // 群签到/打卡
		sign := &event.GroupSignEvent{GroupUin: group_uin}
		for _, templ := range graytip.MsgTemplParam {
			switch templ.Name.Unwrap() {
			case "mqq_uin":
				sign.Uin = utils.S_NUM[uint64](templ.Value.Unwrap())
			case "mqq_nick":
				sign.Nick = templ.Value.Unwrap()
			case "rank_img":
				sign.RankIMG = templ.Value.Unwrap()
			case "user_sign":
				sign.Sign = templ.Value.Unwrap()
				re := regexp.MustCompile(`今日第(\d+)个打卡`)
				matches := re.FindAllStringSubmatch(sign.Sign, -1)
				if len(matches) != 2 {
					return
				}
				sign.Rank = utils.S_NUM[uint32](matches[1][1])
			}
		}
		m.Events.GroupNotify.dispatch(m, sign)
		return
	}

	// 新增: 更多的龙王TemplId和HonorType(see http_api.go)
	// See https://github.com/mamoe/mirai/blob/d000f2ea0f2ab7a9de3b0b346d63b44f02d240ca/mirai-core/src/commonMain/kotlin/network/notice/group/GroupNotificationProcessor.kt#L324

	var honor_event *event.MemberHonorChangedNotifyEvent
	switch graytip.TemplId.Unwrap() {
	case 1053, 1054, 1103, 10093, 10094:
		honor_event = &event.MemberHonorChangedNotifyEvent{Honor: event.Talkative}
	case 1052, 1129:
		honor_event = &event.MemberHonorChangedNotifyEvent{Honor: event.Performer}
	case 1055:
		honor_event = &event.MemberHonorChangedNotifyEvent{Honor: event.Legend}
	case 1067:
		honor_event = &event.MemberHonorChangedNotifyEvent{Honor: event.Emotion}
	case 10111:
		honor_event = &event.MemberHonorChangedNotifyEvent{Honor: event.RED_PACKET}
	default:
		return
	}
	honor_event.GroupUin = group_uin
	for _, templ := range graytip.MsgTemplParam {
		switch templ.Name.Unwrap() {
		case "nick":
			honor_event.Nick = templ.Value.Unwrap()
		case "uin":
			honor_event.Uin = utils.S_NUM[uint64](templ.Value.Unwrap())
		case "uin_last":
			honor_event.Previous = utils.S_NUM[uint32](templ.Value.Unwrap())
		}
	}
	m.Events.GroupNotify.dispatch(m, honor_event)
}
