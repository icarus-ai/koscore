package oidb

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildAddGroupReactionPacket(groupUin uint64, sequence uint64, code string, isAdd bool) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x9082, utils.Ternary[uint32](isAdd, 1, 2), &operation.SetGroupReactionRequest{
		GroupUin: proto.Some(int64(groupUin)),
		Sequence: proto.Some(sequence),
		Code:     proto.Some(code),
		Type:     proto.Some(utils.Ternary[uint64](len(code) > 3, 2, 1)),
	}, false, false)
}

var ParseAddGroupReactionPacket = CheckError
