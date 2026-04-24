package oidb

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/proto"
)

// 处理好友请求
func BuildSetFriendRequestPacket(accept bool, target_uid string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0xb5d, 44, &oidb.OidbSvcTrpcTcp0XB5D_44{
		Accept:    utils.Ternary[uint32](accept, 3, 5),
		TargetUid: target_uid,
	}, false, false)
}

// 给好友点赞
func BuildFriendLikePacket(uid string, count uint32) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x7E5, 104, &oidb.OidbSvcTrpcTcp0X7E5_104{
		TargetUid: proto.Some(uid),
		Source:    71,
		Count:     count,
	}, false, false)
}

// 删除好友
func BuildDeleteFriendPacket(uid string, block bool) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x126B, 0, &oidb.OidbSvcTrpcTcp0X126B_0{
		Field1: &oidb.OidbSvcTrpcTcp0X126B_0_Field1{
			TargetUid: uid,
			Field2: &oidb.OidbSvcTrpcTcp0X126B_0_Field1_2{
				Field1: 130,
				Field2: 109,
				Field3: &oidb.OidbSvcTrpcTcp0X126B_0_Field1_2_3{
					Field1: 8,
					Field2: 8,
					Field3: 50,
				},
			},
			Block:  block,
			Field4: true,
		},
	}, false, false)
}
