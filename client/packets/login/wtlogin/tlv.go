package wtlogin

import (
	"fmt"
	"time"

	"github.com/fumiama/gofastTEA"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
	"github.com/kernel-ai/koscore/utils/crypto"
)

type Tlv struct {
	version *auth.AppInfo
	device  *auth.DeviceInfo
	session *auth.Session
}

func NewTlv(version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session) *Tlv {
	return &Tlv{
		version: version,
		device:  device,
		session: session,
	}
}

func (m *Tlv) T001() []byte {
	return binary.NewBuilder().
		WriteU16(0x0001).
		WriteU32(crypto.RandU32()).
		WriteU32(uint32(m.session.Info.Uin)).
		WriteU32(uint32(time.Now().Unix())).
		WriteU32(0). // dummy IP Address
		WriteU16(0).Pack(0x01)
}

func (m *Tlv) T008() []byte {
	return binary.NewBuilder().
		WriteU16(0).
		WriteU32(2052). // locale_id
		WriteU32(0).Pack(0x08)
}

func (m *Tlv) T018() []byte {
	return binary.NewBuilder().
		WriteU16(0). // pingVersion
		WriteU32(5). // ssoVersion = 5
		WriteU32(0).
		WriteU32(8001). // app client ver
		WriteU32(uint32(m.session.Info.Uin)).
		WriteU16(0). // unknown = 0
		WriteU16(0).Pack(0x18)
}

func (m *Tlv) T018Android() []byte {
	return binary.NewBuilder().
		WriteU16(0x0001).
		WriteU32(0x00000600).
		WriteI32(int32(m.version.AppId)).
		WriteI32(int32(m.version.AppClientVersion)).
		WriteU32(uint32(m.session.Info.Uin)).
		WriteU16(0x0000). // unknown = 0
		WriteU16(0x0000).Pack(0x18)
}

func (m *Tlv) T100() []byte {
	return binary.NewBuilder().
		WriteU16(0). // db buf ver
		WriteU32(5). // sso ver, dont over 7
		WriteU32(m.version.AppId).
		WriteU32(m.version.SubAppId).
		WriteU32(m.version.AppClientVersion). // app client ver
		WriteU32(uint32(m.version.SdkInfo.MainSigMap)).Pack(0x100)
}

func (m *Tlv) T100Android(mainSigMap uint32) []byte {
	return binary.NewBuilder().
		WriteU16(1).                    // db buf ver
		WriteU32(m.version.SsoVersion). // sso ver, dont over 7
		WriteU32(m.version.AppId).
		WriteU32(m.version.SubAppId).
		WriteU32(m.version.AppClientVersion). // app client ver
		WriteU32(mainSigMap).Pack(0x100)
}

func (m *Tlv) T104(verificationToken []byte) []byte {
	return binary.NewBuilder().WriteBytes(verificationToken).Pack(0x104)
}

// Tlv106Pwd TLV_0x106 (密码登录)
func (m *Tlv) T106Pwd(password string) []byte {
	// 构建密钥
	pwd := crypto.MD5Digest([]byte(password))
	key := crypto.MD5Digest(binary.NewBuilder().
		WriteBytes(pwd).
		WriteU32(0). // empty 4 bytes
		WriteU32(uint32(m.session.Info.Uin)).
		ToBytes())

	return binary.NewBuilder().WriteBytes(
		tea.NewTeaCipher(key).Encrypt(
			binary.NewBuilder().
				WriteU16(4). // TGTGT Version
				WriteU32(crypto.RandU32()).
				WriteU32(m.version.SsoVersion).
				WriteU32(m.version.AppId).
				WriteU32(m.version.AppClientVersion).
				WriteU64(m.session.Info.Uin).
				WriteU32(uint32(time.Now().Unix())).
				WriteU32(0). // dummy IP Address
				WriteU8(1).
				WriteBytes(pwd).
				WriteBytes(m.session.Sig.TgtgtKey).
				WriteU32(0). // unknown
				WriteU8(1).  // guidAvailable
				WriteBytes(m.device.GUID).
				WriteU32(m.version.SubAppId).
				WriteU32(1). // flag
				WriteLengthString(fmt.Sprint(m.session.Info.Uin), prefix.Int16|prefix.LengthOnly).
				WriteU16(0).
				ToBytes())).
		Pack(0x106)
}

func (m *Tlv) T106EncryptedA1() []byte {
	return binary.NewBuilder().WriteBytes(m.session.Sig.A1).Pack(0x106)
}

func (m *Tlv) T107() []byte {
	return binary.NewBuilder().
		WriteU16(1).   // pic type
		WriteU8(0x0D). // captcha type
		WriteU16(0).   // pic size
		WriteU8(1).    // ret type
		Pack(0x107)
}

func (m *Tlv) T107Android() []byte {
	return binary.NewBuilder().
		WriteU16(0). // pic type
		WriteU8(0).  // captcha type
		WriteU16(0). // pic size
		WriteU8(1).  // ret type
		Pack(0x107)
}

//func (m *Tlv) T109() []byte { return binary.NewBuilder().WriteBytes(crypto.MD5Digest([]byte(m.device.AndroidId))).Pack(0x109) }

func (m *Tlv) T116() []byte {
	return binary.NewBuilder().
		WriteU8(0). // version
		WriteU32(m.version.SdkInfo.MiscBitMap).
		WriteU32(m.version.SdkInfo.SubSigMap).
		WriteU8(0). // length of SubAppId
		Pack(0x116)
}

func (m *Tlv) T112(qid string) []byte { return binary.NewBuilder().WriteBytes([]byte(qid)).Pack(0x112) }

func (m *Tlv) T11B() []byte { return binary.NewBuilder().WriteU8(2).Pack(0x11B) }

func (m *Tlv) T124() []byte { return binary.NewBuilder().WriteBytes(make([]byte, 12)).Pack(0x124) }

func (m *Tlv) T128() []byte {
	return binary.NewBuilder().
		WriteU16(0).
		WriteU8(0).  // guid new
		WriteU8(0).  // guid available
		WriteU8(0).  // guid changed
		WriteU32(0). // guid flag
		WriteLengthString(m.version.OS.String(), prefix.Int16|prefix.LengthOnly).
		WriteLengthBytes(m.device.GUID, prefix.Int16|prefix.LengthOnly).
		WriteLengthBytes(binary.Empty, prefix.Int16|prefix.LengthOnly). // brand
		Pack(0x128)
}

func (m *Tlv) T141() []byte {
	return binary.NewBuilder().
		WriteU16(0).
		WriteLengthString("Unknown", prefix.Int16|prefix.LengthOnly).
		WriteU32(0).Pack(0x141)
}

func (m *Tlv) T142() []byte {
	return binary.NewBuilder().WriteU16(0).WriteLengthString(m.version.PackageName, prefix.Int16|prefix.LengthOnly).Pack(0x142)
}

func (m *Tlv) T144() []byte {
	return binary.NewBuilder().WriteBytes(
		tea.NewTeaCipher(m.session.Sig.TgtgtKey).
			Encrypt(binary.NewBuilder().WriteTLV(
				m.T16E(), m.T147(),
				m.T128(), m.T124(),
			).ToBytes())).
		Pack(0x144)
}

func (m *Tlv) T145() []byte { return binary.NewBuilder().WriteBytes(m.device.GUID).Pack(0x145) }

func (m *Tlv) T147() []byte {
	return binary.NewBuilder().
		WriteU32(m.version.AppId).
		WriteLengthString(m.version.PtVersion, prefix.Int16|prefix.LengthOnly).
		WriteLengthBytes(m.version.ApkSignatureMd5, prefix.Int16|prefix.LengthOnly).
		Pack(0x147)
}

func (m *Tlv) T166() []byte { return binary.NewBuilder().WriteI8(5).Pack(0x166) }

func (m *Tlv) T16A() []byte {
	return binary.NewBuilder().WriteBytes(m.session.Sig.NoPicSig).Pack(0x16A)
}

func (m *Tlv) T16E() []byte {
	return binary.NewBuilder().WriteBytes([]byte(m.device.DeviceName)).Pack(0x16E)
}

func (m *Tlv) T177() []byte {
	return binary.NewBuilder().
		WriteU8(1).
		WriteU32(0). // sdk build time
		WriteLengthString(m.version.SdkInfo.SdkVersion, prefix.Int16|prefix.LengthOnly).
		Pack(0x177)
}

func (m *Tlv) T191(canWebVerify uint8) []byte {
	return binary.NewBuilder().WriteU8(canWebVerify).Pack(0x191)
}

func (m *Tlv) T318() []byte { return binary.NewBuilder().Pack(0x318) }

func (m *Tlv) T521() []byte {
	return binary.NewBuilder().
		WriteU32(0x13).                                               // productType
		WriteLengthString("basicim", prefix.Int16|prefix.LengthOnly). // productDesc
		Pack(0x521)
}

/*


	tlvs.T521(),
*/
