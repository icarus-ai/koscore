package system_type

import "github.com/kernel-ai/koscore/client/packets/structs/sso_type"

var HeartbeatReq = []byte{0x00, 0x00, 0x00, 0x04}

var (
	AttributeHeartbeat    = sso_type.NewServiceAttribute("Heartbeat.Alive", sso_type.RequestSimple, sso_type.NoEncrypt, true)
	AttributeSsoHeartBeat = sso_type.NewServiceAttributeD2D2("trpc.qq_new_tech.status_svc.StatusService.SsoHeartBeat")
)
