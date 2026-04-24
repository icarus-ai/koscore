package sign

import (
	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/utils/types"
)

type (
	rsp_sign_v2 struct {
		Code  int    `json:"code"`
		Msg   string `json:"message"`
		Value Value  `json:"value"`
	}

	Value struct {
		Sign  types.Bytes `json:"sec_sign"`
		Token types.Bytes `json:"sec_token"`
		Extra types.Bytes `json:"sec_extra"`
	}
)

type (
	Provider interface {
		Sign(cmd string, seq uint32, data []byte) (*Value, error)
		AddRequestHeader(heads types.MapSS)
		AddSignServer(servers ...string)
		GetSignServer() []string
		SetAppInfo(app *auth.AppInfo)
		Release()
		Reset()
		GetStat() string
	}
)
