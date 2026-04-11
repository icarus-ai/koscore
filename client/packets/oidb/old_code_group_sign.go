package oidb

import (
	"errors"
	"strconv"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

// 群打卡

type BotGroupClockInResult struct {
	Title          string // 今日已成功打卡
	KeepDayText    string // 已打卡N天
	GroupRankText  string // 群内排名第N位
	ClockInUtcTime int64  // 打卡时间
	DetailURL      string // Detail info url https://qun.qq.com/v2/signin/detail?...
}

func BuildGroupSignPacket(botuin, gin uint64, app_version string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0xEB7, 1, &oidb.OidbSvcTrpcTcp0XEB7_1_ReqBody{
		SignInWriteReq: &oidb.StSignInWriteReq{
			Uin:        strconv.Itoa(int(botuin)),
			GroupUin:   strconv.Itoa(int(gin)),
			AppVersion: app_version,
		},
	}, false, false)
}

func ParseGroupSignResp(data []byte) (*BotGroupClockInResult, error) {
	rsp, e := ParseOidbPacket[oidb.OidbSvcTrpcTcp0XEB7_1_RspBody](data)
	if e != nil {
		return nil, e
	}

	if rsp.SignInWriteRsp == nil || rsp.SignInWriteRsp.DoneInfo == nil {
		return nil, errors.New("SignInWriteRsp or SignInWriteRsp.DoneInfo nil")
	}

	ret := &BotGroupClockInResult{
		Title:       rsp.SignInWriteRsp.DoneInfo.Title,
		KeepDayText: rsp.SignInWriteRsp.DoneInfo.KeepDayText,
		DetailURL:   rsp.SignInWriteRsp.DoneInfo.DetailUrl,
	}
	if size := len(rsp.SignInWriteRsp.DoneInfo.ClockInInfo); size > 0 {
		ret.GroupRankText = rsp.SignInWriteRsp.DoneInfo.ClockInInfo[0]
		if size > 1 {
			ret.ClockInUtcTime, _ = strconv.ParseInt(rsp.SignInWriteRsp.DoneInfo.ClockInInfo[1], 10, 64)
		}
	}
	return ret, nil
}
