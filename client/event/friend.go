package event

import "github.com/kernel-ai/koscore/client/packets/pb/v2/notify"

type (
	NewFriendRequest struct {
		SourceUin  uint64
		SourceUid  string
		SourceNick string
		Msg        string
		Source     string
	}

	NewFriend struct {
		FromUin  uint64
		FromUid  string
		FromNick string
		Msg      string
	}

	FriendRecall struct {
		FromUin  uint64
		FromUid  string
		Sequence uint64
		Time     int64
		Random   uint32
	}

	Rename struct {
		SubType  uint32 // self 0 friend 1
		Uin      uint64
		Uid      string
		Nickname string
	}
)

func ParseFriendRequestNotice(info *notify.FriendRequestInfo) *NewFriendRequest {
	return &NewFriendRequest{
		SourceUid: info.SourceUid.Unwrap(),
		Msg:       info.Message.Unwrap(),
		Source:    info.Source.Unwrap(),
	}
}

func (fe *FriendRecall) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	fe.FromUin = f(fe.FromUid)
}

func ParseNewFriendEvent(event *notify.NewFriend) *NewFriend {
	info := event.Info
	return &NewFriend{
		FromUid:  info.Uid,
		FromNick: info.NickName,
		Msg:      info.Message,
	}
}

func (fe *NewFriend) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	fe.FromUin = f(fe.FromUid)
}

func ParseFriendRecallEvent(info *notify.FriendRecallInfo) *FriendRecall {
	return &FriendRecall{
		FromUid:  info.FromUid.Unwrap(),
		Sequence: uint64(info.Sequence.Unwrap()),
		//Time:     info.Time,
		//Random:   info.Random,
	}
}

func (fe *Rename) ResolveUin(f func(uid string, groupUin ...uint64) uint64) {
	fe.Uin = f(fe.Uid)
}

func ParseFriendRenameEvent(event *notify.FriendRenameMsg) *Rename {
	return &Rename{
		SubType:  1,
		Uid:      event.Body.Data.Uid,
		Nickname: event.Body.Data.RenameData.NickName,
	}
}
