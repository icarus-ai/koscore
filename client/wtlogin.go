package client

import (
	"strings"
	"time"

	"github.com/kernel-ai/koscore/client/packets/login"
	"github.com/kernel-ai/koscore/client/packets/login/login_type"
	"github.com/kernel-ai/koscore/client/packets/ntlogin"
	"github.com/kernel-ai/koscore/client/packets/ntlogin/ntlogin_type"
	"github.com/kernel-ai/koscore/client/packets/system"
	"github.com/kernel-ai/koscore/client/packets/system/system_type"
)

func (m *QQClient) HeartBeatLoop(interval time.Duration) {
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
			_, e := m.sso_context.SendPacketAndWait(pkt)
			if e != nil {
				m.LOGW("heart_beat: send: %s", e)
			}
		}
		m.is_heart_beat = false
		ticker.Stop()
	}()
}

// protocol pc tx_interval 360s
func (m *QQClient) sso_heart_beat_loop(interval time.Duration) {
	if m.is_online.Load() {
		return
	}
	ticker := time.NewTicker(interval * time.Second)
	pkt := system.BuildSsoHeartBeatPacket()
	go func() {
		for m.is_online.Load() {
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
	sso, e := m.sso_context.SendPacketAndWait(login.BuildTransEmpPacket[login_type.TransEmpReq31](m.version, m.device, m.session, &login_type.TransEmpReq31{
		QRCcodeSize: qrcode_size,
		UnusualSig:  unusual_sig,
	}))
	if e != nil {
		return nil, e
	}
	emp, e := login.ParseTransEmpPacket[login_type.TransEmpRsp31](m.session, sso)
	if e != nil {
		return nil, e
	}

	//comm.LOGD("qr: %s %X", emp.Url, emp.Image)
	//comm.LOGD("qr: sig %X", emp.QrSig)
	m.session.State.QrSig = emp.QrSig
	return emp, nil
}

func (m *QQClient) GetRCodeResult() (login_type.TransEmpState, error) {
	sso, e := m.sso_context.SendPacketAndWait(login.BuildTransEmpPacket[login_type.TransEmpReq12](m.version, m.device, m.session, nil))
	if e != nil {
		return login_type.TransEmpInvalid, e
	}
	emp, e := login.ParseTransEmpPacket[login_type.TransEmpRsp12](m.session, sso)
	if e != nil {
		return login_type.TransEmpInvalid, e
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

func (m *QQClient) QRCodeLogin() (login_type.LoginState, error) {
	sso, e := m.sso_context.SendPacketAndWait(login.BuildLoginPacket(m.version, m.device, m.session, &login_type.LoginReq{Cmd: login_type.LoginTgtgt}))
	if e != nil {
		m.LOGE("QRCodeLogin: send LoginPacket: %v", e)
		return login_type.LoginUnknown, e
	}
	rsp, e := login.ParseLoginPacket(m.session, sso)
	if e != nil {
		m.LOGE("QRCodeLogin: parse LoginPacket: %v", e)
		return login_type.LoginUnknown, e
	}
	if rsp.State == login_type.LoginSuccess {
		if e := login.ParseLoginSig(m.session, rsp.Tlvs); e != nil {
			m.LOGW("QRCodeLogin: parse LoginSig: %v", e)
			return rsp.State, e
		}
	}
	return rsp.State, nil
}

// 上线
func (m *QQClient) Online() bool {
	sso, e := m.sso_context.SendPacketAndWait(system.BuildInfoSyncPacket(m.version, m.device))
	if e != nil {
		m.LOGE("online: send InfoSyncPacket: %v", e)
		return false
	}
	rsp := system.ParseInfoSyncPacket(sso)
	if strings.Contains(rsp.Message, "register success") {
		m.is_online.Store(true)
		m.sso_heart_beat_loop(270)
		//if protocol.IsAndroid { _timers[ExchangeEmpTag].Change(TimeSpan.Zero, TimeSpan.FromDays(1)) }
		return true
	}
	m.LOGE("online: message: %s", rsp.Message)
	return false
}

func (m *QQClient) KeyExchange() bool {
	pkt, e := login.BuildKeyExchangePacket(m.device, m.session)
	if e != nil {
		m.LOGE("KeyExchange: build KeyExchangePacket: %v", e)
		return false
	}
	if pkt, e = m.sso_context.SendPacketAndWait(pkt); e != nil {
		m.LOGE("KeyExchange: send KeyExchangePacket: %v", e)
		return false
	}
	if e = login.ParseKeyExchangePacket(m.session, pkt); e != nil {
		m.LOGE("KeyExchange: parse KeyExchangePacket: %v", e)
		return false
	}
	return true
}

func (m *QQClient) EasyLogin() (*ntlogin_type.EasyLoginRsp, error) {
	pkt, e := ntlogin.BuildEasyLoginPacket(m.version, m.device, m.session)
	if e != nil {
		m.LOGE("KeyExchange: build EasyLoginPacket: %v", e)
		return nil, e
	}
	if pkt, e = m.sso_context.SendPacketAndWait(pkt); e != nil {
		m.LOGE("KeyExchange: send EasyLoginPacket: %v", e)
		return nil, e
	}
	ret, e := ntlogin.ParseEasyLoginPacket(m.session, pkt)
	if e != nil {
		m.LOGE("KeyExchange: parse EasyLoginPacket: %v", e)
		return nil, e
	}
	return ret, nil
}

func (m *QQClient) UnusualEasyLogin() (*ntlogin_type.EasyLoginRsp, error) {
	pkt, e := ntlogin.BuildUnusualEasyLoginPacket(m.version, m.device, m.session)
	if e != nil {
		m.LOGE("KeyExchange: build EasyLoginPacket: %v", e)
		return nil, e
	}
	if pkt, e = m.sso_context.SendPacketAndWait(pkt); e != nil {
		m.LOGE("KeyExchange: send EasyLoginPacket: %v", e)
		return nil, e
	}
	ret, e := ntlogin.ParseEasyLoginPacket(m.session, pkt)
	if e != nil {
		m.LOGE("KeyExchange: parse EasyLoginPacket: %v", e)
		return nil, e
	}
	return ret, nil
}

func (m *QQClient) Logout() bool {
	pkt, e := m.sso_context.SendPacketAndWait(login.BuildSsoUnregisterPacket())
	if e != nil {
		m.LOGE("sso_unregister: send: %X", e)
		return false
	}
	rsp, e := login.ParseSsoUnregisterPacket(pkt.Data)
	if e != nil {
		m.LOGE("sso_unregister: parse: %X", e)
		return false
	}
	if strings.Contains(rsp.Msg.Unwrap(), "unregister success") {
		m.LOGD("sso_unregister: logout success")
		m.is_heart_beat = false
		m.sso_context.Disconnect()
		return true
	}
	m.LOGD("sso_unregister: logout failed: %s", rsp.Msg.Unwrap())
	return false
}

// 快读登录
