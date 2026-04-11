package structs

import (
	"fmt"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/common"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
)

func buildSsoPackerProtocol12(version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session, sso *sso_type.SsoPacket, info *common.SsoSecureInfo) *binary.Builder {
	head := binary.NewBuilder().
		WriteU32(sso.Sequence). // sequence
		WriteU32(version.SubAppId).
		WriteU32(2052). // unknown
		WriteBytes([]byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}).
		WriteLengthBytes(session.Sig.A2, prefix.Int32|prefix.WithPrefix).             // tgt
		WriteLengthString(sso.Command, prefix.Int32|prefix.WithPrefix).               // command
		WriteLengthBytes(binary.Empty, prefix.Int32|prefix.WithPrefix).               // message_cookies
		WriteLengthString(device.GUID.ToLowHexStr(), prefix.Int32|prefix.WithPrefix). // guid
		WriteLengthBytes(binary.Empty, prefix.Int32|prefix.WithPrefix).
		WriteLengthString(version.CurrentVersion, prefix.Int16|prefix.WithPrefix)

	writeSsoReservedField(version, session, head, info)

	return binary.NewBuilder().
		WriteLengthBytes(head.ToBytes(), prefix.Int32|prefix.WithPrefix).
		WriteLengthBytes(sso.Data, prefix.Int32|prefix.WithPrefix) // payload
}

func buildSsoPackerProtocol13(version *auth.AppInfo, session *auth.Session, sso *sso_type.SsoPacket, info *common.SsoSecureInfo) *binary.Builder {
	head := binary.NewBuilder().
		WriteLengthString(sso.Command, prefix.Int32|prefix.WithPrefix). // command
		WriteLengthBytes(binary.Empty, prefix.Int32|prefix.WithPrefix)  // message_cookies
	writeSsoReservedField(version, session, head, info)
	return binary.NewBuilder().
		WriteLengthBytes(head.ToBytes(), prefix.Int32|prefix.WithPrefix).
		WriteLengthBytes(sso.Data, prefix.Int32|prefix.WithPrefix) // payload
}

const __hex = "0123456789abcdef"

func writeSsoReservedField(version *auth.AppInfo, session *auth.Session, writer *binary.Builder, info *common.SsoSecureInfo) {
	trace := make([]byte, 55)
	trace[0], trace[1], trace[2], trace[35], trace[52], trace[53], trace[54] = '0', '1', '-', '-', '-', '0', '1'
	for i := 3; i < 35; i++ {
		trace[i] = __hex[crypto.RandU32()&0x0F]
	}
	for i := 36; i < 52; i++ {
		trace[i] = __hex[crypto.RandU32()&0x0F]
	}
	reserved := &common.SsoReserveFields{
		TraceParent: proto.Some(string(trace)),
		Uid:         proto.Some(session.Info.Uid),
		SecInfo:     info,
	}
	if version.OS == auth.Android {
		reserved.MsgType = proto.Some[uint32](32)
		reserved.NtCoreVersion = proto.Some[uint32](100)
	}
	/*
	   {
	   	data, _ := proto.Marshal(reserved)
	   	comm.LOGD("WriteSsoReservedField: %d %X", len(data), data)
	   }
	*/
	data, _ := proto.Marshal(reserved)
	writer.WriteLenBarrier(binary.NewBuilder().WriteBytes(data), prefix.Int32, true)
}

func parseSsoPacker(pkt *sso_type.SsoPacket) (*sso_type.SsoPacket, error) {
	r := binary.NewReader(pkt.Data)
	head := r.ReadLengthBytes(prefix.Int32 | prefix.WithPrefix)
	headReader := binary.NewReader(head)
	pkt.Sequence = headReader.ReadU32()
	pkt.RetCode = headReader.ReadI32()
	pkt.Extra = headReader.ReadLengthString(prefix.Int32 | prefix.WithPrefix)
	pkt.Command = headReader.ReadLengthString(prefix.Int32 | prefix.WithPrefix)
	headReader.ReadLengthBytes(prefix.Int32 | prefix.WithPrefix) // msgCookie
	dataFlag := headReader.ReadI32()
	headReader.ReadLengthBytes(prefix.Int32 | prefix.WithPrefix) // reserveField

	/*
		comm.LOGD("parseSsoPacker")
		comm.LOGD("  sequence: %v", sequence)
		comm.LOGD("  retCode: %v", retCode)
		comm.LOGD("  extra: %v", extra)
		comm.LOGD("  command: %v", command)
		comm.LOGD("  dataFlag: %v", dataFlag)
	*/
	if pkt.RetCode == 0 {
		body := r.ReadLengthBytes(prefix.Int32 | prefix.WithPrefix)
		switch dataFlag {
		case 0, 4:
			pkt.Data = body // allocation
		case 1:
			pkt.Data = binary.ZlibUncompress(body)
		default:
			return nil, fmt.Errorf("SsoPacker::Parse: argument out of range exception: %d", dataFlag)
		}
	}
	return pkt, nil
}
