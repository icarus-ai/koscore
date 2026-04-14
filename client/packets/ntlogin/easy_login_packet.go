package ntlogin

import (
	"errors"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/ntlogin/ntlogin_type"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/login"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/types"
)

func BuildEasyLoginPacket(version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session) (*sso_type.SsoPacket, error) {
	if len(session.Sig.A1) == 0 {
		return nil, errors.New("invalid operation exception: A1 is not set")
	}
	data, e := nt_login_encode_common(version, device, session, &login.NTLoginEasyLoginReqBody{A1: session.Sig.A1})
	if e != nil {
		return nil, e
	}
	return ntlogin_type.AttributeEasyLogin.NewSsoPacket(0, data), nil
}

func ParseEasyLoginPacket(session *auth.Session, pkt *sso_type.SsoPacket) (ret *ntlogin_type.EasyLoginRsp, e error) {
	state, info, rsp, ee := nt_login_decode_common[login.NTLoginEasyLoginRspBody](session, pkt.Data)
	if ee != nil {
		return nil, ee
	}

	ret = &ntlogin_type.EasyLoginRsp{INTLoginRsp: ntlogin_type.INTLoginRsp{State: state}}
	switch state {
	case login.NTLoginRetCode_SUCCESS:
		if rsp.Tickets == nil {
			return nil, errors.New("invalid operation exception: tickets is nil")
		}
		nt_login_save_ticket(session, rsp.Tickets)
	case login.NTLoginRetCode_ERROR_UNUSUAL_DEVICE:
		ret.UnusualSigs = rsp.SecProtect.UnusualDeviceCheckSig
	default:
		if info != nil {
			ret.Tips = types.Strings{info.StrTipsTitle.Unwrap(), info.StrTipsContent.Unwrap()}
		}
	}
	return
}

func BuildUnusualEasyLoginPacket(version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session) (*sso_type.SsoPacket, error) {
	if len(session.Sig.A1) == 0 {
		return nil, errors.New("invalid operation exception: A1 is not set")
	}
	data, e := nt_login_encode_common(version, device, session, &login.NTLoginEasyLoginUnusualDeviceReqBody{A1: session.Sig.A1})
	if e != nil {
		return nil, e
	}
	return ntlogin_type.AttributeUnusualEasyLogin.NewSsoPacket(0, data), nil
}

func ParseUnusualEasyLoginPacket(session *auth.Session, pkt *sso_type.SsoPacket) (ret *ntlogin_type.EasyLoginRsp, e error) {
	state, info, rsp, ee := nt_login_decode_common[login.NTLoginEasyLoginUnusualDeviceRspBody](session, pkt.Data)
	if ee != nil {
		return nil, ee
	}

	ret = &ntlogin_type.EasyLoginRsp{INTLoginRsp: ntlogin_type.INTLoginRsp{State: state}}
	if state == login.NTLoginRetCode_SUCCESS {
		if rsp.Tickets == nil {
			return nil, errors.New("invalid operation exception: tickets is nil")
		}
		nt_login_save_ticket(session, rsp.Tickets)
	} else if info != nil {
		ret.Tips = types.Strings{info.StrTipsTitle.Unwrap(), info.StrTipsContent.Unwrap()}
	}
	return
}
