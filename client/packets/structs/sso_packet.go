package structs

import (
	"fmt"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/common"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

func BuildSsoPacket(version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session, packet *sso_type.SsoPacket, info *common.SsoSecureInfo) (d []byte, e error) {
	switch packet.RequestType {
	case sso_type.RequestD2Auth:
		return buildServicePackerProtocol12(session, buildSsoPackerProtocol12(version, device, session, packet, info), packet.ServiceAttribute)
	case sso_type.RequestSimple:
		return buildServicePackerProtocol13(session, packet, buildSsoPackerProtocol13(version, session, packet, info))
	default:
		return nil, fmt.Errorf("invalid operation exception: unknown request type: %d", packet.RequestType)
	}
}

func ParseSsoPacket(session *auth.Session, data []byte) (*sso_type.SsoPacket, error) {
	return parseSsoPacker(parseServicePacker(session, data))
}
