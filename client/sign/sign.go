package sign

import (
	"errors"
	"math"
	// "github.com/kernel-ai/koscore/utils/comm"
)

const server_latency_down = math.MaxUint32

var (
	ErrVersionMismatch    = errors.New("sign version mismatch")
	ErrAllSignServiceDown = errors.New("all sign service down")
	k_err_rsp_sign_nil    = errors.New("rsp.sign nil")
)

/*
var sign_map map[string]uint8 // 只在启动时初始化, 无并发问题

func ContainSignPKG(cmd string) bool {
	_,  ok := sign_map[cmd]
	if !ok {
		if strings.Contains(cmd, "OidbSvcTrpcTcp.0x") {
			//sign_map[cmd] = cmd
			return true
		}
		//comm.LOGD("unsign cmd: %s", cmd)
	}
	return ok
}

func AddSignPKG(pkg string) { sign_map[pkg] = 1 }

func init() {
	sign_map = make(map[string]uint8)

	for _, cmd := range []string {
		// see github.com/LagrangeDev/kosa/blob/dd029d8f168ab68248ef4c511c46f6b0ae418652/kosa/src/common/sign.rs

		"trpc.o3.ecdh_access.EcdhAccess.SsoEstablishShareKey",
		"trpc.o3.ecdh_access.EcdhAccess.SsoSecureAccess",
		"trpc.o3.report.Report.SsoReport",
			"trpc.o3.ecdh_access.EcdhAccess.SsoSecureA2Access",
			"trpc.o3.ecdh_access.EcdhAccess.SsoSecureA2Establish",

		"trpc.login.ecdh.EcdhService.SsoKeyExchange",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginNewDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLoginUnusualDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginUnusualDevice",
			"trpc.login.ecdh.EcdhService.SsoNTLoginAuthLogin",
			"trpc.login.ecdh.EcdhService.SsoNTLoginAuthCodeLogin",
			"trpc.login.ecdh.EcdhService.SsoQRLoginGenQr",
			"trpc.login.ecdh.EcdhService.SsoNTLoginTGTExchangeFastLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginRefreshTicket",
		"trpc.login.ecdh.EcdhService.SsoNTLoginRefreshA2",                  //

		"wtlogin.trans_emp",
		"wtlogin.login",
			"wtlogin.exchange_emp",
			"wtlogin_device.login",
			"wtlogin_device.tran_sim_emp",

		"trpc.group.long_msg_interface.MsgService.SsoRecvLongMsg",
		"trpc.group.long_msg_interface.MsgService.SsoSendLongMsg",
			"trpc.group_pro.msgproxy.sendmsg",
		"trpc.msg.msg_svc.MsgService.SsoReadedReport",
		"trpc.msg.msg_svc.MsgService.SsoC2CRecallMsg",
		"trpc.msg.register_proxy.RegisterProxy.SsoInfoSync",
		"trpc.msg.register_proxy.RegisterProxy.SsoGetGroupMsg",
		"trpc.msg.register_proxy.RegisterProxy.SsoGetRoamMsg",
		"trpc.msg.register_proxy.RegisterProxy.SsoGetC2cMsg",
		"trpc.qq_new_tech.status_svc.StatusService.SsoHeartBeat",
		"trpc.qq_new_tech.status_svc.StatusService.SetStatus",

			"trpc.ecom.api_gateway.ApiGateway.SsoForward",
			"trpc.passwd.manager.PasswdManager.SetPasswd",
			"trpc.passwd.manager.PasswdManager.VerifyPasswd",
			"trpc.qqhb.qqhb_proxy.Handler.sso_handle",

			"QQConnectLogin.auth",
			"QQConnectLogin.pre_auth",

			"ConnAuthSvr.fast_qq_login",
			"ConnAuthSvr.sdk_auth_api",
			"ConnAuthSvr.sdk_auth_api_emp",

		"MessageSvc.PbSendMsg",
			"MsgProxy.SendMsg",

		"OidbSvcTrpcTcp.0x587_74",
		"OidbSvcTrpcTcp.0x6d9_4",
		"OidbSvcTrpcTcp.0x758_1", // create group
		"OidbSvcTrpcTcp.0x7c1_1",
		"OidbSvcTrpcTcp.0x7c2_5", // request friend
		"OidbSvcTrpcTcp.0x8a1_7", // request group
		"OidbSvcTrpcTcp.0x89a_0",
		"OidbSvcTrpcTcp.0x89a_15",
		"OidbSvcTrpcTcp.0x88d_0", // fetch group detail
		"OidbSvcTrpcTcp.0x88d_14",
		"OidbSvcTrpcTcp.0xf88_1",
		"OidbSvcTrpcTcp.0xf89_1",
		"OidbSvcTrpcTcp.0xf57_1",
		"OidbSvcTrpcTcp.0xf57_106",
		"OidbSvcTrpcTcp.0xf57_9",
		"OidbSvcTrpcTcp.0xf55_1",
		"OidbSvcTrpcTcp.0xf67_1",
		"OidbSvcTrpcTcp.0xf67_5",
		"OidbSvcTrpcTcp.0x10db_1",
		"OidbSvcTrpcTcp.0x1100_1",
		"OidbSvcTrpcTcp.0x1102_1",
		"OidbSvcTrpcTcp.0x1103_1",
		"OidbSvcTrpcTcp.0x1107_1",
		"OidbSvcTrpcTcp.0x1105_1",
		"OidbSvcTrpcTcp.0x112a_1",
		"OidbSvcTrpcTcp.0x11ec_1",
// sisi
		"QQLBSShareSvc.room_operation",
		"QQAIOMediaSvc.share_trans_check",
		"OidbSvcTrpcTcp.0xdc2_34",
		"OidbSvcTrpcTcp.0x929b_0",
			//"Heartbeat.Alive",

// extends
		"OidbSvcTrpcTcp.0xb77_9",
		"OidbSvcTrpcTcp.0xf51_1",
		"OidbSvcTrpcTcp.0xfe1_2",
		"OidbSvcTrpcTcp.0xfe1_8",
		"OidbSvcTrpcTcp.0x12a9_100",

// ???
		//"OidbSvc.0xcd5",
		//"OidbSvc.0xdc2_34",
		//"OidbSvc.0xb77_9",
		//"OidbSvcTcp.0x102a",
		//"OidbSvcTrpcTcp.0xcd5",

		"OidbSvcTrpcTcp.0xcd5_0",
		"OidbSvcTrpcTcp.0xdc2_58",
		"OidbSvcTrpcTcp.0xdc2_59",
		"OidbSvcTrpcTcp.0xf65_1",
		"OidbSvcTrpcTcp.0xf65_10",
		"OidbSvcTrpcTcp.0xf6e_1",
		"OidbSvcTrpcTcp.0xfa5_1",
		"OidbSvcTrpcTcp.0x101b_1",
		"OidbSvcTrpcTcp.0x101e_1",
		"OidbSvcTrpcTcp.0x101e_2",
		"OidbSvcTrpcTcp.0x102a_0",
		"OidbSvcTrpcTcp.0x102a_1",
		"OidbSvcTrpcTcp.0x10c8_1",
		"OidbSvcTrpcTcp.0x10c8_2",
		"OidbSvcTrpcTcp.0x112a_2",
		"OidbSvcTrpcTcp.0x112e_1",

		"OidbSvcTrpcTcp.0x917b_1",
		"OidbSvcTrpcTcp.0x93d7_1",
		"OidbSvcTrpcTcp.0x962a_1",
		"OidbSvcTrpcTcp.0x9409_7",
		"OidbSvcTrpcTcp.0x9409_10",
		"OidbSvcTrpcTcp.0x9409_11",
		"OidbSvcTrpcTcp.0x9409_12",
 		"OidbSvcTrpcTcp.0x9409_13",
 		"OidbSvcTrpcTcp.0x9409_14",
 		"OidbSvcTrpcTcp.0x9409_15",
 		"OidbSvcTrpcTcp.0x9409_16",
 		"OidbSvcTrpcTcp.0x9409_18",

	} { sign_map[cmd] = 1 }
}
*/
