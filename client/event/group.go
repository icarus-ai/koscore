package event

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/notify"
	"github.com/kernel-ai/koscore/utils"
)

type (
	// 通用群事件
	GroupEvent struct {
		GroupUin uint64
		// 触发事件的主体 详看各事件注释
		UserUin uint64
		UserUid string
	}

	// 群成员权限变更
	GroupMemberPermissionChanged struct {
		GroupEvent
		IsAdmin bool
	}

	// 群名变更
	GroupNameUpdated struct {
		GroupEvent
		NewName string
	}

	// 群内禁言事件 user为被禁言的成员
	GroupMute struct {
		GroupEvent
		OperatorUid string // when TargetUid is empty, mute all members
		OperatorUin uint64
		Duration    uint32 // Duration == math.MaxUint32 when means mute all
	}

	// 群内消息撤回 user为消息发送者
	GroupRecall struct {
		GroupEvent
		OperatorUid string
		OperatorUin uint64
		Sequence    uint64
		Time        int64
		Random      uint32
	}

	// 加群请求 user为请求加群的成员
	GroupMemberJoinRequest struct {
		GroupEvent
		TargetNick string
		InvitorUid string
		InvitorUin uint64
		Answer     string // 问题: (.*) 答案: (.*)
		RequestSeq uint64
	}

	// 群成员增加事件 user为新加群的成员
	GroupMemberIncrease struct {
		GroupEvent
		InvitorUid string
		InvitorUin uint64
		JoinType   uint32
	}

	// 群成员减少事件 user为退群的成员
	GroupMemberDecrease struct {
		GroupEvent
		OperatorUid string
		OperatorUin uint64
		ExitType    uint32
	}

	// 群精华消息 user为消息发送者 from miraigo
	GroupDigestEvent struct {
		GroupEvent
		MessageId         uint32
		InternalMessageId uint32
		OperationType     uint32 // 1 -> 设置精华消息, 2 -> 移除精华消息
		OperateTime       uint32
		OperatorUin       uint64
		SenderNick        string
		OperatorNick      string
	}

	// 群戳一戳事件 user为发送者 from miraigo
	GroupPokeEvent struct {
		GroupEvent
		Receiver uint64
		Suffix   string
		Action   string
	}

	// 群消息表态 user为发送表态的成员
	GroupReactionEvent struct {
		GroupEvent
		TargetSeq uint32
		IsAdd     bool
		IsEmoji   bool
		Code      string
		Count     uint32
	}

	// 群成员头衔更新事件 from miraigo
	MemberSpecialTitleUpdated struct {
		GroupEvent
		NewTitle string
	}
)

type GroupInvite struct {
	GroupUin    uint64
	GroupName   string
	InvitorUid  string
	InvitorUin  uint64
	InvitorNick string
	RequestSeq  uint64
}

type JSONParam struct {
	Cmd  int    `json:"cmd"`
	Data string `json:"data"`
	Text string `json:"text"`
	URL  string `json:"url"`
}

// uid => uin
type Iuid2uin interface {
	ResolveUin(func(uid string, groupUin ...uint64) uint64)
}

func (g *GroupMute) MuteAll() bool { return g.OperatorUid == "" }

func (g *GroupMemberDecrease) IsKicked() bool { return g.ExitType == 131 || g.ExitType == 3 }

func (g *GroupDigestEvent) IsSet() bool { return g.OperationType == 1 }

func (g *GroupMemberPermissionChanged) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	g.UserUin = f(g.UserUid, g.GroupUin)
}

func ParseGroupMemberPermissionChanged(event *notify.GroupAdmin) *GroupMemberPermissionChanged {
	var admin bool
	var uid string
	if event.Body.ExtraEnable != nil {
		admin, uid = true, event.Body.ExtraEnable.AdminUid
	} else if event.Body.ExtraDisable != nil {
		admin, uid = false, event.Body.ExtraDisable.AdminUid
	}
	return &GroupMemberPermissionChanged{
		IsAdmin: admin,
		GroupEvent: GroupEvent{
			GroupUin: uint64(event.GroupUin),
			UserUid:  uid,
		}}
}

func (g *GroupNameUpdated) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	g.UserUin = f(g.UserUid, g.GroupUin)
}

func ParseGroupNameUpdatedEvent(event *notify.NotifyMessageBody, groupName string) *GroupNameUpdated {
	return &GroupNameUpdated{
		NewName: groupName,
		GroupEvent: GroupEvent{
			GroupUin: uint64(event.GroupUin.Unwrap()),
			UserUid:  event.OperatorUid.Unwrap(),
		}}
}

func (g *GroupMemberJoinRequest) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	g.InvitorUin = f(g.InvitorUid, g.GroupUin)
}

// 成员主动加群
func ParseRequestJoinNotice(event *notify.GroupJoin) *GroupMemberJoinRequest {
	return &GroupMemberJoinRequest{
		Answer: event.Comment.Unwrap(),
		GroupEvent: GroupEvent{
			GroupUin: uint64(event.GroupUin.Unwrap()),
			UserUid:  event.TargetUid.Unwrap(),
		}}
}

// 成员被邀请加群
func ParseRequestInvitationNotice(event *notify.GroupInvite) *GroupMemberJoinRequest {
	//inn := event.Info.Inner
	inn := event.Body
	return &GroupMemberJoinRequest{
		InvitorUid: inn.InviterUid.Unwrap(),
		GroupEvent: GroupEvent{
			GroupUin: uint64(inn.GroupUin.Unwrap()),
			UserUid:  inn.TargetUid.Unwrap(),
		}}
}

func (g *GroupInvite) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	g.InvitorUin = f(g.InvitorUid, g.GroupUin)
}

// 被邀请加群
func ParseInviteNotice(event *notify.GroupInvite) *GroupInvite {
	return &GroupInvite{GroupUin: uint64(event.Body.GroupUin.Unwrap()), InvitorUid: event.Body.InviterUid.Unwrap()}
}

func (g *GroupMemberIncrease) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	g.InvitorUin = f(g.InvitorUid, g.GroupUin)
	g.UserUin = f(g.UserUid, g.GroupUin)
}

func ParseMemberIncreaseEvent(event *notify.GroupChange) *GroupMemberIncrease {
	return &GroupMemberIncrease{
		InvitorUid: utils.B2S(event.Operator),
		JoinType:   event.Type.Unwrap(),
		GroupEvent: GroupEvent{
			GroupUin: uint64(event.GroupUin.Unwrap()),
			UserUid:  event.MemberUid.Unwrap(),
		}}
}

func (g *GroupMemberDecrease) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	g.UserUin = f(g.UserUid, g.GroupUin)
	if g.IsKicked() {
		g.OperatorUin = f(g.OperatorUid, g.GroupUin)
	}
}

func ParseMemberDecreaseEvent(event *notify.GroupChange) *GroupMemberDecrease {
	return &GroupMemberDecrease{
		OperatorUid: utils.B2S(event.Operator),
		ExitType:    event.Type.Unwrap(),
		GroupEvent: GroupEvent{
			GroupUin: uint64(event.GroupUin.Unwrap()),
			UserUid:  event.MemberUid.Unwrap(),
		}}
}

func (g *GroupRecall) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	g.OperatorUin = f(g.OperatorUid, g.GroupUin)
	g.UserUin = f(g.UserUid, g.GroupUin)
}

func ParseGroupRecallEvent(event *notify.NotifyMessageBody) *GroupRecall {
	info := event.Recall.RecallMessages[0]
	return &GroupRecall{
		OperatorUid: event.Recall.OperatorUid.Unwrap(),
		Sequence:    info.Sequence.Unwrap(),
		Time:        int64(info.Time.Unwrap()),
		Random:      info.Random.Unwrap(),
		GroupEvent: GroupEvent{
			GroupUin: uint64(event.GroupUin.Unwrap()),
			UserUid:  info.AuthorUid.Unwrap(),
		}}
}

func (g *GroupMute) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	g.OperatorUin = f(g.OperatorUid, g.GroupUin)
	g.UserUin = f(g.UserUid, g.GroupUin)
}

func ParseGroupMuteEvent(event *notify.GroupMute) *GroupMute {
	return &GroupMute{
		OperatorUid: event.OperatorUid.Unwrap(),
		Duration:    event.Data.State.Duration,
		GroupEvent: GroupEvent{
			GroupUin: uint64(event.GroupUin),
			UserUid:  event.Data.State.TargetUid.Unwrap(),
		}}
}

func ParseGroupDigestEvent(event *notify.NotifyMessageBody) *GroupDigestEvent {
	return &GroupDigestEvent{
		MessageId:         uint32(event.EssenceMessage.MsgSequence.Unwrap()),
		InternalMessageId: event.EssenceMessage.Random.Unwrap(),
		OperationType:     event.EssenceMessage.SetFlag.Unwrap(),
		OperateTime:       event.EssenceMessage.TimeStamp.Unwrap(),
		OperatorUin:       uint64(event.EssenceMessage.OperatorUin.Unwrap()),
		SenderNick:        event.EssenceMessage.MemberNickName.Unwrap(),
		OperatorNick:      event.EssenceMessage.OperatorNickName.Unwrap(),
		GroupEvent: GroupEvent{
			GroupUin: uint64(event.EssenceMessage.GroupUin.Unwrap()),
			UserUin:  uint64(event.EssenceMessage.MemberUin.Unwrap()),
		}}
}

func ParseGroupPokeEvent(event *notify.NotifyMessageBody, groupUin uint64) *GroupPokeEvent {
	e := ParsePokeEvent(event.GeneralGrayTip)
	return &GroupPokeEvent{
		Receiver: e.Receiver,
		Suffix:   e.Suffix,
		Action:   e.Action,
		GroupEvent: GroupEvent{
			GroupUin: groupUin,
			UserUin:  e.Sender,
		}}
}

func ParseGroupMemberSpecialTitleUpdatedEvent(event *notify.GroupSpecialTitle, groupUin uint64) *MemberSpecialTitleUpdated {
	re := regexp.MustCompile(`<({.*?})>`)
	matches := re.FindAllStringSubmatch(event.Content, -1)
	if len(matches) != 2 {
		return nil
	}
	var medalData JSONParam
	if e := json.Unmarshal([]byte(matches[1][1]), &medalData); e != nil {
		return nil
	}
	return &MemberSpecialTitleUpdated{
		NewTitle: medalData.Text,
		GroupEvent: GroupEvent{
			GroupUin: groupUin,
			UserUin:  uint64(event.TargetUin),
		}}
}

func (g *GroupReactionEvent) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	g.UserUin = f(g.UserUid, g.GroupUin)
}

func ParseGroupReactionEvent(event *notify.NotifyMessageBody) *GroupReactionEvent {
	code := event.Reaction.Data.Data.Data.Code.Unwrap()
	return &GroupReactionEvent{
		TargetSeq: uint32(event.Reaction.Data.Data.Target.Sequence.Unwrap()),
		IsAdd:     event.Reaction.Data.Data.Data.Type.Unwrap() == 1,
		IsEmoji:   len(code) > 3,
		Code:      code,
		Count:     event.Reaction.Data.Data.Data.CurrentCount.Unwrap(),
		GroupEvent: GroupEvent{
			GroupUin: uint64(event.GroupUin.Unwrap()),
			UserUid:  event.Reaction.Data.Data.Data.OperatorUid.Unwrap(),
		}}
}

func (g *GroupPokeEvent) From() uint64 { return g.GroupUin }

func (g *GroupPokeEvent) Content() string {
	if g.Suffix == "" {
		return fmt.Sprintf("%d%s%d", g.UserUin, g.Action, g.Receiver)
	}
	return fmt.Sprintf("%d%s%d的%s", g.UserUin, g.Action, g.Receiver, g.Suffix)
}
