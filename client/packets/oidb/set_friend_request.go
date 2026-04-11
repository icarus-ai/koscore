package oidb

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils"
)

func BuildSetFriendRequestPacket(accept bool, target_uid string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0xb5d, 44, &oidb.OidbSvcTrpcTcp0XB5D_44{
		Accept:    utils.Ternary[uint32](accept, 3, 5),
		TargetUid: target_uid,
	}, false, false)
}

var ParseSetFriendRequestPacket = CheckError
