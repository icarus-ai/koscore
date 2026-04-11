package message

import (
	"errors"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/message/message_type"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildMultiMsgUploadPacket(bot_uid string, gin uint64, msg []*message.CommonMessage, version *auth.AppInfo) *sso_type.SsoPacket {
	data, _ := proto.Marshal(&message.PbMultiMsgTransmit{
		Items: []*message.PbMultiMsgItem{{
			FileName: proto.Some("MultiMsg"),
			Buffer:   &message.PbMultiMsgNew{Msg: msg},
		}},
	})
	data, _ = proto.Marshal(&message.LongMsgInterfaceReq{
		SendReq: &message.LongMsgSendReq{
			MsgType:  proto.Some(utils.Ternary[uint32](gin == 0, 1, 3)), // 4 for wpamsg, 5 for grpmsg temp
			PeerInfo: &message.LongMsgPeerInfo{PeerUid: proto.Some(bot_uid)},
			GroupUin: proto.Some(int64(gin)),
			Payload:  binary.GZipCompress(data),
		},
		Attr: build_multi_msg_settings(utils.Ternary[uint32](version.OS.IsAndroid(), 3, 4), version),
	})
	return message_type.AttributeSsoSendLongMsg.NewSsoPacket(0, data)
}

func BuildMultiMsgDownloadPcket(bot_uid string, resid string, isgroup bool, version *auth.AppInfo) *sso_type.SsoPacket {
	data, _ := proto.Marshal(&message.LongMsgInterfaceReq{
		RecvReq: &message.LongMsgRecvReq{
			MsgType:  proto.Some(utils.Ternary[uint32](isgroup, 1, 3)), // 4 for wpamsg, 5 for grpmsg temp
			PeerInfo: &message.LongMsgPeerInfo{PeerUid: proto.Some(bot_uid)},
			ResId:    proto.Some(resid),
		},
		Attr: build_multi_msg_settings(utils.Ternary[uint32](version.OS.IsAndroid(), 3, 2), version),
	})
	return message_type.AttributeSsoRecvLongMsg.NewSsoPacket(0, data)
}

func build_multi_msg_settings(sub uint32, version *auth.AppInfo) *message.LongMsgAttr {
	return &message.LongMsgAttr{
		SubCmd: proto.Some(sub), // 1 -> Android 2 -> NTPC 0 -> Undefined
		ClientType: proto.Some(func() uint32 {
			switch version.OS {
			case auth.WIN, auth.MAC, auth.LINUX:
				return 1
			case auth.Android:
				return 2
			//case auth.IOS: return 3
			//case auth.IPad: return 4
			case auth.APad:
				return 5
			default:
				return 0
			}
		}()),
		Platform: proto.Some(func() uint32 {
			switch version.OS {
			case auth.WIN:
				return 3
			case auth.LINUX:
				return 6
			case auth.MAC:
				return 7
			case auth.Android, auth.APad:
				return 9
			default:
				return 0
			}
		}()),
		//ProxyType: proto.Some[uint32](0),
	}
}

func ParseMultiMsgUploadPacket(data []byte) (*message.LongMsgSendRsp, error) {
	rsp, e := proto.Unmarshal[message.LongMsgInterfaceRsp](data)
	if e != nil {
		return nil, e
	}
	if rsp.SendRsp == nil {
		return nil, errors.New("empty response data")
	}
	return rsp.SendRsp, nil
}

func ParseMultiMsgDownloadPacket(data []byte) ([]*message.CommonMessage, error) {
	rsp, e := proto.Unmarshal[message.LongMsgInterfaceRsp](data)
	if e != nil {
		return nil, e
	}
	if rsp.RecvRsp == nil || rsp.RecvRsp.Payload == nil {
		return nil, errors.New("empty response data")
	}

	trans, e := proto.Unmarshal[message.PbMultiMsgTransmit](binary.GZipUncompress(rsp.RecvRsp.Payload))
	if e != nil {
		return nil, e
	}

	for _, item := range trans.Items {
		if item.FileName.Unwrap() == "MultiMsg" {
			if len(item.Buffer.Msg) > 0 {
				return item.Buffer.Msg, nil
			}
		}
	}
	return nil, errors.New("empty response data")
}
