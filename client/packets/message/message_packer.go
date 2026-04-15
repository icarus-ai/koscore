package message

import (
	"github.com/kernel-ai/koscore/client/packets/message/message_type"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildRawMessage(cli_seq, random uint32, route *message.SendRoutingHead, body *message.MessageBody) (*sso_type.SsoPacket, error) {
	/*
		// grp_id not null
		if (route.Grp != nil && route.Grp.GroupCode.IsSome()) || (route.GrpTmp != nil && route.GrpTmp.GroupUin.IsSome()) {
			msg.Ctrl = &pb_msg.MessageControl{MsgFlag: int32(utils.TimeStamp())}
		}
	*/
	data, e := proto.Marshal(&message.PbSendMsgReq{
		RoutingHead: route,
		ContentHead: &message.SendContentHead{
			PkgNum:    proto.Some[uint32](1),
			PkgIndex:  proto.Some[uint32](0),
			DivSeq:    proto.Some[uint32](0),
			AutoReply: proto.Some[uint32](0),
		},
		MessageBody:    body,
		ClientSequence: proto.Some(uint64(cli_seq)),
		Random:         proto.Some(random),
	})
	if e != nil {
		return nil, e
	}
	return message_type.AttributePbSendMsg.NewSsoPacket(cli_seq, data), nil
}

func ParseMessagePacket(data []byte) (*message.PbSendMsgResp, error) {
	return proto.Unmarshal[message.PbSendMsgResp](data)
}
