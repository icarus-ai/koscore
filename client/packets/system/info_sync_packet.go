package system

import (
	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/system"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/client/packets/system/system_type"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildInfoSyncPacket(version *auth.AppInfo, device *auth.DeviceInfo) *sso_type.SsoPacket {
	data, _ := proto.Marshal(&system.SsoInfoSyncRequest{
		SyncFlag:         proto.Some[uint32](735),
		ReqRandom:        proto.Some(crypto.RandU32()),
		CurActiveStatus:  proto.Some[uint32](2),
		GroupLastMsgTime: proto.Some[uint64](0),
		C2CSyncInfo: &system.SsoC2CSyncInfo{
			C2CMsgCookie:     &system.SsoC2CMsgCookie{C2CLastMsgTime: proto.Some[uint64](0)},
			C2CLastMsgTime:   proto.Some[uint64](0),
			LastC2CMsgCookie: &system.SsoC2CMsgCookie{C2CLastMsgTime: proto.Some[uint64](0)},
		},
		NormalConfig: &system.NormalConfig{IntCfg: make(map[uint32]int32)},
		RegisterInfo: &system.RegisterInfo{
			Guid:                       proto.Some(device.GUID.ToUpHexStr()),
			KickPc:                     proto.Some[uint32](0),
			BuildVer:                   proto.Some(version.CurrentVersion),
			IsFirstRegisterProxyOnline: proto.Some[uint32](1),
			LocaleId:                   proto.Some[uint32](2052),
			DeviceInfo: &system.DeviceInfo{
				DevName:      proto.Some(device.DeviceName),
				DevType:      proto.Some(version.Kernel),
				OsVer:        proto.Some(""),
				Brand:        proto.Some(""),
				VendorOsName: proto.Some(version.VendorOS),
			},
			SetMute:            proto.Some[uint32](0),
			RegisterVendorType: proto.Some[uint32](6),
			RegType:            proto.Some[uint32](0),
			BusinessInfo: &system.OnlineBusinessInfo{
				NotifySwitch:        proto.Some[uint32](1),
				BindUinNotifySwitch: proto.Some[uint32](1),
			},
			BatteryStatus: proto.Some[uint32](0),
			Field12:       proto.Some[int32](1),
		},
		Unknown: map[uint32]uint32{0: 2},
		AppState: &system.CurAppState{
			IsDelayRequest: proto.Some[uint32](0),
			AppStatus:      proto.Some[uint32](0),
			SilenceStatus:  proto.Some[uint32](0),
		},
	})
	return system_type.AttributeSsoInfoSync.NewSsoPacket(0, data)
}

func ParseInfoSyncPacket(pkt *sso_type.SsoPacket) *system_type.InfoSyncRsp {
	rsp, e := proto.Unmarshal[system.SsoSyncInfoResponse](pkt.Data)
	ret := &system_type.InfoSyncRsp{}
	if e != nil {
		ret.Message = e.Error()
	} else if rsp.RegisterResponse == nil || rsp.RegisterResponse.Msg.IsNone() {
		ret.Message = "failed"
	} else {
		ret.Message = rsp.RegisterResponse.Msg.Unwrap()
	}
	return ret
}
