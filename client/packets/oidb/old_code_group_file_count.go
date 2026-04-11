package oidb

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

func BuildGroupFileCountReq(groupUin uint64) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x6D8, 2, &oidb.OidbSvcTrpcTcp0X6D8{
		Count: &oidb.OidbSvcTrpcTcp0X6D8Count{
			GroupUin: uint32(groupUin),
			AppId:    7,
			BusId:    6,
		},
	}, false, true)
}

func ParseGroupFileCountResp(data []byte) (*oidb.OidbSvcTrpcTcp0X6D8_1ResponseCount, error) {
	rsp, err := ParseOidbPacket[oidb.OidbSvcTrpcTcp0X6D8_1Response](data)
	if err != nil {
		return nil, err
	}
	return rsp.Count, nil
}
