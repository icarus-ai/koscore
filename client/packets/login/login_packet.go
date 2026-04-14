package login

import (
	"fmt"

	"github.com/fumiama/gofastTEA"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/login/login_type"
	"github.com/kernel-ai/koscore/client/packets/login/wtlogin"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/system"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
	"github.com/kernel-ai/koscore/utils/proto"
	"github.com/kernel-ai/koscore/utils/types"
	//"github.com/kernel-ai/koscore/utils/comm"
)

func BuildLoginPacket(version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session, req *login_type.LoginReq) *sso_type.SsoPacket {
	switch req.Cmd {
	case login_type.LoginTgtgt:
		return login_type.AttributeLogin.NewSsoPacket(0, wtlogin.BuildOicq09(version, device, session))
	default:
		return nil
	}
}

func ParseLoginPacket(session *auth.Session, pkt *sso_type.SsoPacket) (*login_type.LoginRsp, error) {
	/*
		comm.LOGD("Command: %s", pkt.Command)
		comm.LOGD("Sequence: %d", pkt.Sequence)
		comm.LOGD("Data: %X", pkt.Data)
		comm.LOGD("Extra: %s", pkt.Extra)
		comm.LOGD("RetCode: %d", pkt.RetCode)
		comm.LOGD("Sig:\n%s", ctx.Session.Sig.String())
		comm.LOGD("State:\n%s", ctx.Session.State.String())
	*/
	cmd, rsp, e := wtlogin.Parse(session, pkt.Data)
	if e != nil {
		return nil, e
	}

	//comm.LOGD("cmd: %02X", cmd)
	//comm.LOGD("rsp: %X", rsp)

	if cmd != 0x810 {
		return nil, fmt.Errorf("logins::ParsePacket: cmd: %02X", cmd)
	}

	reader := binary.NewReader(rsp)
	_ = login_type.LoginCommand(reader.ReadU16()) // internal_cmd
	state := reader.ReadU8()
	tlvs := reader.ReadTlv()

	//comm.LOGD("internal_cmd: %02X", internal_cmd)
	//comm.LOGD("state: %02X", state)
	//comm.LOGD("tlvs: size: %d 0x146: %d 0x119: %d", len(tlvs), len(tlvs[0x146]), len(tlvs[0x119]))

	ret := &login_type.LoginRsp{
		RetCode: state,
		State:   login_type.LoginState(state),
		Tlvs:    tlvs,
	}
	if ret.State == login_type.LoginSuccess {
		if v, ok := tlvs[0x119]; ok {
			tlv119 := tea.NewTeaCipher(session.Sig.TgtgtKey).Decrypt(v)
			ret.Tlvs = binary.NewReader(tlv119).ReadTlv() // tlvCollection
		}
	} else {
		if v, ok := tlvs[0x146]; ok {
			reader = binary.NewReader(v)
			code := reader.ReadU32() // error code
			title := reader.ReadLengthString(prefix.Int16 | prefix.LengthOnly)
			message := reader.ReadLengthString(prefix.Int16 | prefix.LengthOnly)
			ret.Err = fmt.Sprintf("%s(%d): %s", title, code, message)
		}
	}
	return ret, nil
}

func ParseLoginSig(session *auth.Session, tlvs types.Tlvs) (e error) {
	for k, v := range tlvs {
		switch k {
		case 0x103:
			session.Sig.StWeb = v
		case 0x108:
			session.Sig.Ksid = v
		case 0x10A:
			session.Sig.A2 = v
		case 0x143:
			session.Sig.D2 = v
		case 0x10C:
			session.Sig.A1Key = v
		case 0x10D:
			session.Sig.A2Key = v
		case 0x10E:
			session.Sig.StKey = v
		case 0x114:
			session.Sig.St = v
		case 0x11A: // bot_info_data
			r := binary.NewReader(v)
			r.ReadU16() // face_id
			session.Info.Age = uint32(r.ReadU8())
			session.Info.Gender = uint32(r.ReadU8())
			session.Info.Name = r.ReadLengthString(prefix.Int8 | prefix.LengthOnly)
		case 0x120:
			session.Sig.SKey = v
		case 0x133:
			session.Sig.WtSessionTicket = v
		case 0x134:
			session.Sig.WtSessionTicketKey = v
		case 0x305:
			session.Sig.D2Key = v
		case 0x106:
			session.Sig.A1 = v
		case 0x16A:
			session.Sig.NoPicSig = v
		case 0x16D:
			session.Sig.SuperKey = v
		case 0x512:
			session.Sig.PsKey = make(types.MapSS)
			r := binary.NewReader(v)
			domainCount := int(r.ReadU16())
			for i := 0; i < domainCount; i++ {
				domain := r.ReadLengthString(prefix.Int16 | prefix.LengthOnly)
				key := r.ReadLengthString(prefix.Int16 | prefix.LengthOnly)
				r.ReadLengthString(prefix.Int16 | prefix.LengthOnly) // pt4Token
				session.Sig.PsKey[domain] = key
			}
		case 0x543: // login_resp_data
			rsp, ee := proto.Unmarshal[system.ThirdPartyLoginResponse](v)
			if ee != nil {
				e = ee
				continue
			}
			session.Info.Uid = rsp.CommonInfo.RspNT.Uid.Unwrap()
			//default: comm.LOGD("0x%X", k) // 0x118 0x11F 0x130 0x138 0x167 0x163 0x510 0x523 0x550
		}
	}
	return
}

func BuildLoginPacketAndroid(version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session, req *login_type.LoginReq) *sso_type.SsoPacket {
	panic(types.ERROR_NOT_IMPL)
}
func ParseLoginPacketAndroid(session *auth.Session, pkt *sso_type.SsoPacket) (*login_type.LoginRsp, error) {
	panic(types.ERROR_NOT_IMPL)
}
