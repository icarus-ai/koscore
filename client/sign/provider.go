package sign

import (
	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/common"
	"github.com/kernel-ai/koscore/utils/types"
)

type (
	Response struct {
		Platform string `json:"platform"`
		Version  string `json:"version"`
		Value    struct {
			Sign  types.Bytes `json:"sign"`
			Extra types.Bytes `json:"extra"`
			Token types.Bytes `json:"token"`
		} `json:"value"`
	}
	ResponseV2 struct {
		Code  int    `json:"code"`
		Msg   string `json:"message"`
		Value struct {
			Sign  types.Bytes `json:"sec_sign"`
			Token types.Bytes `json:"sec_token"`
			Extra types.Bytes `json:"sec_extra"`
		} `json:"value"`
	}
)

type (
	header map[string]string

	Provider interface {
		Sign(cmd string, seq uint32, data []byte) (*common.SsoSecureInfo, error)
		AddRequestHeader(heads header)
		AddSignServer(servers ...string)
		GetSignServer() []string
		SetAppInfo(app *auth.AppInfo)
		Release()
		Reset()
		GetStat() string
	}
)
