package system

/*
import (
	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/pball/v2/system"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/client/packets/system/system_type"
	"github.com/kernel-ai/koscore/utils/crypto"
	"google.golang.org/protobuf/proto"
)

func BuildInfoSyncPacket(version *auth.AppInfo, device *auth.DeviceInfo) []byte {
	data, _ := proto.Marshal(&system.SsoInfoSyncRequest {
		SyncFlag              : proto.Uint32(735),
		ReqRandom             : proto.Uint32(crypto.RandU32()),
		CurActiveStatus       : proto.Uint32(2),
		GroupLastMsgTime      : proto.Uint64(0),
		C2CSyncInfo           : &system.SsoC2CSyncInfo {
			C2CMsgCookie      : &system.SsoC2CMsgCookie { C2CLastMsgTime: proto.Uint64(0) },
			C2CLastMsgTime    : proto.Uint64(0),
			LastC2CMsgCookie  : &system.SsoC2CMsgCookie { C2CLastMsgTime: proto.Uint64(0) },
		},
		NormalConfig          : &system.NormalConfig { IntCfg: make(map[uint32]int32) },
		RegisterInfo          : &system.RegisterInfo {
			Guid              : proto.String(device.GUID.ToUpHexStr()),
			KickPc            : proto.Uint32(0),
			BuildVer          : proto.String(version.CurrentVersion),
			IsFirstRegisterProxyOnline: proto.Uint32(1),
			LocaleId          : proto.Uint32(2052),
			DeviceInfo        : &system.DeviceInfo {
				DevName       : proto.String(device.DeviceName),
				DevType       : proto.String(version.Kernel),
				OsVer         : proto.String(""),
				Brand         : proto.String(""),
				VendorOsName  : proto.String(version.VendorOS),
			},
			SetMute           : proto.Uint32(0),
			RegisterVendorType: proto.Uint32(6),
			RegType           : proto.Uint32(0),
			BusinessInfo      : &system.OnlineBusinessInfo {
				NotifySwitch       : proto.Uint32(1),
				BindUinNotifySwitch: proto.Uint32(1),
			},
			BatteryStatus     : proto.Uint32(0),
			Field12           : proto.Int32(1),
		},
		Unknown               : map[uint32]uint32 {0: 2},
		AppState              : &system.CurAppState{
			IsDelayRequest    : proto.Uint32(0),
			AppStatus         : proto.Uint32(0),
			SilenceStatus     : proto.Uint32(0),

		},
	})
	return data
}

func ParseInfoSyncPacket(pkt *sso_type.SsoPacket) *system_type.InfoSyncRsp {
	var rsp system.SsoSyncInfoResponse
	e   := proto.Unmarshal(pkt.Data, &rsp)
	ret := &system_type.InfoSyncRsp { }
	if e != nil { ret.Message = e.Error()
	}  else if rsp.RegisterResponse == nil || rsp.RegisterResponse.Msg == nil { ret.Message = "failed"
	}  else     { ret.Message = *rsp.RegisterResponse.Msg }
	return ret
}
*/
