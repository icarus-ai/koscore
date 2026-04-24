package oidb

import (
	"fmt"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/exception"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildOidbPacket(cmd, sub uint32, body any, islafter, isuid bool) (*sso_type.SsoPacket, error) {
	data, e := proto.Marshal(body)
	if e != nil {
		return nil, e
	}
	if data, e = proto.Marshal(&oidb.Oidb{
		Command:  proto.Some(cmd),
		Service:  proto.Some(sub),
		Body:     data,
		Reserved: proto.Some(uint32(utils.Bool2Int(isuid))),
	}); e != nil {
		return nil, e
	}
	return sso_type.NewServiceAttributeD2D2(fmt.Sprintf("OidbSvcTrpcTcp.0x%02x_%d", cmd, sub)).NewSsoPacket(0, data), nil
}

func ParseOidbPacket[T any](body []byte, nobody ...bool) (*T, error) {
	base, e := proto.Unmarshal[oidb.Oidb](body)
	if e != nil {
		return nil, e
	}
	if base.Result.Unwrap() != 0 {
		return nil, exception.NewOperationExceptionCode(base.Result.Unwrap(), base.Message.Unwrap())
	}
	if len(nobody) > 0 && nobody[0] {
		return nil, nil
	}
	return proto.Unmarshal[T](base.Body)
}

func CheckError(data []byte) error {
	_, e := ParseOidbPacket[uint8](data, true)
	return e
}

func CheckTypedError[T any](data []byte) error {
	_, e := ParseOidbPacket[T](data)
	return e
}

func ParseTypedError[T any](data []byte) (*T, error) { return ParseOidbPacket[T](data) }
