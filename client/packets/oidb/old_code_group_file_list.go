package oidb

import (
	"errors"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

func BuildGroupFileListReq(groupUin uint64, targetDirectory string, startIndex uint32, fileCount uint32) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x6D8, 1, &oidb.OidbSvcTrpcTcp0X6D8{
		List: &oidb.OidbSvcTrpcTcp0X6D8List{
			GroupUin:        uint32(groupUin),
			AppId:           7,
			TargetDirectory: targetDirectory,
			FileCount:       fileCount,
			SortBy:          1,
			StartIndex:      startIndex,
			Field17:         2,
			Field18:         0,
		},
	}, false, true)
}

func ParseGroupFileListResp(data []byte) (*oidb.OidbSvcTrpcTcp0X6D8_1Response, error) {
	rsp, err := ParseOidbPacket[oidb.OidbSvcTrpcTcp0X6D8_1Response](data)
	if err != nil {
		return nil, err
	}
	if rsp.List.RetCode != 0 {
		return nil, errors.New(rsp.List.ClientWording)
	}
	return rsp, nil
}
