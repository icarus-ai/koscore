package client

import (
	"fmt"
	"strings"

	"github.com/kernel-ai/koscore/utils/exception"
	"github.com/pkg/errors"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/common"
	"github.com/kernel-ai/koscore/client/packets/structs"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/client/sign"
	"github.com/kernel-ai/koscore/utils/proto"
)

func (m *PacketContext) uniPacket(packet *sso_type.SsoPacket) (seq uint32, d []byte, e error) {
	var info *common.SsoSecureInfo = nil
	var val *sign.Value = nil
	if sign.ContainSignPKG(packet.Command) {
		val, e = m.sig_context.Sign(packet.Command, packet.Sequence, packet.Data)
		if e != nil {
			return
		}
		info = &common.SsoSecureInfo{SecSign: val.Sign, SecToken: val.Token, SecExtra: val.Extra}
	}
	d, e = structs.BuildSsoPacket(m.version, m.device, m.session, packet, info)
	seq = packet.Sequence
	return
}

func (m *QQClient) sendUniPacketAndWait(cmd string, buf []byte) (*sso_type.SsoPacket, error) {
	return m.sso_context.SendPacketAndWait(sso_type.NewServiceAttributeD2D2(cmd).NewSsoPacket(m.session.GetAndIncreaseSequence(), buf))
}
func (m *QQClient) sendOidbPacketAndWait(pkt *sso_type.SsoPacket) (*sso_type.SsoPacket, error) {
	if pkt.Sequence == 0 {
		pkt.Sequence = m.session.GetAndIncreaseSequence()
	}
	return m.sso_context.SendPacketAndWait(pkt)
}

// android ??
func (m *QQClient) webSsoRequest(host, webcmd, data string) (string, error) {
	sub, s := "", strings.Split(host, `.`)
	for i := len(s) - 1; i >= 0; i-- {
		sub += s[i]
		if i != 0 {
			sub += "_"
		}
	}
	req, _ := proto.Marshal(&common.WebSsoRequestBody{
		Type: proto.Some[uint32](0),
		Data: proto.Some(data),
	})
	sso, e := m.sendUniPacketAndWait(fmt.Sprintf("MQUpdateSvc_%s.web.%s", sub, webcmd), req)
	if e != nil {
		return "", errors.Wrap(e, "send web sso request error")
	}
	rsp, e := proto.Unmarshal[common.WebSsoResponseBody](sso.Data)
	if e != nil {
		return "", exception.NewUnmarshalProtoException(e, "response")
	}
	return rsp.Data.Unwrap(), nil
}
