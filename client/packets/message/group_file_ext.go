package message

import (
	"fmt"

	pkt_oidb "github.com/kernel-ai/koscore/client/packets/oidb"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildGroup2FileSendPacket(gin uint64, fileid string, random uint32) (*sso_type.SsoPacket, error) {
	return pkt_oidb.BuildOidbPacket(0x6D9, 4, &oidb.D6D9ReqBody{
		FeedsInfoReq: &oidb.FeedsReqBody{
			GroupCode: proto.Some(gin),
			AppId:     proto.Some[uint32](2),
			FeedsInfoList: []*message.FeedsInfo{{
				BusId:     proto.Some[uint32](102),
				FileId:    proto.Some(fileid),
				MsgRandom: proto.Some[uint32](random),
				FeedFlag:  proto.Some[uint32](1),
			}},
		},
	}, false, false)
}

func ParseGroupFil2eSendPacket(data []byte) error {
	rsp, e := pkt_oidb.ParseOidbPacket[oidb.D6D9RspBody](data)
	if e != nil {
		return e
	}
	if rsp.FeedsInfoRsp.RetCode.Unwrap() == 0 {
		return nil
	}
	return fmt.Errorf("%s (%d)", rsp.FeedsInfoRsp.RetMsg.Unwrap(), rsp.FeedsInfoRsp.RetCode.Unwrap())
}
