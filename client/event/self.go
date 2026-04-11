package event

import (
	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/notify"
)

func ParseSelfRenameEvent(event *notify.SelfRenameMsg, sig *auth.BotInfo) *Rename {
	sig.Name = event.Body.RenameData.NickName
	return &Rename{
		Uin:      uint64(event.Body.Uin),
		Nickname: event.Body.RenameData.NickName,
	}
}
