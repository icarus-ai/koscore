package auth

import (
	"strings"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/login"
	"github.com/kernel-ai/koscore/utils/types"
)

var AppList APP_INFO_MAP

type APP_INFO_MAP map[SYS_OS]map[string]*AppInfo

type WtLoginSdkInfo struct {
	SdkBuildTime uint32 `json:"sdk_build_time"`
	SdkVersion   string `json:"sdk_version"`
	MiscBitMap   uint32 `json:"misc_bit_map"`
	SubSigMap    uint32 `json:"sub_sig_map"`
	MainSigMap   Sig    `json:"main_sig_map"`
}

type AppInfo struct {
	OS       SYS_OS `json:"os"`
	Kernel   string `json:"kernel"`
	VendorOS string `json:"vendor_os"`

	QUA              string         `json:"qua"`
	CurrentVersion   string         `json:"current_version"`
	PtVersion        string         `json:"pt_version"`
	SsoVersion       uint32         `json:"pt_os_version"`
	PackageName      string         `json:"package_name"`
	ApkSignatureMd5  types.Bytes    `json:"package_sign"`
	SdkInfo          WtLoginSdkInfo `json:"sdk_info"`
	AppId            uint32         `json:"app_id"`
	SubAppId         uint32         `json:"sub_app_id"`
	AppClientVersion uint32         `json:"app_client_version"`
}

type SYS_OS string

const (
	LINUX   = "Linux"
	WIN     = "Winwows"
	MAC     = "Mac"
	Android = "Android" // Android and AWatch
	APad    = "ANDROID"
)

func (m SYS_OS) String() string       { return string(m) }
func (m SYS_OS) ProtocolName() string { return strings.ToLower(string(m)) }
func (m SYS_OS) ProtocolCode() login.NTLoginPlatform {
	switch m {
	case LINUX:
		return login.NTLoginPlatform_LINUX
	case WIN:
		return login.NTLoginPlatform_WINDOWS
	case MAC:
		return login.NTLoginPlatform_MAC
	case Android, APad:
		return login.NTLoginPlatform_ANDROID
	default:
		return login.NTLoginPlatform_UNKNOWN
	}
}
func (m SYS_OS) IsPC() bool      { return m == LINUX || m == WIN || m == MAC }
func (m SYS_OS) IsAndroid() bool { return m == Android || m == APad }

type Sig uint32

const (
	WLOGIN_A5        = 1 << 1
	WLOGIN_RESERVED  = 1 << 4
	WLOGIN_ST_WEB    = 1 << 5
	WLOGIN_A2        = 1 << 6
	WLOGIN_ST        = 1 << 7
	WLOGIN_LS_KEY    = 1 << 9
	WLOGIN_S_KEY     = 1 << 12
	WLOGIN_SIG64     = 1 << 13
	WLOGIN_OPEN_KEY  = 1 << 14
	WLOGIN_TOKEN     = 1 << 15
	WLOGIN_V_KEY     = 1 << 17
	WLOGIN_D2        = 1 << 18
	WLOGIN_SId       = 1 << 19
	WLOGIN_PS_KEY    = 1 << 20
	WLOGIN_AQ_SIG    = 1 << 21
	WLOGIN_LH_SIG    = 1 << 22
	WLOGIN_PAY_TOKEN = 1 << 23
	WLOGIN_PF        = 1 << 24
	WLOGIN_DA2       = 1 << 25
	WLOGIN_QR_PUSH   = 1 << 26
	WLOGIN_PT4_TOKEN = 1 << 27
)
