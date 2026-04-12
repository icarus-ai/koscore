package message_type

import "github.com/kernel-ai/koscore/client/packets/structs/sso_type"

var (
	AttributeSsoGroupRecallMsg = sso_type.NewServiceAttributeD2D2("trpc.msg.msg_svc.MsgService.SsoGroupRecallMsg")
	AttributeSsoC2CRecallMsg   = sso_type.NewServiceAttributeD2D2("trpc.msg.msg_svc.MsgService.SsoC2CRecallMsg")
)
