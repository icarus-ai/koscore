package oidb

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildNudgePacket(gin, uin uint64) (*sso_type.SsoPacket, error) {
	body := &oidb.DED3ReqBody{ToUin: proto.Some(int64(uin))}
	if gin > 0 {
		body.GroupCode = proto.Some(int64(gin))
	} else {
		body.AioUin = proto.Some(int64(uin))
	}
	return BuildOidbPacket(0xED3, 1, body, false, false)
}
