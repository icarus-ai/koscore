package structs

import (
	"fmt"

	"github.com/fumiama/gofastTEA"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
)

var __EmptyD2Key = make([]byte, 16)

func buildServicePackerProtocol12(session *auth.Session, sso *binary.Builder, options *sso_type.ServiceAttribute) ([]byte, error) {
	cipher := sso.ToBytes()
	switch options.EncryptType {
	case sso_type.NoEncrypt:
	case sso_type.EncryptEmpty:
		cipher = tea.NewTeaCipher(__EmptyD2Key).Encrypt(cipher)
	case sso_type.EncryptD2Key:
		cipher = tea.NewTeaCipher(session.Sig.D2Key).Encrypt(cipher)
	default:
		return nil, fmt.Errorf("ServicePacker::BuildProtocol12: argument out of range exception: unknown encrypt type: %d", options.EncryptType)
	}

	w := binary.NewBuilder().WriteU32(12).WriteU8(uint8(options.EncryptType))
	if options.EncryptType == sso_type.EncryptD2Key {
		w.WriteLengthBytes(session.Sig.D2, prefix.Int32|prefix.WithPrefix)
	} else {
		w.WriteU32(4)
	}
	w.WriteU8(0).WriteLengthString(fmt.Sprint(session.Info.Uin), prefix.Int32|prefix.WithPrefix).WriteBytes(cipher)

	return binary.NewBuilder().WriteLenBarrier(w, prefix.Int32, true).ToBytes(), nil
}

func buildServicePackerProtocol13(session *auth.Session, sso *sso_type.SsoPacket, payload *binary.Builder) ([]byte, error) {
	cipher := payload.ToBytes()
	switch sso.EncryptType {
	case sso_type.NoEncrypt:
	case sso_type.EncryptEmpty:
		tea.NewTeaCipher(__EmptyD2Key).Encrypt(cipher)
	case sso_type.EncryptD2Key:
		tea.NewTeaCipher(session.Sig.D2Key).Encrypt(cipher)
	default:
		return nil, fmt.Errorf("ServicePacker::BuildProtocol13: argument out of range exception: unknown encrypt type: %d", sso.EncryptType)
	}
	return binary.NewBuilder().WriteLenBarrier(
		binary.NewBuilder().
			WriteU32(13).
			WriteU8(uint8(sso.EncryptType)).
			WriteU32(sso.Sequence).
			WriteU8(0).
			WriteLengthString(fmt.Sprint(session.Info.Uin), prefix.Int32|prefix.WithPrefix).
			WriteBytes(cipher),
		prefix.Int32, true).
		ToBytes(), nil
}

func parseServicePacker(session *auth.Session, data []byte) *sso_type.SsoPacket {
	r := binary.NewReader(data)
	r.ReadU32() // length
	r.ReadI32() // RequestType protocol
	//if protocol != RequestTypeLogin && protocol != RequestTypeSimple && protocol != RequestTypeNT { return resp, ErrInvalidPacketType }
	authFlag := sso_type.EncryptType(r.ReadU8())         // resp.EncryptType authFlag
	r.ReadU8()                                           // dummy
	r.ReadLengthString(prefix.Int32 | prefix.WithPrefix) // uin
	body := r.ReadAll()

	switch authFlag {
	case sso_type.NoEncrypt: // nothing to do
	case sso_type.EncryptD2Key:
		body = tea.NewTeaCipher(session.Sig.D2Key).Decrypt(body)
	case sso_type.EncryptEmpty:
		body = tea.NewTeaCipher(__EmptyD2Key).Decrypt(body)
	}
	return &sso_type.SsoPacket{
		ServiceAttribute: &sso_type.ServiceAttribute{
			EncryptType: authFlag,
		},
		Data: body,
	}
}
