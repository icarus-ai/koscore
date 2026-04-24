package message

import (
	"errors"
	"strings"

	"github.com/kernel-ai/koscore/client/packets/message/message_type"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildMarketFaceKeyPacket(face_ids ...string) *sso_type.SsoPacket {
	for i, v := range face_ids {
		face_ids[i] = strings.ToLower(v)
	}
	data, _ := proto.Marshal(&operation.MarketFaceKeyReq{
		Field1: 3,
		Info:   &operation.MarketFaceKeyReqInfo{FaceIds: face_ids},
	})
	return message_type.AttributeTabOpReq.NewSsoPacket(0, data)
}

func ParseMarketFaceKeyPacket(data []byte) ([]string, error) {
	info, e := proto.Unmarshal[operation.MarketFaceKeyRsp](data)
	if e != nil {
		return nil, e
	}
	if info.Info == nil {
		return nil, errors.New("valid ids")
	}
	return info.Info.Keys, nil
}
