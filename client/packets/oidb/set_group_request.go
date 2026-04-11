package oidb

import (
	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildSetGroupRequestPcket(isFiltered bool, operate entity.GroupRequestOperate, sequence uint64, typ uint32, groupUin uint64, message string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x10C8, utils.Ternary[uint32](isFiltered, 2, 1), &operation.SetGroupNotificationRequest{
		Operate: proto.Some(uint64(operate)),
		Body: &operation.SetGroupNotificationRequestBody{
			Sequence: proto.Some(sequence),
			Type:     proto.Some(uint64(typ)),
			GroupUin: proto.Some(int64(groupUin)),
			Message:  proto.Some(message),
		},
	}, false, false)
}

var ParseSetGroupRequestPaacket = CheckError
