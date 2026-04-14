package ntlogin_type

import (
	"fmt"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/login"
)

func NTLoginRetCodeString(m login.NTLoginRetCode) string {
	switch m {
	case login.NTLoginRetCode_ERROR_DEFAULT:
		return fmt.Sprintf("DEFAULT (%d)", m)
	case login.NTLoginRetCode_ERROR_INVALID_PARAMETER:
		return fmt.Sprintf("INVALID_PARAMETER (%d)", m)
	case login.NTLoginRetCode_ERROR_SYSTEM_FAILED:
		return fmt.Sprintf("SYSTEM_FAILED (%d)", m)
	case login.NTLoginRetCode_ERROR_TIMEOUT_RETRY:
		return fmt.Sprintf("TIMEOUT_RETRY (%d)", m)
	case login.NTLoginRetCode_ERROR_NEED_UPDATE:
		return fmt.Sprintf("NEED_UPDATE (%d)", m)
	case login.NTLoginRetCode_ERROR_FROZEN:
		return fmt.Sprintf("FROZEN (%d)", m)
	case login.NTLoginRetCode_ERROR_PROTECT:
		return fmt.Sprintf("PROTECT (%d)", m)
	case login.NTLoginRetCode_ERROR_STRICT:
		return fmt.Sprintf("STRICT (%d)", m)
	case login.NTLoginRetCode_ERROR_PROOF_WATER:
		return fmt.Sprintf("PROOF_WATER (%d)", m)
	case login.NTLoginRetCode_ERROR_REFUSE_PASSWORD_LOGIN:
		return fmt.Sprintf("REFUSE_PASSWORD_LOGIN (%d)", m)
	case login.NTLoginRetCode_ERROR_NEW_DEVICE:
		return fmt.Sprintf("NEW_DEVICE (%d)", m)
	case login.NTLoginRetCode_ERROR_UNUSUAL_DEVICE:
		return fmt.Sprintf("UNUSUAL_DEVICE (%d)", m)
	case login.NTLoginRetCode_ERROR_INVALID_COOKIE:
		return fmt.Sprintf("INVALID_COOKIE (%d)", m)
	case login.NTLoginRetCode_ERROR_ACCOUNT_OR_PASSWORD_ERROR:
		return fmt.Sprintf("ACCOUNT_OR_PASSWORD_ERROR (%d)", m)
	case login.NTLoginRetCode_ERROR_EXPIRE_TICKET:
		return fmt.Sprintf("EXPIRE_TICKET (%d)", m)
	case login.NTLoginRetCode_ERROR_KICKED_TICKET:
		return fmt.Sprintf("KICKED_TICKET (%d)", m)
	case login.NTLoginRetCode_ERROR_ILLEGAL_TICKET:
		return fmt.Sprintf("ILLEGAL_TICKET (%d)", m)
	case login.NTLoginRetCode_ERROR_SEC_BEAT:
		return fmt.Sprintf("SEC_BEAT (%d)", m)
	case login.NTLoginRetCode_ERROR_ACCOUNT_NOT_UIN:
		return fmt.Sprintf("ACCOUNT_NOT_UIN (%d)", m)
	case login.NTLoginRetCode_ERROR_NEED_VERIFY_REAL_NAME:
		return fmt.Sprintf("NEED_VERIFY_REAL_NAME (%d)", m)
	case login.NTLoginRetCode_ERROR_NICE_ACCOUNT_EXPIRED:
		return fmt.Sprintf("NICE_ACCOUNT_EXPIRED (%d)", m)
	case login.NTLoginRetCode_ERROR_BLACK_ACCOUNT:
		return fmt.Sprintf("BLACK_ACCOUNT (%d)", m)
	case login.NTLoginRetCode_ERROR_TOO_OFTEN:
		return fmt.Sprintf("TOO_OFTEN (%d)", m)
	case login.NTLoginRetCode_ERROR_TOO_MANY_TIMES_TODAY:
		return fmt.Sprintf("TOO_MANY_TIMES_TODAY (%d)", m)
	case login.NTLoginRetCode_ERROR_UNREGISTERED:
		return fmt.Sprintf("UNREGISTERED (%d)", m)
	case login.NTLoginRetCode_ERROR_NICE_ACCOUNT_PARENT_CHILD_EXPIRED:
		return fmt.Sprintf("NICE_ACCOUNT_PARENT_CHILD_EXPIRED (%d)", m)
	case login.NTLoginRetCode_ERROR_SMS_INVALID:
		return fmt.Sprintf("SMS_INVALID (%d)", m)
	case login.NTLoginRetCode_ERROR_TGTGT_EXCHANGE_A1_FORBID:
		return fmt.Sprintf("TGTGT_EXCHANGE_A1_FORBID (%d)", m)
	case login.NTLoginRetCode_ERROR_REMIND_CANCELLED_STATUS:
		return fmt.Sprintf("REMIND_CANCELLED_STATUS (%d)", m)
	case login.NTLoginRetCode_ERROR_MULTIPLE_PASSWORD_INCORRECT:
		return fmt.Sprintf("MULTIPLE_PASSWORD_INCORRECT (%d)", m)
	default:
		return fmt.Sprintf("UNKNOWN_ERROR (%d)", m)
	}
}

// NewDeviceVerify || CaptchaVerify || UnusualVerify
func NeedVerify(m login.NTLoginRetCode) bool {
	return m == login.NTLoginRetCode_ERROR_NEW_DEVICE || m == login.NTLoginRetCode_ERROR_PROOF_WATER || m == login.NTLoginRetCode_ERROR_UNUSUAL_DEVICE
}
