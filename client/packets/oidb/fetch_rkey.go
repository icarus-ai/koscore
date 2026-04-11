package oidb

import (
	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

// ***** HAS_OLD_CODE BIGIN *****

func BuildFetchRKeyPacket() (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x9067, 202, &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: proto.Some[uint32](1),
				Command:   proto.Some[uint32](202),
			},
			Scene: &oidb.SceneInfo{
				RequestType:  proto.Some[uint32](2),
				BusinessType: proto.Some[uint32](1),
				SceneType:    proto.Some[uint32](0),
			},
			Client: &oidb.ClientMeta{AgentType: proto.Some[uint32](2)},
		},
		DownloadRKey: &oidb.DownloadRKeyReq{
			Types: []int32{10, 20, 2},
		},
	}, false, false)
}

func ParseFetchRKeyPacket(data []byte) (entity.RKeyMap, error) {
	rsp, e := ParseOidbPacket[oidb.NTV2RichMediaResp](data)
	if e != nil {
		return nil, e
	}
	ret := make(entity.RKeyMap)
	for _, v := range rsp.DownloadRKey.RKeys {
		typ := entity.RKeyType(v.Type.Unwrap())
		ret[typ] = &entity.RKeyInfo{
			RKey:       v.Rkey.Unwrap(),
			RKeyType:   typ,
			CreateTime: uint64(v.RkeyCreateTime.Unwrap()),
			ExpireTime: uint64(v.RkeyCreateTime.Unwrap()) + v.RkeyTtlSec.Unwrap(),
		}
	}
	return ret, nil
}

// ***** HAS_OLD_CODE END *****
