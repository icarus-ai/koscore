package login_type

import (
	"fmt"

	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/types"
)

type LoginReq struct {
	Cmd      LoginCommand
	Password string
	Ticket   string
	Code     string
}

type LoginRsp struct {
	RetCode byte
	State   LoginState
	Tlvs    types.Tlvs
	Err     string
}

// protocol pc or android
var AttributeLogin = sso_type.NewServiceAttributeD2Empty("wtlogin.login")

type LoginCommand uint8

const (
	LoginTgtgt         LoginCommand = 0x09
	LoginCaptcha       LoginCommand = 0x02
	LoginFetchSMSCode  LoginCommand = 0x08
	LoginSubmitSMSCode LoginCommand = 0x07
)

// login_stste
type LoginState uint8

const (
	LoginSuccess                 LoginState = 0
	LoginCaptchaVerify           LoginState = 2
	LoginSmsRequired             LoginState = 160
	LoginDeviceLock              LoginState = 204
	LoginDeviceLockViaSmsNewArea LoginState = 239

	LoginPreventByIncorrectPassword     LoginState = 1
	LoginPreventByReceiveIssue          LoginState = 3
	LoginPreventByTokenExpired          LoginState = 15
	LoginPreventByAccountBanned         LoginState = 40
	LoginPreventByOperationTimeout      LoginState = 155
	LoginPreventBySmsSentFailed         LoginState = 162
	LoginPreventByIncorrectSmsCode      LoginState = 163
	LoginPreventByLoginDenied           LoginState = 167
	LoginPreventByOutdatedVersion       LoginState = 235
	LoginPreventByHighRiskOfEnvironment LoginState = 237

	LoginUnknown LoginState = 240
)

func (m LoginState) String() string {
	switch m {
	case LoginSuccess:
		return "Success"
	case LoginCaptchaVerify:
		return fmt.Sprintf("验证码验证(%d)", m)
	case LoginSmsRequired:
		return fmt.Sprintf("需要短信(%d)", m)
	case LoginDeviceLock:
		return fmt.Sprintf("设备锁(%d)", m)
	case LoginDeviceLockViaSmsNewArea:
		return fmt.Sprintf("设备锁.短信(%d)", m)
	case LoginPreventByIncorrectPassword:
		return fmt.Sprintf("密码错误(%d)", m)
	case LoginPreventByReceiveIssue:
		return fmt.Sprintf("PreventByReceiveIssue(%d)", m)
	case LoginPreventByTokenExpired:
		return fmt.Sprintf("令牌已过期(%d)", m)
	case LoginPreventByAccountBanned:
		return fmt.Sprintf("账户已被禁止(%d)", m)
	case LoginPreventByOperationTimeout:
		return fmt.Sprintf("操作超时(%d)", m)
	case LoginPreventBySmsSentFailed:
		return fmt.Sprintf("短信发送失败(%d)", m)
	case LoginPreventByIncorrectSmsCode:
		return fmt.Sprintf("短信验证码错误(%d)", m)
	case LoginPreventByLoginDenied:
		return fmt.Sprintf("登录被拒(%d)", m)
	case LoginPreventByOutdatedVersion:
		return fmt.Sprintf("版本过时(%d)", m)
	case LoginPreventByHighRiskOfEnvironment:
		return fmt.Sprintf("环境高风险(%d)", m)
	case LoginUnknown:
		return fmt.Sprintf("Unknown(%d)", m)
	default:
		return fmt.Sprintf("state unknown(%d)", m)
	}
}

/*
	case   6: text = "服务连接中.. (一定程度锁定了账号, 例如, 新账号被要求先在手机登录)"
	case  45: text = "被限制只能在最新版手机客户端登录(配置 SignServer 后重试)"
	case 235: text = "当前QQ版本过低(设备信息被封禁), 删除device.kson后重试"
	case 237: text = "当前网络不稳定, 登录过于频繁, 请在手机QQ登录并根据提示完成认证后等一段时间重试"
	case 238: text = "禁止密码登录, 强制要求扫码或者短信"
*/
