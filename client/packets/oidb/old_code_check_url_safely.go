package oidb

import (
	"errors"

	"github.com/RomiChan/protobuf/proto"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

// ref https://github.com/Mrs4s/MiraiGo/blob/master/client/security.go

type URLSecurityLevel int

const (
	URLSecurityLevelSafe URLSecurityLevel = iota + 1
	URLSecurityLevelUnknown
	URLSecurityLevelDanger
)

func (m URLSecurityLevel) String() string {
	switch m {
	case URLSecurityLevelSafe:
		return "safe"
	case URLSecurityLevelDanger:
		return "danger"
	case URLSecurityLevelUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

func BuildURLCheckRequest(botuin uint64, url string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0xBCB, 0, &oidb.OidbSvcTrpcTcp0XBCB_0_ReqBody{
		CheckUrlReq: &oidb.CheckUrlReq{
			Url:         []string{url},
			QqPfTo:      proto.String("mqq.group"),
			Type:        proto.Uint32(2),
			SendUin:     proto.Uint64(botuin),
			ReqType:     proto.String("webview"),
			OriginalUrl: proto.String(url),
			IsArk:       proto.Bool(false),
			IsFinish:    proto.Bool(false),
			SrcUrls:     []string{url},
			SrcPlatform: proto.Uint32(1),
			Qua:         proto.String("AQQ_2013 4.6/2013 8.4.184945&NA_0/000000&ADR&null18&linux&2017&C2293D02BEE31158&7.1.2&V3"),
		}}, false, false)
}

func ParseURLCheckResponse(data []byte) (URLSecurityLevel, error) {
	rsp, e := ParseOidbPacket[oidb.OidbSvcTrpcTcp0XBCB_0_RspBody](data)
	if e != nil {
		return URLSecurityLevelUnknown, e
	}
	if rsp.CheckUrlRsp == nil || len(rsp.CheckUrlRsp.Results) == 0 {
		return URLSecurityLevelUnknown, errors.New("response is empty")
	}
	if rsp.CheckUrlRsp.Results[0].JumpUrl.IsSome() {
		return URLSecurityLevelDanger, nil
	}
	if rsp.CheckUrlRsp.Results[0].UmrType.Unwrap() == 2 {
		return URLSecurityLevelSafe, nil
	}
	return URLSecurityLevelUnknown, nil
}
