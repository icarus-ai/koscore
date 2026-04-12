package message

import (
	"github.com/kernel-ai/koscore/client/packets/message/message_type"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildGroupRecallMessagePacket(gin, seq uint64) *sso_type.SsoPacket {
	data, _ := proto.Marshal(&message.SsoGroupRecallMsgReq{
		Type:     proto.Some[uint32](1),
		GroupUin: proto.Some(int64(gin)),
		Field3:   &message.SsoGroupRecallMsgReqField3{Sequence: proto.Some(seq)},
		Field4:   &message.SsoGroupRecallMsgReqField4{Field1: proto.Some[uint32](0)},
	})
	return message_type.AttributeSsoGroupRecallMsg.NewSsoPacket(0, data)
}

func BuildC2CRecallMessagePacket(target_uid string, seq, random, client_seq uint64, timestamp uint32) *sso_type.SsoPacket {
	data, _ := proto.Marshal(&message.SsoC2CRecallMsgReq{
		Type:      proto.Some[uint32](1),
		TargetUid: proto.Some(target_uid),
		Info: &message.SsoC2CRecallMsgReqInfo{
			Sequence:       proto.Some(seq),
			Random:         proto.Some(uint32(random)),
			MessageId:      proto.Some(0x01000000<<32 | random),
			Timestamp:      proto.Some(timestamp),
			Field5:         proto.Some[uint32](0),
			ClientSequence: proto.Some(client_seq),
		},
		Settings: &message.SsoC2CRecallMsgReqSettings{
			Field1: proto.FALSE,
			Field2: proto.FALSE,
		},
		Field6: proto.FALSE,
	})
	return message_type.AttributeSsoC2CRecallMsg.NewSsoPacket(0, data)
}
