package event

type (
	FriendPokeEvent2 struct {
		Sender       uint64 // uin1
		Receiver     uint64 // uin2
		Action       string
		Suffix       string
		ActionImgUrl string
		PeerUin      uint64 // msg.Message.ResponseHead.FromUin
		MsgSeq       uint32
		MsgTime      uint32
		TipsSeqId    uint64
	}

	GroupPokeEvent2 struct {
		GroupEvent GroupEvent
		//Sender       uint32 // uin1  GroupEvent.UserUin
		Receiver     uint64 // uin2
		Action       string
		Suffix       string
		ActionImgUrl string
		MsgSeq       uint32
		MsgTime      uint32
		TipsSeqId    uint64
	}

	FriendPokeRecallEvent struct {
		PeerUid     string
		PeerUin     uint64
		OperatorUid string
		OperatorUin uint64
		TipsSeqId   uint64
	}

	GroupPokeRecallEvent struct {
		GroupEvent  GroupEvent
		OperatorUid string
		OperatorUin uint64
		TipsSeqId   uint64
	}
)

func (m *FriendPokeRecallEvent) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	m.PeerUin = f(m.PeerUid)
	m.OperatorUin = f(m.OperatorUid)
}

func (m *GroupPokeRecallEvent) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	m.GroupEvent.UserUin = f(m.GroupEvent.UserUid, m.GroupEvent.GroupUin)
	m.OperatorUin = f(m.OperatorUid, m.GroupEvent.GroupUin)
}
