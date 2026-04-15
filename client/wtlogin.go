package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	pb_login "github.com/kernel-ai/koscore/client/packets/pb/v2/login"

	"github.com/kernel-ai/koscore/client/event"
	"github.com/kernel-ai/koscore/client/packets/login"
	"github.com/kernel-ai/koscore/client/packets/login/login_type"
	"github.com/kernel-ai/koscore/client/packets/ntlogin"
	"github.com/kernel-ai/koscore/client/packets/ntlogin/ntlogin_type"
	"github.com/kernel-ai/koscore/client/packets/system"
	"github.com/kernel-ai/koscore/client/packets/system/system_type"
)

func (m *QQClient) heart_beat_loop(interval time.Duration) {
	if !m.sso_context.IsConnect() {
		_ = m.sso_context.Connect()
	}
	if m.is_heart_beat {
		return
	}
	m.is_heart_beat = true
	ticker := time.NewTicker(interval * time.Second)
	pkt := system_type.AttributeHeartbeat.NewSsoPacket(0, system_type.HeartbeatReq)
	go func() {
		for m.is_heart_beat {
			<-ticker.C
			pkt.Sequence = m.session.GetAndIncreaseSequence()
			if e := m.sso_context.SendPacket(pkt); e != nil {
				m.LOGW("heart_beat: send: %s", e)
			}
		}
		m.is_heart_beat = false
		ticker.Stop()
	}()
}

// protocol pc tx_interval 360s
func (m *QQClient) sso_heart_beat_loop(interval time.Duration) {
	if m.Online.Load() {
		return
	}
	ticker := time.NewTicker(interval * time.Second)
	pkt := system.BuildSsoHeartBeatPacket()
	go func() {
		for m.Online.Load() {
			<-ticker.C
			//m.LOGD("sso_heart_beat: send: cmd: %s seq: %d data: %X", sso.Command, sso.Sequence, sso.Data)
			pkt.Sequence = m.session.GetAndIncreaseSequence()
			rsp, e := m.sso_context.SendPacketAndWait(pkt)
			if e != nil {
				m.LOGW("sso_heart_beat: send: %v", e)
				continue
			}
			_, e = system.ParseSsoHeartBeatPacket(rsp)
			if e != nil {
				m.LOGW("sso_heart_beat: rsp: %v", e)
			}
		}
		ticker.Stop()
	}()
}

// qrcode login
//   FetchQRode
//   GetRCodeResult
//   QRCodeLogin

// qrcode_size 2
func (m *QQClient) FetchQRode(qrcode_size uint32, unusual_sig []byte) (*login_type.TransEmpRsp31, error) {
	m.heart_beat_loop(2)
	sso, e := m.sendOidbPacketAndWait(login.BuildTransEmpPacket[login_type.TransEmpReq31](m.version, m.device, m.session, &login_type.TransEmpReq31{
		QRCcodeSize: qrcode_size,
		UnusualSig:  unusual_sig,
	}))
	if e != nil {
		return nil, errors.Wrap(e, "fetch qrcode: send")
	}
	emp, e := login.ParseTransEmpPacket[login_type.TransEmpRsp31](m.session, sso)
	if e != nil {
		return nil, errors.Wrap(e, "fetch qrcode: parse")
	}
	//comm.LOGD("qr: %s %X", emp.Url, emp.Image)
	//comm.LOGD("qr: sig %X", emp.QrSig)
	m.session.State.QrSig = emp.QrSig
	return emp, nil
}

func (m *QQClient) GetRCodeResult() (login_type.TransEmpState, error) {
	sso, e := m.sendOidbPacketAndWait(login.BuildTransEmpPacket[login_type.TransEmpReq12](m.version, m.device, m.session, nil))
	if e != nil {
		return login_type.TransEmpInvalid, errors.Wrap(e, "get qrcode result: send")
	}
	emp, e := login.ParseTransEmpPacket[login_type.TransEmpRsp12](m.session, sso)
	if e != nil {
		return login_type.TransEmpInvalid, errors.Wrap(e, "get qrcode result: parse")
	}
	if emp.State == login_type.TransEmpConfirmed {
		m.session.Sig.TgtgtKey = emp.TgtgtKey
		m.session.Sig.NoPicSig = emp.NoPicSig
		m.session.Sig.A1 = emp.TempPassword
		m.session.Info.Uin = emp.Uin
		return login_type.TransEmpConfirmed, nil
	}
	return emp.State, nil
}

func (m *QQClient) QRCodeLogin() error {
	sso, e := m.sendOidbPacketAndWait(login.BuildLoginPacket(m.version, m.device, m.session, &login_type.LoginReq{Cmd: login_type.LoginTgtgt}))
	if e != nil {
		return errors.Wrap(e, "qrcode login: send")
	}
	rsp, e := login.ParseLoginPacket(m.session, sso)
	if e != nil {
		return errors.Wrap(e, "qrcode login: parse")
	}
	if rsp.State == login_type.LoginSuccess {
		if e = login.ParseLoginSig(m.session, rsp.Tlvs); e != nil {
			return errors.Wrap(e, "qrcode login: parse login sig")
		}
		return m.register()
	}
	return errors.New(rsp.State.String())
}

// 上线
func (m *QQClient) register() error {
	sso, e := m.sendOidbPacketAndWait(system.BuildInfoSyncPacket(m.version, m.device))
	if e != nil {
		return errors.Wrap(e, "register: send")
	}
	rsp := system.ParseInfoSyncPacket(sso)
	if strings.Contains(rsp.Message, "register success") {
		m.Online.Store(true)
		m.sso_heart_beat_loop(270)
		//if protocol.IsAndroid { _timers[ExchangeEmpTag].Change(TimeSpan.Zero, TimeSpan.FromDays(1)) }
		return nil
	}
	return fmt.Errorf("register: %s", rsp.Message)
}

func (m *QQClient) keyExchange() error {
	pkt, e := login.BuildKeyExchangePacket(m.device, m.session)
	if e != nil {
		return errors.Wrap(e, "key exchange: build")
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return errors.Wrap(e, "key exchange: send")
	}
	if e = login.ParseKeyExchangePacket(m.session, pkt); e != nil {
		return errors.Wrap(e, "key exchange: parse")
	}
	return nil
}

func (m *QQClient) easyLogin() (*ntlogin_type.EasyLoginRsp, error) {
	pkt, e := ntlogin.BuildEasyLoginPacket(m.version, m.device, m.session)
	if e != nil {
		return nil, errors.Wrap(e, "easy login: build")
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, errors.Wrap(e, "easy login: send")
	}
	ret, e := ntlogin.ParseEasyLoginPacket(m.session, pkt)
	if e != nil {
		return nil, errors.Wrap(e, "easy login: parse")
	}
	if ret.State == pb_login.NTLoginRetCode_SUCCESS {
		return ret, m.register()
	}
	return ret, nil
}

func (m *QQClient) UnusualEasyLogin() error {
	pkt, e := ntlogin.BuildUnusualEasyLoginPacket(m.version, m.device, m.session)
	if e != nil {
		return errors.Wrap(e, "unusual easy login: build")
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return errors.Wrap(e, "unusual easy login: send")
	}
	ret, e := ntlogin.ParseUnusualEasyLoginPacket(m.session, pkt)
	if e != nil {
		return errors.Wrap(e, "unusual easy login: parse")
	}
	if ret.State == pb_login.NTLoginRetCode_SUCCESS {
		return m.register()
	}
	return fmt.Errorf("unusual easy login: %s (%s)", ret.Tips.String(), ntlogin_type.NTLoginRetCodeString(ret.State))
}

func (m *QQClient) Logout() error {
	pkt, e := m.sendOidbPacketAndWait(login.BuildSsoUnregisterPacket())
	if e != nil {
		return errors.Wrap(e, "logout: send")
	}
	rsp, e := login.ParseSsoUnregisterPacket(pkt.Data)
	if e != nil {
		return errors.Wrap(e, "logout: parse")
	}
	if strings.Contains(rsp.Msg.Unwrap(), "unregister success") {
		m.LOGD("sso_unregister: logout success")
		m.is_heart_beat = false
		m.sso_context.Disconnect()
		return nil
	}
	return fmt.Errorf("logout: %s", rsp.Msg.Unwrap())
}

// 快读登录
func (m *QQClient) FastLogin() error {
	m.heart_beat_loop(2)
	if len(m.session.Sig.A2) > 0 && len(m.session.Sig.D2) > 0 {
		if m.Online.Load() {
			return event.ErrAlreadyOnline
		}
		m.LOGD("valid session detected, doing online task")
		return m.register()
	}
	return errors.New("no login cache")
}

// return state unusual error
func (m *QQClient) ExchangeEasyLogin() (login_type.LoginState, []byte, error) {
	m.heart_beat_loop(2)
	if m.version.OS.IsPC() && len(m.session.Sig.A1) > 0 {
		if e := m.keyExchange(); e != nil {
			return 0, nil, e
		}
		rsp, e := m.easyLogin()
		if e != nil {
			return 0, nil, e
		}
		switch rsp.State {
		case pb_login.NTLoginRetCode_SUCCESS:
			return login_type.LoginSuccess, nil, nil
		case pb_login.NTLoginRetCode_ERROR_UNUSUAL_DEVICE:
			if len(rsp.UnusualSigs) > 0 {
				return 0, rsp.UnusualSigs, nil
			}
			return 0, nil, errors.New("easy login: unusual sigs nil")
		default:
			return 0, nil, fmt.Errorf("easy login: state: %d, %s", rsp.State, rsp.Tips.String())
		}
	}
	return login_type.LoginUnknown, nil, nil
}

//func (m *QQClient) PasswordLogin() (login_type.LoginState, error) { panic(types.ERROR_NOT_IMPL) }
//func (c *QQClient) SubmitCaptcha(ticket, randStr, aid string) (login_type.LoginState, error) { panic(types.ERROR_NOT_IMPL) }
