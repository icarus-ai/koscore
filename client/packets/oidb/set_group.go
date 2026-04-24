package oidb

import (
	"math"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/proto"
)

// 处理加群请求
func BuildSetGroupRequestPacket(isFiltered bool, operate entity.GroupRequestOperate, sequence uint64, typ uint32, groupUin uint64, message string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x10C8, utils.Ternary[uint32](isFiltered, 2, 1), &operation.SetGroupNotificationRequest{
		Operate: proto.Some(uint64(operate)),
		Body: &operation.SetGroupNotificationRequestBody{
			Sequence: proto.Some(sequence),
			Type:     proto.Some(uint64(typ)),
			GroupUin: proto.Some(int64(groupUin)),
			Message:  proto.Some(message),
		},
	}, false, false)
}

// 设置群消息表态
func BuildGroupReactionPacket(groupUin uint64, sequence uint64, code string, isAdd bool) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x9082, utils.Ternary[uint32](isAdd, 1, 2), &operation.SetGroupReactionRequest{
		GroupUin: proto.Some(int64(groupUin)),
		Sequence: proto.Some(sequence),
		Code:     proto.Some(code),
		Type:     proto.Some(utils.Ternary[uint64](len(code) > 3, 2, 1)),
	}, false, false)
}

// 群管理
func BuildSetGroupAdminPacket(group_uin uint64, uid string, is_admin bool) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x1096, 1, &oidb.OidbSvcTrpcTcp0X1096_1{
		GroupUin: group_uin,
		Uid:      uid,
		IsAdmin:  is_admin,
	}, false, false)
}

// 群全员禁言|解除禁言
func BuildSetGroupGlobalMutePacket(group_uin uint64, is_mute bool) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x89A, 0, &oidb.D89AReqBody{
		GroupCode: proto.Some(int64(group_uin)),
		Group: &oidb.D89AReqBodyGroupInfo{
			ShutupTime: proto.Some(utils.Ternary[uint32](is_mute, math.MaxUint32, 0)),
		},
	}, false, false)
}

func ParseSetGroupGlobalMutePacket(data []byte) error {
	return CheckTypedError[oidb.D89ARspBody](data)
}

// 禁言群成员
func BuildSetGroupMemberMutePacket(group_uin uint64, uid string, duration uint32) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x1253, 1, &oidb.OidbSvcTrpcTcp0X1253_1{
		GroupUin: group_uin,
		Type:     1,
		Body: &oidb.OidbSvcTrpcTcp0X1253_1Body{
			TargetUid: uid,
			Duration:  duration,
		},
	}, false, false)
}

func ParseSetGroupMemberMutePacket(data []byte) error {
	return CheckTypedError[oidb.OidbSvcTrpcTcp0X1253_1Response](data)
}

// 设置群聊备注
func BuildSetGroupRemarkPacket(group_uin uint64, mark string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0xF16, 1, &oidb.OidbSvcTrpcTcp0XF16_1{
		Body: &oidb.OidbSvcTrpcTcp0XF16_1Body{
			GroupUin:     group_uin,
			TargetRemark: mark,
		}}, false, false)
}

// 设置群聊名称
func BuildSetGroupNamePacket(group_uin uint64, name string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x89A, 15, &oidb.D89AReqBody{
		GroupCode: proto.Some(int64(group_uin)),
		Group:     &oidb.D89AReqBodyGroupInfo{GroupName: proto.Some(name)},
	}, false, false)
}

// 设置群成员昵称
func BuildSetGroupMemberNamePacket(group_uin uint64, uid, name string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x8FC, 3, &oidb.D8FCReqBody{
		GroupCode: proto.Some(int64(group_uin)),
		MemLevelInfo: []*oidb.MemberInfo{{
			Uid:            proto.Some(uid),
			MemberCardName: []byte(name),
		}},
	}, false, false)
}

// 设置群成员专属头衔
func BuildSetGroupMemberSpecialTitlePacket(group_uin uint64, uid, title string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x8FC, 2, &oidb.D8FCReqBody{
		GroupCode: proto.Some(int64(group_uin)),
		MemLevelInfo: []*oidb.MemberInfo{{
			Uid:          proto.Some(uid),
			SpecialTitle: []byte(title),
		}},
	}, false, false)
}

// 退出群聊
func BuildSetGroupLeavePacket(group_uin uint64) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x1097, 1, &oidb.D1097ReqBody{GroupCode: proto.Some(int64(group_uin))}, false, false)
}

// 踢出群成员，可选是否拒绝加群请求
func BuildKickGroupMemberPacket(group_uin uint64, uid string, reject_add_request bool) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x8A0, 1, &oidb.OidbSvcTrpcTcp0X8A0_1{
		GroupUin:         group_uin,
		TargetUid:        uid,
		RejectAddRequest: reject_add_request,
	}, false, false)
}

func ParseKickGroupMemberPacket(data []byte) error {
	return CheckTypedError[oidb.OidbSvcTrpcTcp0X8A0_1Response](data)
}

// 设置群聊精华消息
func BuildSetEssenceMessagePacket(group_uin, seq, random uint64, is_set bool) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0xEAC, utils.Ternary[uint32](is_set, 1, 2), &oidb.OidbSvcTrpcTcp0XEAC{
		GroupUin: group_uin,
		Sequence: seq,
		Random:   random,
	}, false, false)
}
