package ntlogin_type

import "github.com/kernel-ai/koscore/utils/types"

type INTLoginRsp struct {
	State NTLoginRetCode
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
	State      NTLoginRetCode
	Tips       types.Strings
	JumpingUrl string
}

type EasyLoginRsp struct {
	State       NTLoginRetCode
	Tips        types.Strings
	UnusualSigs []byte
}

type UnusualEasyLoginRsp struct {
	State NTLoginRetCode
	Tips  types.Strings
}

type NewDeviceLoginRsp struct {
	State NTLoginRetCode
	Tips  types.Strings
}

type RefreshTicketRsp struct {
	State NTLoginRetCode
	Tips  types.Strings
}

type RefreshA2Rsp struct {
	State NTLoginRetCode
	Tips  types.Strings
}
