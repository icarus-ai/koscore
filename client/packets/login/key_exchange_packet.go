package login

import (
	"fmt"
	"time"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/login/login_type"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/login"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/crypto/ecdh"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildKeyExchangePacket(device *auth.DeviceInfo, session *auth.Session) (*sso_type.SsoPacket, error) {
	publicKey := ecdh.P256.PublicKey()
	shareKey, e := ecdh.P256.Exange(login_type.SERVER_PUB_KYE)
	if e != nil {
		return nil, e
	}

	secret, _ := proto.Marshal(&login.KeyExchangeRequestBuf{
		Uin:  proto.Some(fmt.Sprint(session.Info.Uin)),
		Guid: device.GUID,
	})
	secretBuf, e := crypto.AESGCMEncrypt(secret, shareKey)
	if e != nil {
		return nil, e
	}

	timestamp := time.Now().Unix()
	verifyHash, e := crypto.AESGCMEncrypt(crypto.SHA256Digest(binary.NewBuilder().
		WriteBytes(publicKey).
		WriteU32(1). // type
		WriteBytes(secretBuf).
		WriteU32(0).
		WriteU32(uint32(timestamp)).
		ToBytes()), login_type.VERIFY_HASH_KEY)
	if e != nil {
		return nil, e
	}

	data, _ := proto.Marshal(&login.KeyExchangeRequest{
		PublicKey:  publicKey,
		Type:       proto.Some[uint32](1),
		Secret:     secretBuf,
		Timestamp:  proto.Some[int64](timestamp),
		VerifyHash: verifyHash,
	})
	return login_type.AttributeSsoKeyExchange.NewSsoPacket(0, data), nil
}

func ParseKeyExchangePacket(session *auth.Session, pkt *sso_type.SsoPacket) error {
	rsp, e := proto.Unmarshal[login.KeyExchangeResponse](pkt.Data)
	if e != nil {
		return e
	}

	shareKey, e := ecdh.P256.Exange(rsp.PublicKey)
	if e != nil {
		return e
	}

	payload, e := crypto.AESGCMDecrypt(rsp.Secret, shareKey)
	if e != nil {
		return e
	}

	secret, e := proto.Unmarshal[login.KeyExchangeResponseSecret](payload)
	if e != nil {
		return e
	}

	session.State.KeyExchange = &auth.KeyExchange{
		SessionKey:    secret.SessionKey,
		SessionTicket: secret.SessionTicket,
	}
	return nil
}
