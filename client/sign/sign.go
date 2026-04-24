package sign

import (
	"errors"
	"math"
	"strings"

	"github.com/kernel-ai/koscore/utils/comm"
)

const server_latency_down = math.MaxUint32

var (
	ErrVersionMismatch    = errors.New("sign version mismatch")
	ErrAllSignServiceDown = errors.New("all sign service down")
	k_err_sign_rsp        = errors.New("sign rsp 0")
)

// signExtraHexLower = fmt.Sprintf("%x", proto.DynamicMessage{2: c.app.PackageSign}.Encode())

var sign_map map[string]uint8 // 只在启动时初始化, 无并发问题

func ContainSignPKG(cmd string) bool {
	_, ok := sign_map[cmd]
	if !ok {
		ok = strings.Contains(cmd, "OidbSvcTrpcTcp.0x")
	}
	if !ok {
		comm.LOGD("unsign cmd: %s", cmd)
	}
	return ok
}

func AddSignPKG(pkg string) { sign_map[pkg] = 1 }

func init() {
	sign_map = make(map[string]uint8)

	for _, cmd := range []string{
		"trpc.o3.ecdh_access.EcdhAccess.SsoEstablishShareKey",
		"trpc.o3.ecdh_access.EcdhAccess.SsoSecureAccess",
		"trpc.o3.report.Report.SsoReport",
		"MessageSvc.PbSendMsg",
		"wtlogin.trans_emp", //
		"wtlogin.login",
		"wtlogin.exchange_emp",
		"trpc.login.ecdh.EcdhService.SsoKeyExchange", //
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginNewDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLoginUnusualDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginUnusualDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginRefreshTicket",
		"trpc.login.ecdh.EcdhService.SsoNTLoginRefreshA2",
		"OidbSvcTrpcTcp.0x11ec_1",
		"OidbSvcTrpcTcp.0x758_1", // create group
		"OidbSvcTrpcTcp.0x7c1_1",
		"OidbSvcTrpcTcp.0x7c2_5", // request friend
		"OidbSvcTrpcTcp.0x10db_1",
		"OidbSvcTrpcTcp.0x8a1_7", // request group
		"OidbSvcTrpcTcp.0x89a_0",
		"OidbSvcTrpcTcp.0x89a_15",
		"OidbSvcTrpcTcp.0x88d_0", // fetch group detail
		"OidbSvcTrpcTcp.0x88d_14",
		"OidbSvcTrpcTcp.0x112a_1",
		"OidbSvcTrpcTcp.0x587_74",
		"OidbSvcTrpcTcp.0x1100_1",
		"OidbSvcTrpcTcp.0x1102_1",
		"OidbSvcTrpcTcp.0x1103_1",
		"OidbSvcTrpcTcp.0x1107_1",
		"OidbSvcTrpcTcp.0x1105_1",
		"OidbSvcTrpcTcp.0xf88_1",
		"OidbSvcTrpcTcp.0xf89_1",
		"OidbSvcTrpcTcp.0xf57_1",
		"OidbSvcTrpcTcp.0xf57_106",
		"OidbSvcTrpcTcp.0xf57_9",
		"OidbSvcTrpcTcp.0xf55_1",
		"OidbSvcTrpcTcp.0xf67_1",
		"OidbSvcTrpcTcp.0xf67_5",
		"OidbSvcTrpcTcp.0x6d9_4",
		//"OidbSvcTrpcTcp.0xb77_9",
		// sisi
		"QQLBSShareSvc.room_operation",
		"QQAIOMediaSvc.share_trans_check",
		"OidbSvcTrpcTcp.0xdc2_34",
		"OidbSvcTrpcTcp.0x929b_0",
		"Heartbeat.Alive",
		// extends
		"OidbSvcTrpcTcp.0x12a9_100",
		"OidbSvcTrpcTcp.0xb77_9",
		"OidbSvcTrpcTcp.0xf51_1",
		"OidbSvcTrpcTcp.0xfe1_2",
		"OidbSvcTrpcTcp.0xfe1_8",
		// king add
		"trpc.msg.register_proxy.RegisterProxy.SsoInfoSync",
		"trpc.qq_new_tech.status_svc.StatusService.SsoHeartBeat",
		"trpc.qq_new_tech.status_svc.StatusService.SetStatus",
	} {
		sign_map[cmd] = 1
	}
}
