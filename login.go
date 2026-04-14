package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/kernel-ai/koscore/client"
	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/login/login_type"
	"github.com/kernel-ai/koscore/client/sign"
	"github.com/kernel-ai/koscore/utils/comm"
)

const (
	k_SIGN_V2_URI = "https://sign.lagrangecore.org/api/sign/sec-sign"
	k_SIGN_V2_KEY = "you sign key"
	k_SIGN_V2_UIN = 0 // you sign uin
	TOKEN_PATH    = "session_bin"
	QRCODE_PATH   = "qrcode.png"
	k_GUID        = "you guid hex"

	dump_path = "dump"
)

func main() {
	qq_login(k_SIGN_V2_UIN, "")
}

func qq_login(uin uint64, password string) {
	ctx := new_client(uin, password)
	d, e := os.ReadFile(TOKEN_PATH)
	if e != nil {
		comm.FAIL("load session: open: %v", e)
	}
	session, e := auth.UnmarshalSigInfo(d, true)
	if e != nil {
		comm.FAIL("load session: unmarshal: %v", e)
	}
	//session, _ := proto.Unmarshal[auth.Session](d)
	ctx.UseSig(session)
	manual_login(ctx)

	if e = ctx.RefreshFriendCache(); e == nil {
		infos := ctx.GetCachedAllFriendsInfo()
		comm.LOGI("好友数: %d", len(infos))
	} else {
		comm.LOGW("刷新好友列表失败: %v", e)
	}

	if us, e := ctx.GetUnidirectionalFriendList(); e == nil {
		comm.LOGI("单向好友数: %d", len(us))
	} else {
		comm.LOGW("刷新单向好友列表失败: %v", e)
	}

	if e = ctx.RefreshAllGroupMembersCache(); e == nil {
		infos := ctx.GetCachedAllGroupsInfo()
		comm.LOGI("群数: %d", len(infos))
	} else {
		comm.LOGW("刷新群列表&成员失败: %v", e)
	}

	comm.LOGD("wait signal kill")
	comm.WaitSignalKill()
}

func manual_login(ctx *client.QQClient) login_type.LoginState {
	if e := ctx.FastLogin(); e == nil {
		return loginResponseProcessor(ctx, login_type.LoginSuccess)
	}
	state, unusual, e := ctx.ExchangeEasyLogin()
	if e != nil {
		comm.FAIL("%v", e)
	}
	if state == login_type.LoginSuccess {
		return loginResponseProcessor(ctx, login_type.LoginSuccess)
	}
	return loginResponseProcessor(ctx, qrcode_login(ctx, unusual))
}

func qrcode_login(m *client.QQClient, unusual_sigs []byte) login_type.LoginState {
	// if m.GetVersion().OS.IsAndroid() && len(password) == 0 { comm.FAIL("Android Platform can not use QRLogin, Please fill in password") }

	unusual := len(unusual_sigs) > 0
	im, e := m.FetchQRode(3, unusual_sigs)
	if e != nil {
		comm.FAIL("%v", e)
	}
	session := m.Session()
	comm.LOGD("qrcode url: %s", im.Url)
	comm.LOGD("debug m.State.QrSig: %x", session.State.QrSig)
	_ = os.WriteFile(QRCODE_PATH, im.Image, 0o644)

	comm.LOGD("GetQrcodeResult")
	time.Sleep(time.Second)

	state, e := m.GetRCodeResult()
	if e != nil {
		comm.FAIL("%v", e)
	}
	prevState := state
	for {
		time.Sleep(time.Second)
		state, e = m.GetRCodeResult()
		if e != nil {
			comm.LOGW("%v", e)
			continue
		}
		if prevState == state {
			continue
		}
		prevState = state
		switch state {
		case login_type.TransEmpWaitingForScan:
			comm.LOGI("扫码...")
		case login_type.TransEmpWaitingForConfirm:
			comm.LOGI("扫码成功, 请在手机端确认登录.")
		case login_type.TransEmpCanceled:
			comm.FAIL("扫码被用户取消.")
		case login_type.TransEmpCodeExpired:
			comm.FAIL("二维码过期")
		case login_type.TransEmpConfirmed:
			comm.LOGI("扫码成功: 手机端已确认登录")
			if unusual {
				if e := m.UnusualEasyLogin(); e != nil {
					comm.FAIL("扫码: unusual easy login: %v", e)
				}
			} else if e := m.QRCodeLogin(); e != nil {
				comm.FAIL("扫码: login %v", e)
			}
			return login_type.LoginSuccess
		default:
			comm.LOGW("扫码 code: %v", state.String())
		}
	}
}

func loginResponseProcessor(m *client.QQClient, stata login_type.LoginState) login_type.LoginState {
	if stata == login_type.LoginSuccess {
		if !m.Online.Load() {
			comm.FAIL("no online")
		}
		if e := os.WriteFile(TOKEN_PATH, m.Session().Marshal(), 0o644); e != nil {
			comm.LOGW("m.session.Save: %v", e)
		}
		comm.LOGD("online")
	}
	return stata
}

const (
	k_APP_KERNEL = "amd64"
	k_VERSION    = "46494"
)

func new_client(uin uint64, password string) *client.QQClient {
	config_init()
	guid, _ := hex.DecodeString(k_GUID)
	version := auth.AppList[auth.LINUX][k_VERSION]
	device := &auth.DeviceInfo{
		GUID:          guid,
		DeviceName:    "koscore-20260331",
		SystemKernel:  "linux" + k_APP_KERNEL,
		KernelVersion: k_APP_KERNEL,
	}
	ctx := client.NewClient(uin, password)
	ctx.SetVersion(version)
	ctx.SetDevice(device)
	ctx.SetSignProvider(sign.NewSignerV2(uint32(uin), version, device, []string{k_SIGN_V2_URI, k_SIGN_V2_KEY}))
	ctx.SetLogger(&logfmt{})
	return ctx
}

func config_init() {
	fn := func(version string, app_client_version, appid, sub_appid, misc_bitMap uint32) *auth.AppInfo {
		return &auth.AppInfo{
			OS:               auth.LINUX,
			Kernel:           auth.LINUX,
			VendorOS:         "linux",
			PtVersion:        "2.0.0",
			SsoVersion:       19,
			PackageName:      "com.tencent.qq",
			ApkSignatureMd5:  []byte("com.tencent.qq"),
			CurrentVersion:   fmt.Sprintf("%s-%d", version, app_client_version),
			QUA:              fmt.Sprintf("V1_LNX_NQ_%s_%d_GW_B", version, app_client_version),
			AppId:            appid,
			SubAppId:         sub_appid,
			AppClientVersion: app_client_version,
			SdkInfo: auth.WtLoginSdkInfo{
				SdkBuildTime: 0,
				SdkVersion:   "nt.wtlogin.0.0.1",
				MiscBitMap:   misc_bitMap,
				SubSigMap:    0,
				MainSigMap:   169742560,
				//wtlogin.WLOGIN_ST_WEB | wtlogin.WLOGIN_A2 | wtlogin.WLOGIN_ST |
				//wtlogin.WLOGIN_S_KEY | wtlogin.WLOGIN_V_KEY  | wtlogin.WLOGIN_D2  |
				//wtlogin.WLOGIN_SId   | wtlogin.WLOGIN_PS_KEY | wtlogin.WLOGIN_DA2 | wtlogin.WLOGIN_PT4_TOKEN,
			},
		}
	}

	auth.AppList = auth.APP_INFO_MAP{
		auth.LINUX: map[string]*auth.AppInfo{
			"46494": fn("3.2.26", 46494, 1600001615, 537345891, 32764),
		},
	}
}

type logfmt struct{}

func (p logfmt) mytag(tag, format string) string {
	return "[" + time.Now().UTC().Format("20060102_150405") + "][" + tag + "]: " + format + "\n"
}
func (p logfmt) LOGI(format string, arg ...any) { fmt.Printf(p.mytag("I", format), arg...) }
func (p logfmt) LOGD(format string, arg ...any) { fmt.Printf(p.mytag("D", format), arg...) }
func (p logfmt) LOGW(format string, arg ...any) { fmt.Printf(p.mytag("W", format), arg...) }
func (p logfmt) LOGE(format string, arg ...any) { fmt.Printf(p.mytag("E", format), arg...) }
func (p logfmt) DUMP(data []byte, format string, arg ...any) {
	msg := fmt.Sprintf(format, arg...)
	if _, e := os.Stat(dump_path); e != nil {
		if e = os.MkdirAll(dump_path, 0o755); e != nil {
			p.LOGE("出现错误 %v. 详细信息转储失败", msg)
			return
		}
	}
	fs := fmt.Sprintf("%s/%s.log", dump_path, time.Now().Format("20060102_150405"))
	if e := os.WriteFile(fs, data, 0o644); e != nil {
		p.LOGE("出现错误 %v. 详细信息转储失败", msg)
		return
	}
	p.LOGE("出现错误 %v. 详细信息已转储至文件 %v 请连同日志提交给开发者处理", msg, fs)
}
