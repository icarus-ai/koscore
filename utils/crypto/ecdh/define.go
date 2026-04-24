package ecdh

import "encoding/hex"

var (
	P256 Exchanger
	S192 Exchanger
)

func init() {
	DEBUG = true
	k_p256_pub_src, _ := hex.DecodeString("049D1423332735980EDABE7E9EA451B3395B6F35250DB8FC56F25889F628CBAE3E8E73077914071EEEBC108F4E0170057792BB17AA303AF652313D17C1AC815E79")
	k_s192_pub_src, _ := hex.DecodeString("04928D8850673088B343264E0C6BACB8496D697799F37211DEB25BB73906CB089FEA9639B4E0260498B51A992D50813DA8")
	P256 = new_exchanger(&Prime256V1, false, k_p256_pub_src)
	S192 = new_exchanger(&Secp192k1, true, k_s192_pub_src)
}

type Exchanger interface {
	PublicKey() []byte
	SharedKey() []byte
	Exange(remote []byte) ([]byte, error)
}

type exchanger struct {
	provider *Provider
	public   []byte
	shared   []byte
	compress bool
}

func new_exchanger(curve *EllipticCurve, is_hash bool, pubkey []byte) Exchanger {
	ctx, e := NewEcdhProvider(curve)
	if e != nil {
		panic(e)
	}
	shk, e := ctx.KeyExchange(pubkey, is_hash)
	if e != nil {
		panic(e)
	}
	return &exchanger{
		provider: ctx,
		public:   ctx.PackPublic(is_hash),
		shared:   shk,
		compress: is_hash,
	}
}

func (e *exchanger) PublicKey() []byte { return e.public }
func (e *exchanger) SharedKey() []byte { return e.shared }
func (e *exchanger) Exange(key []byte) ([]byte, error) {
	return e.provider.KeyExchange(key, e.compress)
}
