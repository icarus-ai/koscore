package event

type (
	DecodersCall  func(data []byte) bool
	DecodersEvent map[string]DecodersCall
)

const (
	DECODE_CMD_0210_015D            = "0210_015D"
	DECODE_CMD_SVC_PUSHREQ          = "ConfigPushSvc.PushReq"
	DECODE_CMD_SVC_MSF_LOGIN_NOTIFY = "StatSvc.SvcReqMSFLoginNotify"
)
