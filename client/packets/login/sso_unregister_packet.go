package login

import (
	"github.com/kernel-ai/koscore/client/packets/login/login_type"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/system"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildSsoUnregisterPacket() *sso_type.SsoPacket {
	data, _ := proto.Marshal(&system.SsoUnregister{
		RegType:     proto.Some[int32](1),
		DeviceInfo:  &system.DeviceInfo{},
		UserTrigger: proto.Some[int32](1),
	})
	return login_type.AttributeUnRegister.NewSsoPacket(0, data)
}

func ParseSsoUnregisterPacket(data []byte) (*system.RegisterResponse, error) {
	return proto.Unmarshal[system.RegisterResponse](data)
}
