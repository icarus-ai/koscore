package client

import (
	"regexp"
	"strconv"

	"github.com/kernel-ai/koscore/client/event"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/notify"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
	"github.com/kernel-ai/koscore/utils/proto"
)

// commit 53bc9c04123967aa745e216d00c14755e61b969c
// miraigo 旧代码 可能有bug

func S_NUM[T uint32 | uint64](val string) T {
	v, _ := strconv.Atoi(val)
	return T(v)
}

func (m *QQClient) decodeOlPushServicePacket_group_notify_msg_0x210(subType int32, pkg *message.CommonMessage) error {
	msg_content := pkg.MessageBody.MsgContent
	switch subType {
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
		m.grayTipProcessor(0, pkg, nil, pb)
		/*
			case 0x15D: // 上下线
				if fn, ok := fnmap_decoders_data["0x210_0x15D"]; ok { _, err = fn("0x210_0x15D", msg_content) }
				return err
		*/
	case 226: // 好友验证消息，申请，同意都有
	case 179: // new friend 主动加好友且对方同意
		pb, err := proto.Unmarshal[notify.NewFriend](msg_content)
		if err != nil {
			return err
		}
		ev := event.ParseNewFriendEvent(pb)
		_ = m.ResolveUin(ev)
		m.Events.NewFriend.dispatch(m, ev)

	//case 38 : // group member notice
	//case 212: // group kick notice
	case 321: // friend recall poke
		pb, err := proto.Unmarshal[notify.FriendRecallPokeInfo](msg_content)
		if err != nil {
			return err
		}
		ev := &event.FriendPokeRecallEvent{
			PeerUid: pb.PeerUid.Unwrap(),
		}
		ev.OperatorUid = pb.OperatorUid.Unwrap()
		ev.TipsSeqId = pb.TipsSeqId.Unwrap()
		_ = m.ResolveUin(ev)

		m.Events.FriendPokeRecall.dispatch(m, ev)
	default:
		m.LOGD("unknown subtype %d of type 0x210, proto data: %x", subType, msg_content)
	}
	return nil
}

func (m *QQClient) decodeOlPushServicePacket_group_notify_msg_0x2DC(subType int32, pkg *message.CommonMessage) error {
	msg_content := pkg.MessageBody.MsgContent
	switch subType {
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
			m.grayTipProcessor(group_uin, pkg, pb, pb.GeneralGrayTip)
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
			ev := &event.GroupPokeRecallEvent{
				GroupEvent: event.GroupEvent{
					GroupUin: uint64(pb.GroupRecallNudge.GroupUin.Unwrap()),
					UserUid:  pb.GroupRecallNudge.OperatorUid.Unwrap(),
				},
			}
			ev.TipsSeqId = pb.GroupRecallNudge.TipsSeqId.Unwrap()
			_ = m.ResolveUin(ev)
			m.Events.GroupPokeRecall.dispatch(m, ev)
		}

		if pb.Reaction != nil {
			// pb.SubType == 13 // group reaction
			ev := event.ParseGroupReactionEvent(pb)
			_ = m.ResolveUin(ev)
			m.Events.GroupReaction.dispatch(m, ev)
			return nil
		}
	}
	m.LOGD("Unsupported group event, subType: %v, proto data: %x", subType, msg_content)
	return nil
}

// grayTipProcessor 提取出来专门用于处理群内 notify tips
func (m *QQClient) grayTipProcessor(groupUin uint64, pkg *message.CommonMessage, nmsg *notify.NotifyMessageBody, graytip *notify.GeneralGrayTipInfo) {
	if graytip == nil {
		if nmsg == nil {
			return
		}
		graytip = nmsg.GeneralGrayTip
	}

	//fmt.Printf("notify gray tip: busi_type: %d templ_id: %d\n", graytip.BusiType.Unwrap(), graytip.TemplId.Unwrap())

	if graytip.BusiType.Unwrap() == 12 && graytip.BusiId.Unwrap() == 1061 {
		var sender, receiver uint64
		suffix, action := "", "戳了戳"
		for _, data := range graytip.MsgTemplParam {
			switch data.Name.Unwrap() {
			case "uin_str1":
				sender = S_NUM[uint64](data.Value.Unwrap())
			case "uin_str2":
				receiver = S_NUM[uint64](data.Value.Unwrap())
			case "suffix_str":
				suffix = data.Value.Unwrap()
			case "action_str", "alt_str1":
				action = data.Value.Unwrap()
				//case "action_img_url":
			}
		}
		if sender != 0 {
			if receiver == 0 {
				receiver = m.Uin()
			}
			poke_event := event.PokeEventBase{
				Action:    action,
				Sender:    sender,
				Receiver:  receiver,
				Suffix:    suffix,
				MsgSeq:    nmsg.MsgSequence.Unwrap(),
				MsgTime:   pkg.ContentHead.Time.Unwrap(),
				TipsSeqId: nmsg.TipsSeqId.Unwrap(),
			}
			if groupUin == 0 {
				ev := &event.FriendPokeEvent{
					PokeEventBase: poke_event,
					PeerUin:       uint64(pkg.RoutingHead.FromUin.Unwrap()),
				}
				//_ = m.ResolveUin(ev)
				m.Events.FriendPoke.dispatch(m, ev)
			} else {
				ev := &event.GroupPokeEvent{
					PokeEventBase: poke_event,
					GroupEvent: event.GroupEvent{
						GroupUin: groupUin,
						UserUin:  sender,
					},
				}
				//_ = m.ResolveUin(ev)
				m.Events.GroupPoke.dispatch(m, ev)
			}
		}
		return
	}

	if graytip.TemplId.Unwrap() == 10036 || graytip.TemplId.Unwrap() == 10038 { // 群签到/打卡
		sign := &event.GroupSignEvent{GroupUin: groupUin}
		for _, templ := range graytip.MsgTemplParam {
			switch templ.Name.Unwrap() {
			case "mqq_uin":
				sign.Uin = S_NUM[uint64](templ.Value.Unwrap())
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
				sign.Rank = S_NUM[uint32](matches[1][1])
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
	honor_event.GroupUin = groupUin
	for _, templ := range graytip.MsgTemplParam {
		switch templ.Name.Unwrap() {
		case "nick":
			honor_event.Nick = templ.Value.Unwrap()
		case "uin":
			honor_event.Uin = S_NUM[uint64](templ.Value.Unwrap())
		case "uin_last":
			honor_event.Previous = S_NUM[uint32](templ.Value.Unwrap())
		}
	}
	m.Events.GroupNotify.dispatch(m, honor_event)
}
