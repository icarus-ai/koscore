package ntlogin

import (
	"fmt"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/login"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
	"github.com/kernel-ai/koscore/utils/types"
)

// PC
func nt_login_encode_common[T any](version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session, body *T) ([]byte, error) {
	if session.State.KeyExchange == nil {
		return nil, fmt.Errorf("invalid operation exception: key exchange session is not initialized.")
	}
	data, e := proto.Marshal(nt_login_build_common(version, device, session, body))
	if e != nil {
		return nil, e
	}
	data, e = crypto.AESGCMEncrypt(data, session.State.KeyExchange.SessionKey)
	if e != nil {
		return nil, e
	}
	return proto.Marshal(&login.NTLoginForwardRequest{
		SessionTicket: session.State.KeyExchange.SessionTicket,
		Buffer:        data,
		Type:          proto.Some[uint32](1),
	})
}

func nt_login_decode_common[T any](session *auth.Session, payload []byte) (login.NTLoginRetCode, *login.NTLoginErrorInfo, *T, error) {
	if session.State.KeyExchange == nil {
		return 0, nil, nil, fmt.Errorf("invalid operation exception: key exchange session is not initialized.")
	}

	forward, e := proto.Unmarshal[login.NTLoginForwardRequest](payload)
	if e != nil {
		return 0, nil, nil, e
	}

	payload, e = crypto.AESGCMDecrypt(forward.Buffer, session.State.KeyExchange.SessionKey)
	if e != nil {
		return 0, nil, nil, e
	}

	common, e := proto.Unmarshal[login.NTLoginCommon](payload)
	if e != nil {
		return 0, nil, nil, e
	}

	rsp, e := proto.Unmarshal[T](common.Body)
	if e != nil {
		return 0, nil, nil, e
	}

	if common.Head.ErrorInfo == nil || common.Head.ErrorInfo.ErrCode.Unwrap() == 0 {
		return login.NTLoginRetCode_SUCCESS, nil, rsp, nil
	}
	return login.NTLoginRetCode(common.Head.ErrorInfo.ErrCode.Unwrap()), common.Head.ErrorInfo, rsp, nil
}

func nt_login_build_common[T any](version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session, body *T) (ret *login.NTLoginCommon) {
	ret = &login.NTLoginCommon{
		Head: &login.NTLoginHead{
			UserInfo: &login.NTLoginUserInfo{
				Account: proto.Some(fmt.Sprint(session.Info.Uin)),
			},
			ClientInfo: &login.NTLoginClientInfo{
				DeviceType: proto.Some(version.OS.String()),
				DeviceName: proto.Some(device.DeviceName),
				Platform:   proto.Some[login.NTLoginPlatform](version.OS.ProtocolCode()),
				Guid:       device.GUID,
			},
			AppInfo: &login.NTLoginAppInfo{
				Version: proto.Some(version.Kernel),
				AppId:   proto.Some(int32(version.AppId)),
				AppName: proto.Some(version.PackageName),
				Qua:     proto.Some(version.QUA),
			},
			SdkInfo: &login.NTLoginSdkInfo{Version: proto.Some[uint32](1)},
			Cookie: func() *login.NTLoginCookie {
				if session.State.Cookie == "" {
					return &login.NTLoginCookie{}
				}
				return &login.NTLoginCookie{CookieContent: proto.Some(session.State.Cookie)}
			}(),
		},
	}
	ret.Body, _ = proto.Marshal(body)
	return
}

func nt_login_save_ticket(session *auth.Session, tickets *login.NTLoginTickets) {
	session.Sig.A1 = tickets.A1
	session.Sig.A2 = tickets.A2
	session.Sig.D2 = tickets.D2
	session.Sig.D2Key = tickets.D2Key
}

func nt_login_encode_common_android[T any](version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session, body *T) ([]byte, error) {
	panic(types.ERROR_NOT_IMPL)
}
func nt_login_decode_common_android[T any](session *auth.Session, body []byte) (login.NTLoginRetCode, *login.NTLoginErrorInfo, *T, error) {
	panic(types.ERROR_NOT_IMPL)
}
