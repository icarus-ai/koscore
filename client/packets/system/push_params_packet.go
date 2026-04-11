package system

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/system"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func ParsePushParamsPacket(pkt *sso_type.SsoPacket) (*system.PushParams, error) {
	return proto.Unmarshal[system.PushParams](pkt.Data)
}
