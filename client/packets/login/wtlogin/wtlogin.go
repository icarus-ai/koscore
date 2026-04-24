package wtlogin

import (
	"time"

	"github.com/fumiama/gofastTEA"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
	"github.com/kernel-ai/koscore/utils/crypto/ecdh"
	"github.com/kernel-ai/koscore/utils/exception"
)

func BuildTransEmp31(version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session, unusualSig []byte, qrcode_size uint32) []byte {
	build := binary.NewBuilder().
		WriteU16(0).WriteU32(version.AppId).
		WriteU64(0).              // uin
		WriteBytes(binary.Empty). // TGT
		WriteU8(0).WriteLenBytes(binary.Empty)
	tlvs := NewTlvQrCode(version, device)
	if unusualSig == nil {
		build.WriteTLV(
			tlvs.T16(), tlvs.T1B(qrcode_size), tlvs.T1D(),
			tlvs.T33(), tlvs.T35(), tlvs.T66(),
			tlvs.TD1(),
		)
	} else {
		build.WriteTLV(
			tlvs.T11(unusualSig),
			tlvs.T16(), tlvs.T1B(qrcode_size), tlvs.T1D(),
			tlvs.T33(), tlvs.T35(), tlvs.T66(),
			tlvs.TD1(),
		)
	}

	//comm.FAIL("xxxxxxx g: %d(%04X) %X", len(src), len(src), src)
	return buildCode2dPacket(version, session, 0x31, build.ToBytes(), EM_ECDH_ST, false, false)
}

// / poll qrcode
func BuildTransEmp12(version *auth.AppInfo, session *auth.Session) []byte {
	build := binary.NewBuilder().
		WriteU16(0).
		WriteU32(version.AppId).
		WriteLengthBytes(session.State.QrSig, prefix.Int16|prefix.LengthOnly).
		WriteU64(0).              // uin
		WriteBytes(binary.Empty). // TGT
		WriteU8(0).
		WriteLengthBytes(binary.Empty, prefix.Int16|prefix.LengthOnly).
		WriteU16(0) // tlv count = 0
	return buildCode2dPacket(version, session, 0x12, build.ToBytes(), EM_ECDH_ST, false, false)
}

func BuildOicq09(version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session) []byte {
	tlvs := NewTlv(version, device, session)
	build := binary.NewBuilder().WriteU16(0x09).WriteTLV(
		tlvs.T106EncryptedA1(),
		tlvs.T144(), tlvs.T116(),
		tlvs.T142(), tlvs.T145(),
		tlvs.T018(), tlvs.T141(),
		tlvs.T177(), tlvs.T191(0),
		tlvs.T100(), tlvs.T107(),
		tlvs.T318(), tlvs.T16A(),
		tlvs.T166(), tlvs.T521(),
	)
	return buildPacket(version, session, 0x810, build.ToBytes(), EM_ECDH_ST, false)
}

func buildCode2dPacket(version *auth.AppInfo, session *auth.Session, command uint16, tlv []byte, method EncryptMethod, encrypt, useWtSession bool) []byte {
	// BuildTransEmp12 c: 21 00015F5E164F00000000000000000001      0200000001
	// BuildTransEmp12 g: 23 00015F5E164F00000000000000000001 0000 0200000001

	// BuildTransEmp31 g: 00005F5E164F00000000000000000000000008001100030B2D0E00160043000000005F5E164F20073F631A9B5BF7FFC04B72D36A363D240B4C21000E636F6D2E74656E63656E742E71710005322E302E30000E636F6D2E74656E63656E742E7171001B001E000000000000000000000003000000040000004800000002000000020000001D000A0100007FFC0000000000003300101A9B5BF7FFC04B72D36A363D240B4C210035000400000013006600040000001300D100200A1A0A054C696E757812114C616772616E67652D353245463038383822023001
	// BuildTransEmp31 c:
	//comm.LOGD("BuildCode2dPacket: 0: cmd: %02X: data: %d %04X %X", command, len(tlv), len(tlv), tlv)

	build := binary.NewBuilder().WriteU32(uint32(time.Now().Unix())).
		WriteU8(2). // encryptMethod == EncryptMethod.EM_ST || encryptMethod == EncryptMethod.EM_ECDH_ST | Section of length 43 + tlv.Length + 1
		WriteLenBarrier(binary.NewBuilder().
			WriteU16(command).
			WriteBytes(make([]byte, 21)).
			WriteU8(3).                 // flag, 4 for oidb_func, 1 for register, 3 for code_2d, 2 for name_func, 5 for devlock
			WriteU16(0x00).             // close
			WriteU16(0x32).             // Version Code: 50
			WriteU32(0).                // trans_emp sequence
			WriteU64(session.Info.Uin). // dummy uin
			WriteBytes(tlv).
			WriteU8(3), // oicq.wlogin_sdk.code2d.c.get_request
			prefix.Int16, true, 1)
	data := build.ToBytes()

	if encrypt {
		data = tea.NewTeaCipher(session.Sig.StKey).Encrypt(data)
	}

	// data c: 69C1F5D2 02 0105 0031000000000000000000000000000000000000000000030000003200000000000000000C905FB100005F5E164F00000000000000000000000008001100030B2D0E00160043000000005F5E164F20073F631A9B5BF7FFC04B72D36A363D240B4C21000E636F6D2E74656E63656E742E71710005322E302E30000E636F6D2E74656E63656E742E7171001B001E000000000000000000000003000000040000004800000002000000020000001D000A0100007FFC0000000000003300101A9B5BF7FFC04B72D36A363D240B4C210035000400000013006600040000001300D100200A1A0A054C696E757812114C616772616E67652D35324546303838382202300103
	// data g: 69C1FBC1 02 0105 0031000000000000000000000000000000000000000000030000003200000000000000000C905FB100005F5E164F00000000000000000000000008001100030B2D0E00160043000000005F5E164F20073F631A9B5BF7FFC04B72D36A363D240B4C21000E636F6D2E74656E63656E742E71710005322E302E30000E636F6D2E74656E63656E742E7171001B001E000000000000000000000003000000040000004800000002000000020000001D000A0100007FFC0000000000003300101A9B5BF7FFC04B72D36A363D240B4C210035000400000013006600040000001300D100200A1A0A054C696E757812114C616772616E67652D35324546303838382202300103
	//comm.LOGD("BuildCode2dPacket: 1: cmd: %02X: data: %d %04X %X", command, len(data), len(data), data)

	build = binary.NewBuilder().
		WriteU8(utils.Ternary[uint8](encrypt, 1, 0)). // flag for encrypt, if 1, encrypt by StKey
		WriteU16(uint16(len(data))).
		WriteU32(version.AppId).
		WriteU32(0x72) // Role
	if encrypt {
		build.WriteLengthBytes(session.Sig.St, prefix.Int16|prefix.LengthOnly) // uSt
	} else {
		build.WriteLengthBytes(binary.Empty, prefix.Int16|prefix.LengthOnly) // uSt
	}
	build.WriteLengthBytes(binary.Empty, prefix.Int8|prefix.LengthOnly) // rollback
	build.WriteBytes(data)                                              // oicq.wlogin_sdk.request.d0

	return buildPacket(version, session, 0x812, build.ToBytes(), method, useWtSession)
}

func buildPacket(version *auth.AppInfo, session *auth.Session, command uint16, payload []byte, method EncryptMethod, useWtSession bool) []byte {
	var key []byte
	switch method {
	case EM_ECDH, EM_ECDH_ST:
		key = ecdh.S192.SharedKey()
	case EM_ST:
		if useWtSession {
			key = session.Sig.WtSessionTicketKey
		} else {
			key = session.Sig.RandomKey
		}
	default: // NewArgumenttOfRangeException("unknown method: %d", method)
	}
	cipher := tea.NewTeaCipher(key).Encrypt(payload)

	return binary.NewBuilder().
		WriteU8(2). // getRequestEncrptedPackage
		WriteLenBarrier(
			buildEncryptHead(session, binary.NewBuilder().
				WriteU16(8001). // version
				WriteU16(command).
				WriteU16(0). // sequence
				WriteU32(uint32(session.Info.Uin)).
				WriteU8(3).
				WriteU8(uint8(method)).
				WriteU32(0).
				WriteU8(2).
				WriteU16(0).                                // insId
				WriteU16(uint16(version.AppClientVersion)). // insId
				WriteU32(0),                                // retryTime
				useWtSession).
				WriteBytes(cipher).WriteU8(3),
			prefix.Int16, true, 1).
		ToBytes()
}

func buildEncryptHead(session *auth.Session, byt *binary.Builder, useWtSession bool) *binary.Builder {
	if useWtSession {
		byt.WriteLengthBytes(session.Sig.WtSessionTicket, prefix.Int16|prefix.LengthOnly)
	} else {
		byt.WriteU8(1).WriteU8(1).
			WriteBytes(session.Sig.RandomKey).
			WriteU16(0x102). // encrypt type
			WriteLengthBytes(ecdh.S192.PublicKey(), prefix.Int16|prefix.LengthOnly)
	}
	return byt
}

func Parse(session *auth.Session, data []byte) (cmd uint16, rsp []byte, e error) {
	reader := binary.NewReader(data)
	reader.ReadU8()  // header
	reader.ReadU16() // len
	reader.ReadU16() // version?
	cmd = reader.ReadU16()
	reader.ReadU16() // seq
	reader.ReadU32() // uin
	reader.ReadU8()  // flag
	encrypt_type := reader.ReadU8()
	state := reader.ReadU8()
	encrypt_data := reader.ReadBytes(reader.Len() - 1)
	var key []byte
	//comm.LOGD("----- ***** ----- ----- ----- -----")
	//comm.LOGD("wtlogin::Parse: cmd: %02X", cmd)
	//comm.LOGD("wtlogin::Parse: encrypt_type: %02X", encrypt_type)
	switch encrypt_type {
	case 0:
		if state == 180 {
			key = session.Sig.RandomKey
		} else {
			key = ecdh.S192.SharedKey()
		}
	case 3:
		key = session.Sig.WtSessionTicketKey
	case 4:
		raw := tea.NewTeaCipher(ecdh.S192.SharedKey()).Decrypt(encrypt_data)
		byt := binary.NewReader(raw)
		key = byt.ReadLengthBytes(prefix.Int16 | prefix.LengthOnly)
		//comm.LOGD("wtlogin::Parse:4 raw: %02X", raw)
		//comm.LOGD("wtlogin::Parse:4 key: %d %02X %02X", len(key), len(key), key) // 0x19-2=0x17 == 25-2=23)
		key, e = ecdh.S192.Exange(key)
		if e != nil {
			return cmd, nil, exception.NewFormat("key exchange: %v", e)
		}
		encrypt_data = byt.ReadAll()
		//comm.LOGD("wtlogin::Parse:4 key: %d %02X %02X", len(key), len(key), key)
		//comm.LOGD("wtlogin::Parse:4 encrypt_data: %d %02X %02X", len(encrypt_data), len(encrypt_data), encrypt_data)
		//comm.LOGD("----- ----- ----- ----- ***** -----")
	default:
		return cmd, nil, exception.NewFormat("unknown encrypt type: %d", encrypt_type)
	}

	rsp = tea.NewTeaCipher(key).Decrypt(encrypt_data)
	return
}

func ParseCode2dPacket(session *auth.Session, data []byte) (cmd uint16, rsp []byte) {
	//comm.LOGD("----- ***** ----- ----- ----- -----")
	//comm.LOGD("raw: %d %02X %02X", len(data), len(data), data)
	//comm.LOGD("----- ----- ----- ----- ***** -----")
	encrypt := data[1]
	layer := utils.B_U16(data[2:4])
	data = data[5 : layer+5]
	if encrypt != 0 {
		data = tea.NewTeaCipher(session.Sig.StKey).Decrypt(data)
	}
	byt := binary.NewReader(data)
	byt.ReadU8()  // header
	byt.ReadU16() // length
	command := byt.ReadU16()
	byt.SkipBytes(21)
	byt.ReadU8()  // flag
	byt.ReadU16() // retryTime
	byt.ReadU16() // version
	byt.ReadU32() // sequence
	byt.ReadU64() // uin
	/*
		comm.LOGD("header: %02X", header)
		comm.LOGD("length: %d", length)
		comm.LOGD("command: %02X", command)
		comm.LOGD("flag: %02X", flag)
		comm.LOGD("retryTime: %d", retryTime)
		comm.LOGD("version: %02X", version)
		comm.LOGD("sequence: %d", sequence)
		comm.LOGD("uin: %d", uin)
	*/
	return command, byt.ReadAll()
}
