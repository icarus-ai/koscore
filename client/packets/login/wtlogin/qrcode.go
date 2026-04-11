package wtlogin

import (
	"fmt"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/login"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/proto"
)

type TlvQrCode struct {
	version *auth.AppInfo
	device  *auth.DeviceInfo
}

func NewTlvQrCode(app *auth.AppInfo, dev *auth.DeviceInfo) *TlvQrCode {
	return &TlvQrCode{version: app, device: dev}
}

// ??
func (m *TlvQrCode) T02() []byte { return binary.NewBuilder().WriteI32(0).WriteI32(0x0B).Pack(0x02) }

// ??
func (m *TlvQrCode) T04(uin uint32) []byte {
	return binary.NewBuilder().WriteI16(0x00). // uin for 0, uid for 1
							WritePacketString(fmt.Sprint(uin), "u16", false).
							Pack(0x04)
}

// ??
func (m *TlvQrCode) T09() []byte {
	return binary.NewBuilder().WriteBytes([]byte(m.version.PackageName)).Pack(0x09)
}

func (m *TlvQrCode) T11(unusualSign []byte) []byte {
	return binary.NewBuilder().WriteBytes(unusualSign).Pack(0x11)
}

func (m *TlvQrCode) T16() []byte {
	return binary.NewBuilder().
		WriteU32(0).
		WriteU32(m.version.AppId).
		WriteU32(m.version.SubAppId).
		WriteBytes(m.device.GUID).
		WritePacketString(m.version.PackageName, "u16", false).
		WritePacketString(m.version.PtVersion, "u16", false).
		WritePacketString(m.version.PackageName, "u16", false).
		Pack(0x16)
}

func (m *TlvQrCode) T1B(size uint32) []byte {
	return binary.NewBuilder().WriteStruct(
		uint32(0),  // micro
		uint32(0),  // version
		size,       // size default 3
		uint32(4),  // margin
		uint32(72), // dpi
		uint32(2),  // ecLevel
		uint32(2),  // hint
		uint16(0)). // unknown
		Pack(0x1B)
}

func (m *TlvQrCode) T1D() []byte {
	return binary.NewBuilder().
		WriteU8(1).
		WriteU32(m.version.SdkInfo.MiscBitMap).
		WriteU32(0).WriteU8(0).
		Pack(0x1D)
}

// tils.MustParseHexStr(m._keystore.GUID)

func (m *TlvQrCode) T33() []byte { return binary.NewBuilder().WriteBytes(m.device.GUID).Pack(0x33) }
func (m *TlvQrCode) T35() []byte {
	return binary.NewBuilder().WriteU32(m.version.SsoVersion).Pack(0x35)
}
func (m *TlvQrCode) T66() []byte {
	return binary.NewBuilder().WriteU32(m.version.SsoVersion).Pack(0x66)
}

func (m *TlvQrCode) TD1() []byte {
	d, _ := proto.Marshal(&login.QrExtInfo{ //tlv_0Xd1
		DevInfo: &login.DevInfo{
			DevType: proto.Some(m.version.OS.String()),
			DevName: proto.Some(m.device.DeviceName),
		},
		GenInfo: &login.GenInfo{Field6: proto.Some[uint32](1)},
	})
	return binary.NewBuilder().WriteBytes(d).Pack(0xd1)
	/*
		return binary.NewBuilder().WriteBytes(proto.DynamicMessage{
			1: proto.DynamicMessage{
				1: m._appInfo.OS,
				2: m._keystore.DeviceName,
			},
			4: proto.DynamicMessage{ 6: 1 },
		}.Encode()).Pack(0xd1)
	*/
}
