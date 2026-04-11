package system_type

import "github.com/kernel-ai/koscore/client/packets/structs/sso_type"

type InfoSyncRsp struct {
	Message string
}

var (
	// protocol pc
	AttributeSsoInfoSync = sso_type.NewServiceAttributeD2D2("trpc.msg.register_proxy.RegisterProxy.SsoInfoSync")
	// protocol all
	AttributeInfoSyncPush = sso_type.NewServiceAttributeD2D2("trpc.msg.register_proxy.RegisterProxy.InfoSyncPush")
)
