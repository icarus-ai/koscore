package system

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/system"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/client/packets/system/system_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

// 08011202080018002000
func BuildSsoHeartBeatPacket() *sso_type.SsoPacket {
	data, _ := proto.Marshal(&system.SsoHeartBeatRequest{
		Type:         proto.Some[uint32](1),
		LocalSilence: &system.SilenceState{LocalSilence: proto.Some[uint32](0)},
		BatteryState: proto.Some[uint32](0),
		Time:         proto.Some[uint64](0),
	})
	return system_type.AttributeSsoHeartBeat.NewSsoPacket(0, data)
}

func ParseSsoHeartBeatPacket(pkt *sso_type.SsoPacket) (*system.SsoHeartBeatResponse, error) {
	return proto.Unmarshal[system.SsoHeartBeatResponse](pkt.Data)
}
