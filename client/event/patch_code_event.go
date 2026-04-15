package event

import "fmt"

type (
	// 群内抢红包运气王提示事件
	GroupRedBagLuckyKingNotifyEvent struct {
		GroupUin  uint64
		Sender    uint64
		LuckyKing uint64
	}

	// 群成员荣誉变更提示事件
	MemberHonorChangedNotifyEvent struct {
		GroupUin uint64
		Honor    HonorType
		Uin      uint64
		Nick     string
		Previous uint32 // 上一个
	}

	// 群打卡事件
	GroupSignEvent struct {
		GroupUin uint64
		Uin      uint64
		Nick     string
		Sign     string
		Rank     uint32
		RankIMG  string
	}
)

// 戳一戳事件
type (
	PokeEventBase struct {
		Sender       uint64 // uin1
		Receiver     uint64 // uin2
		Action       string
		Suffix       string
		ActionImgUrl string
		MsgSeq       uint64
		MsgTime      int64
		TipsSeqId    uint64
	}

	// from miraigo
	FriendPokeEvent struct {
		PokeEventBase
		PeerUin uint64 // msg.Message.ResponseHead.FromUin

	}

	// user为发送者 from miraigo
	GroupPokeEvent struct {
		PokeEventBase // Sender => GroupEvent.UserUin
		GroupEvent
	}
)

// 戳一戳撤回事件
type (
	PokeRecallEventBase struct {
		OperatorUid string
		OperatorUin uint64
		TipsSeqId   uint64
	}

	FriendPokeRecallEvent struct {
		PokeRecallEventBase
		PeerUid string
		PeerUin uint64
	}

	GroupPokeRecallEvent struct {
		PokeRecallEventBase
		GroupEvent
	}
)

func (g *FriendPokeEvent) From() uint64 { return g.Sender }
func (g *FriendPokeEvent) Content() string {
	if g.Suffix != "" {
		return fmt.Sprintf("%d%s%d的%s", g.Sender, g.Action, g.Receiver, g.Suffix)
	}
	return fmt.Sprintf("%d%s%d", g.Sender, g.Action, g.Receiver)
}

func (m *FriendPokeRecallEvent) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	m.PeerUin = f(m.PeerUid)
	m.OperatorUin = f(m.OperatorUid)
}

func (g *GroupPokeEvent) From() uint64 { return g.GroupUin }
func (g *GroupPokeEvent) Content() string {
	if g.Suffix == "" {
		return fmt.Sprintf("%d%s%d", g.UserUin, g.Action, g.Receiver)
	}
	return fmt.Sprintf("%d%s%d的%s", g.UserUin, g.Action, g.Receiver, g.Suffix)
}

func (m *GroupPokeRecallEvent) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	m.GroupEvent.UserUin = f(m.GroupEvent.UserUid, m.GroupEvent.GroupUin)
	m.OperatorUin = f(m.OperatorUid, m.GroupEvent.GroupUin)
}

func (e *GroupSignEvent) From() uint64 { return e.GroupUin }
func (e *GroupSignEvent) Content() string {
	return fmt.Sprintf("%s 今天第 %d 个打卡", e.Nick, e.Rank)
}

func (e *GroupRedBagLuckyKingNotifyEvent) From() uint64 { return e.GroupUin }
func (e *GroupRedBagLuckyKingNotifyEvent) Content() string {
	return fmt.Sprintf("%d发的红包被领完, %d是运气王", e.Sender, e.LuckyKing)
}

type HonorType uint8

const (
	Talkative    HonorType = 1 // 龙王
	Performer    HonorType = 2 // 群聊之火
	Legend       HonorType = 3 // 群聊炙焰
	StrongNewbie HonorType = 5 // 冒尖小春笋
	Emotion      HonorType = 6 // 快乐源泉

	BRONZE     HonorType = 7  // 学术新星
	SILVER     HonorType = 8  // 顶尖学霸
	GOLDEN     HonorType = 9  // 至尊学神
	WHIRLWIND  HonorType = 10 // 一笔当先
	RICHER     HonorType = 11 // 壕礼皇冠
	RED_PACKET HonorType = 12 // 善财福禄寿
)

func (e *MemberHonorChangedNotifyEvent) From() uint64 { return e.GroupUin }
func (e *MemberHonorChangedNotifyEvent) Content() string {
	switch e.Honor {
	case Talkative:
		return fmt.Sprintf("昨日 %s(%d) 在群 %d 内发言最积极, 获得 龙王 标识。", e.Nick, e.Uin, e.GroupUin)
	case Performer:
		return fmt.Sprintf("%s(%d) 在群 %d 里连续发消息超过7天, 获得 群聊之火 标识。", e.Nick, e.Uin, e.GroupUin)
	case Legend:
		return fmt.Sprintf("%s(%d) 在群 %d 里连续发消息超过30天, 获得 群聊炙焰 标识。", e.Nick, e.Uin, e.GroupUin)
	case Emotion:
		return fmt.Sprintf("%s(%d) 在群聊 %d 中连续发表情包超过3天，且累计数量超过20条，获得 快乐源泉 标识。", e.Nick, e.Uin, e.GroupUin)
	case RED_PACKET:
		return fmt.Sprintf("昨日 %s(%d) 在群 %d 内发红包个数最多, 获得 善财福禄寿 标识。", e.Nick, e.Uin, e.GroupUin)

	case StrongNewbie:
		return fmt.Sprintf("%s(%d) 在群 %d, 获得 冒尖小春笋 标识。", e.Nick, e.Uin, e.GroupUin)
	case BRONZE:
		return fmt.Sprintf("%s(%d) 在群 %d, 获得 学术新星 标识。", e.Nick, e.Uin, e.GroupUin)
	case SILVER:
		return fmt.Sprintf("%s(%d) 在群 %d, 获得 顶尖学霸 标识。", e.Nick, e.Uin, e.GroupUin)
	case GOLDEN:
		return fmt.Sprintf("%s(%d) 在群 %d, 获得 至尊学神 标识。", e.Nick, e.Uin, e.GroupUin)
	case WHIRLWIND:
		return fmt.Sprintf("%s(%d) 在群 %d, 获得 一笔当先 标识。", e.Nick, e.Uin, e.GroupUin)
	case RICHER:
		return fmt.Sprintf("%s(%d) 在群 %d, 获得 壕礼皇冠 标识。", e.Nick, e.Uin, e.GroupUin)

	default:
		return "ERROR"
	}
}
