package oidb

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/group_msg.go#L213

func BuildGetAtAllRemainRequest(uin, gin uint64) (*sso_type.SsoPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X8A7_0_ReqBody{
		SubCmd:                    1,
		LimitIntervalTypeForUin:   2,
		LimitIntervalTypeForGroup: 1,
		Uin:                       uin,
		GroupUin:                  gin,
	}
	return BuildOidbPacket(0x8A7, 0, body, false, true)
}

type AtAllRemainInfo struct {
	CanAtAll      bool
	CountForGroup uint32 // 当前群默认可用次数
	CountForUin   uint32 // 当前QQ剩余次数
}

func ParseGetAtAllRemainResponse(data []byte) (*AtAllRemainInfo, error) {
	rsp, err := ParseOidbPacket[oidb.OidbSvcTrpcTcp0X8A7_0_RspBody](data)
	if err != nil {
		return nil, err
	}
	return &AtAllRemainInfo{
		CanAtAll:      rsp.CanAtAll,
		CountForGroup: rsp.CountForGroup,
		CountForUin:   rsp.CountForUin,
	}, nil
}
