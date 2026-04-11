package oidb

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

type FetchClientKeyRep struct {
	ClientKey  string
	Expiration uint32
}

func BuildFetchClientKeyPacket() (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x102A, 1, &oidb.D102AReqBody{}, false, false)
}

func ParseFetchClientKeyPacket(data []byte) (*FetchClientKeyRep, error) {
	rsp, e := ParseOidbPacket[oidb.D102ARspBody](data)
	if e != nil {
		return nil, e
	}
	return &FetchClientKeyRep{
		ClientKey:  rsp.ClientKey.Unwrap(),
		Expiration: rsp.Expiration.Unwrap(),
	}, nil
}
