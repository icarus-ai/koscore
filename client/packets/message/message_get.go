package message

import (
	"fmt"

	"github.com/kernel-ai/koscore/client/packets/message/message_type"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildGetGroupMessagePacket(gin, start_seq, end_seq uint64) *sso_type.SsoPacket {
	data, _ := proto.Marshal(&message.SsoGetGroupMsg{
		Info: &message.SsoGetGroupMsgInfo{
			GroupUin:      proto.Some(int64(gin)),
			StartSequence: proto.Some(start_seq),
			EndSequence:   proto.Some(end_seq),
		},
		Filter: proto.Some[uint32](1), // 1 for no filter, 2 for filter of only 10 msg within 3 days
	})
	return message_type.AttributeSsoGetGroupMsg.NewSsoPacket(0, data)
}

func ParseGetGroupMessagePacket(data []byte) ([]*message.CommonMessage, error) {
	rsp, e := proto.Unmarshal[message.SsoGetGroupMsgRsp](data)
	if e != nil {
		return nil, e
	}
	if rsp.RetCode.Unwrap() != 0 {
		return nil, fmt.Errorf("%s (%d)", rsp.ErrorMsg.Unwrap(), rsp.RetCode.Unwrap())
	}
	return rsp.Body.Messages, nil
}

func BuildGetRoamMessagePacket(peer_uid string, timestamp, count uint32) *sso_type.SsoPacket {
	data, _ := proto.Marshal(&message.SsoGetRoamMsgReq{
		PeerUid:   proto.Some(peer_uid),
		Time:      proto.Some(timestamp),
		Random:    proto.Some[uint32](0),
		Count:     proto.Some(count),
		Direction: proto.Some[uint32](0),
	})
	return message_type.AttributeSsoGetRoamMsg.NewSsoPacket(0, data)
}

func ParseGetRoamMessagePacket(data []byte) ([]*message.CommonMessage, error) {
	rsp, e := proto.Unmarshal[message.SsoGetRoamMsgRsp](data)
	if e != nil {
		return nil, e
	}
	return rsp.Messages, nil
}

func BuildGetC2CMessagePacket(peer_uid string, start_seq, end_seq uint64) *sso_type.SsoPacket {
	data, _ := proto.Marshal(&message.SsoGetC2CMsgReq{
		PeerUid:       proto.Some(peer_uid),
		StartSequence: proto.Some(start_seq),
		EndSequence:   proto.Some(end_seq),
	})
	return message_type.AttributeSsoGetC2cMsg.NewSsoPacket(0, data)
}

func ParseGetC2CMessagePacket(data []byte) ([]*message.CommonMessage, error) {
	rsp, e := proto.Unmarshal[message.SsoGetC2CMsgRsp](data)
	if e != nil {
		return nil, e
	}
	if rsp.RetCode.Unwrap() != 0 {
		return nil, fmt.Errorf("%s (%d)", rsp.Message.Unwrap(), rsp.RetCode.Unwrap())
	}
	return rsp.Messages, nil
}
