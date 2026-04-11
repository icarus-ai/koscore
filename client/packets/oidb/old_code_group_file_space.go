package oidb

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

func BuildGroupFileSpaceReq(groupUin uint64) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x6D8, 3, &oidb.OidbSvcTrpcTcp0X6D8{
		Space: &oidb.OidbSvcTrpcTcp0X6D8Space{
			GroupUin: uint32(groupUin),
			AppId:    7,
		},
	}, false, true)
}

func ParseGroupFileSpaceResp(data []byte) (*oidb.OidbSvcTrpcTcp0X6D8_1ResponseSpace, error) {
	rsp, err := ParseOidbPacket[oidb.OidbSvcTrpcTcp0X6D8_1Response](data)
	if err != nil {
		return nil, err
	}
	return rsp.Space, nil
}
