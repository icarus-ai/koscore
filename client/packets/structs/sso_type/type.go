package sso_type

/*
type SsoPacketProvider interface {
	BuildSsoPacket(version *auth.AppInfo, device *auth.DeviceInfo, session *auth.Session, packet *SsoPacket, options *ServiceAttribute, info *common.SsoSecureInfo) ([]byte, error)
	ParseSsoPacket(session *auth.Session, data []byte) (*SsoPacket, error)
}
*/

type SsoPacket struct {
	*ServiceAttribute
	Sequence uint32
	Data     []byte
	Extra    string
	RetCode  int32
}

type EncryptType uint8

const (
	NoEncrypt    EncryptType = 0x00
	EncryptD2Key EncryptType = 0x01
	EncryptEmpty EncryptType = 0x02
)

type RequestType uint8

const (
	RequestD2Auth RequestType = 0x0C
	RequestSimple RequestType = 0x0D
)

type ServiceAttribute struct {
	Command     string
	RequestType RequestType
	EncryptType EncryptType
	DisableLog  bool
}

func NewServiceAttributeD2Empty(command string, disablelog ...bool) *ServiceAttribute {
	return NewServiceAttribute(command, RequestD2Auth, EncryptEmpty, disablelog...)
}
func NewServiceAttributeD2D2(command string, disablelog ...bool) *ServiceAttribute {
	return NewServiceAttribute(command, RequestD2Auth, EncryptD2Key, disablelog...)
}
func NewServiceAttribute(command string, requestType RequestType, encryptType EncryptType, disablelog ...bool) *ServiceAttribute {
	return &ServiceAttribute{
		Command:     command,
		RequestType: requestType, // common.D2Auth
		EncryptType: encryptType, // common.EncryptD2Key
		DisableLog:  len(disablelog) > 0 && disablelog[0],
	}
}

func (m *ServiceAttribute) NewSsoPacket(seq uint32, data []byte) *SsoPacket {
	return &SsoPacket{
		ServiceAttribute: m,
		Sequence:         seq,
		Data:             data,
	}
}
