package login_type

import (
	"fmt"

	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

// ***** event *****

// trans_emp_31_req
type TransEmpReq31 struct {
	QRCcodeSize uint32
	UnusualSig  []byte
}

// trans_emp_31_rsp
type TransEmpRsp31 struct {
	Url   string
	Image []byte
	QrSig []byte
}

// trans_emp_12_req
type TransEmpReq12 struct{}

// trans_emp_12_rsp
type TransEmpRsp12 struct {
	State TransEmpState
	Uin   uint64
	// data
	TgtgtKey     []byte
	NoPicSig     []byte
	TempPassword []byte
}

// ***** state *****

// protocols pc and android_watch
var AttributeTransEmp = sso_type.NewServiceAttributeD2Empty("wtlogin.trans_emp")

// trans_emp_stste
type TransEmpState uint8

const (
	TransEmpConfirmed         TransEmpState = 0
	TransEmpCodeExpired       TransEmpState = 17
	TransEmpWaitingForScan    TransEmpState = 48
	TransEmpWaitingForConfirm TransEmpState = 53
	TransEmpCanceled          TransEmpState = 54
	TransEmpInvalid           TransEmpState = 144
)

func (m TransEmpState) String() string {
	switch m {
	case TransEmpConfirmed:
		return "已确认"
	case TransEmpCodeExpired:
		return "已过期"
	case TransEmpWaitingForScan:
		return "等待扫描"
	case TransEmpWaitingForConfirm:
		return "等待确认"
	case TransEmpCanceled:
		return "已取消"
	case TransEmpInvalid:
		return "无效"
	default:
		return fmt.Sprintf("qrcode state unknow: %d", m)
	}
}
