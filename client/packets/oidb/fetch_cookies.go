package oidb

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/types"
)

func BuildFetchCookiesPacket(domains []string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x102A, 0, &oidb.D102AReqBody{Domain: domains}, false, false)
}

func ParseFetchCookiesPacket(data []byte) (types.MapSS, error) {
	rsp, e := ParseOidbPacket[oidb.D102ARspBody](data)
	if e != nil {
		return nil, e
	}
	return rsp.PsKeys, nil
}
