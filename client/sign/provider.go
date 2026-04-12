package sign

import (
	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/common"
	"github.com/kernel-ai/koscore/utils/types"
)

type ResponseV2 struct {
	Code  int    `json:"code"`
	Msg   string `json:"message"`
	Value struct {
		Sign  types.Bytes `json:"sec_sign"`
		Token types.Bytes `json:"sec_token"`
		Extra types.Bytes `json:"sec_extra"`
	} `json:"value"`
}

type (
	Provider interface {
		Sign(cmd string, seq uint32, data []byte) (*common.SsoSecureInfo, error)
		AddRequestHeader(heads types.MapSS)
		AddSignServer(servers ...string)
		GetSignServer() []string
		SetAppInfo(app *auth.AppInfo)
		Release()
		Reset()
		GetStat() string
	}
)
