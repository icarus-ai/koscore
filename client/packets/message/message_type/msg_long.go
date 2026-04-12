package message_type

import "github.com/kernel-ai/koscore/client/packets/structs/sso_type"

var (
	AttributeSsoSendLongMsg = sso_type.NewServiceAttributeD2D2("trpc.group.long_msg_interface.MsgService.SsoSendLongMsg")
	AttributeSsoRecvLongMsg = sso_type.NewServiceAttributeD2D2("trpc.group.long_msg_interface.MsgService.SsoRecvLongMsg")
)
