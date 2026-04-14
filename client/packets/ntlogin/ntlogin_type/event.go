package ntlogin_type

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/login"
	"github.com/kernel-ai/koscore/utils/types"
)

type INTLoginRsp struct {
	State login.NTLoginRetCode
	Tips  types.Strings
}

type UnusualEasyLoginReq struct{}

type NewDeviceLoginReq struct{ Sig []byte }

type RefreshTicketReq struct{}
type RefreshA2Req struct{}

type PasswordLoginReq struct {
	password string
	captcha  types.Strings
}

type PasswordLoginRsp struct {
	INTLoginRsp
	JumpingUrl string
}

type EasyLoginRsp struct {
	INTLoginRsp
	UnusualSigs []byte
}

type (
	UnusualEasyLoginRsp = INTLoginRsp
	NewDeviceLoginRsp   = INTLoginRsp
	RefreshTicketRsp    = INTLoginRsp
	RefreshA2Rsp        = INTLoginRsp
)
