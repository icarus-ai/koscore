package login

import (
	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/login/login_type"
	"github.com/kernel-ai/koscore/client/packets/login/wtlogin"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/login"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
	"github.com/kernel-ai/koscore/utils/exception"
	"github.com/kernel-ai/koscore/utils/proto"
)

// trans_emp_packet

func BuildTransEmpPacket[T login_type.TransEmpReq31 | login_type.TransEmpReq12](version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session, req *T) *sso_type.SsoPacket {
	switch r := any(req).(type) {
	case *login_type.TransEmpReq31:
		return login_type.AttributeTransEmp.NewSsoPacket(0, wtlogin.BuildTransEmp31(version, device, session, r.UnusualSig, r.QRCcodeSize))
	case *login_type.TransEmpReq12:
		return login_type.AttributeTransEmp.NewSsoPacket(0, wtlogin.BuildTransEmp12(version, session))
	default:
		return nil
	}
}

func ParseTransEmpPacket[T login_type.TransEmpRsp31 | login_type.TransEmpRsp12](session *auth.Session, sso *sso_type.SsoPacket) (*T, error) {
	//comm.LOGD("transEmpService_Parse: raw %02X %X", len(sso.Data), sso.Data)
	cmd, rsp, e := wtlogin.Parse(session, sso.Data)
	if e != nil {
		return nil, e
	}
	cmd, rsp = wtlogin.ParseCode2dPacket(session, rsp)
	//comm.LOGD("m.wt.Parse: %02X %X", cmd, rsp)
	//comm.LOGD("m.wt.ParseCode2dPacket(rsp): cmd: %02X %X", cmd, rsp)

	byt := binary.NewReader(rsp)
	byt.ReadU16() // dummy
	byt.ReadU32() // appid
	retcode := byt.ReadU8()

	//comm.LOGD("dummy: %02X", dummy)
	//comm.LOGD("appid: %d", appid)
	//comm.LOGD("retcode: %d", retcode)
	//comm.LOGD("qrsig: %02X", qrsig)

	switch cmd {
	case 0x31:
		sig := byt.ReadLengthBytes(prefix.Int16 | prefix.LengthOnly)
		tlvs := byt.ReadTlv()
		ext, e := proto.Unmarshal[login.QrExtInfo](tlvs[0xD1])
		if e != nil {
			return nil, e
		}
		emp := &login_type.TransEmpRsp31{
			Url:   ext.QrUrl.Unwrap(),
			Image: tlvs[0x17],
			QrSig: sig,
		}
		return any(emp).(*T), nil
	case 0x12:
		emp := &login_type.TransEmpRsp12{
			State: login_type.TransEmpState(retcode),
		}
		if retcode == 0 {
			emp.Uin = byt.ReadU64()
			byt.ReadU32() // retry
			tlvs := byt.ReadTlv()
			emp.TgtgtKey = tlvs[0x1e]
			emp.NoPicSig = tlvs[0x19]
			emp.TempPassword = tlvs[0x18]
		}
		return any(emp).(*T), nil
	default:
		return nil, exception.NewNotSupportedException("unknown trans_emp command: %d", cmd)
	}
}
