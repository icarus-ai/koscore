package message_type

import "github.com/kernel-ai/koscore/client/packets/structs/sso_type"

var (
	AttributeSsoGetGroupMsg = sso_type.NewServiceAttributeD2D2("trpc.msg.register_proxy.RegisterProxy.SsoGetGroupMsg")
	AttributeSsoGetRoamMsg  = sso_type.NewServiceAttributeD2D2("trpc.msg.register_proxy.RegisterProxy.SsoGetRoamMsg")
	AttributeSsoGetC2cMsg   = sso_type.NewServiceAttributeD2D2("trpc.msg.register_proxy.RegisterProxy.SsoGetC2cMsg")
)
