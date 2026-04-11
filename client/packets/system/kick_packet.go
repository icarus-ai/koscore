package system

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/common"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func ParseKickPacket(pkt *sso_type.SsoPacket) (*common.KickNTReq, error) {
	return proto.Unmarshal[common.KickNTReq](pkt.Data)
}
