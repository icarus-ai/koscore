package event

import (
	"fmt"
	"strconv"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/notify"
)

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

	// FriendPokeEvent 好友戳一戳事件 from miraigo
	FriendPokeEvent struct {
		Sender   uint64
		Receiver uint64
		Suffix   string
		Action   string
	}
)

func ParseFriendRequestNotice(event *notify.FriendRequest) *NewFriendRequest {
	info := event.Info
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

func ParseFriendRecallEvent(event *notify.FriendRecall) *FriendRecall {
	info := event.Info
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

func ParsePokeEvent(event *notify.GeneralGrayTipInfo) *FriendPokeEvent {
	e := FriendPokeEvent{}
	e.Action = "戳了戳"
	for _, data := range event.MsgTemplParam {
		switch data.Name.Unwrap() {
		case "uin_str1":
			sender, _ := strconv.Atoi(data.Value.Unwrap())
			e.Sender = uint64(sender)
		case "uin_str2":
			receiver, _ := strconv.Atoi(data.Value.Unwrap())
			e.Receiver = uint64(receiver)
		case "suffix_str":
			e.Suffix = data.Value.Unwrap()
		case "alt_str1":
			e.Action = data.Value.Unwrap()
		}
	}
	return &e
}

func (g *FriendPokeEvent) From() uint64 { return g.Sender }

func (g *FriendPokeEvent) Content() string {
	if g.Suffix != "" {
		return fmt.Sprintf("%d%s%d的%s", g.Sender, g.Action, g.Receiver, g.Suffix)
	}
	return fmt.Sprintf("%d%s%d", g.Sender, g.Action, g.Receiver)
}
