package highway

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildHighWayURLReq(sig_a2 []byte) ([]byte, error) {
	return proto.Marshal(&oidb.C501ReqBody{
		ReqBody: &oidb.SubCmd0X501ReqBody{
			Uin:            proto.Some[uint64](0),
			IdcId:          proto.Some[uint32](0),
			Appid:          proto.Some[uint32](16),
			LoginSigType:   proto.Some[uint32](0),
			LoginSigTicket: sig_a2,
			ServiceTypes:   []uint32{1, 5, 10, 21},
			RequestFlag:    proto.Some[uint32](3),
			Field9:         proto.Some[int32](2),
			Field10:        proto.Some[int32](9),
			Field11:        proto.Some[int32](8),
			Version:        proto.Some("1.0.1"),
		},
	})
}

func ParseHighWayURLReq(data []byte) (*oidb.C501RspBody, error) {
	return proto.Unmarshal[oidb.C501RspBody](data)
}
