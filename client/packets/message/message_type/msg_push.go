package message_type

import "github.com/kernel-ai/koscore/client/packets/structs/sso_type"

var AttributeMsgPush = sso_type.NewServiceAttributeD2D2("trpc.msg.olpush.OlPushService.MsgPush")

type MSG_TYPE uint16

// MsgType
const (
	GROUP_MESSAGE   = 82  // 群消息
	PRIVATE_MESSAGE = 166 // 私聊消息
	TEMP_MESSAGE    = 141 // 临时消息

	GROUP_MEMBER_INCREASE_NOTICE = 33 // 群成员增加
	GROUP_MEMBER_DECREASE_NOTICE = 34 // 群成员减少
	GROUP_ADMIN_CHANGED_NOTICE   = 44 // 群管理员变更
	GROUP_JOIN_NOTICE            = 84 // 加群通知
	GROUP_INVITE_NOTICE          = 87 // 群邀请

	Event0x20D   = 525 // 0x20D 群组请求邀请通知 group request invitation notice
	EVENT_FRIEND = 528 // 0x210 好友相关事件
	EVENT_GROUP  = 732 // 0x2DC 群相关事件
)
