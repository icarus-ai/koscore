package client

import (
	"io"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils/types"
)

// SetOnlineStatus 设置在线状态
func (m *QQClient) SetOnlineStatus(status operation.SetStatus) error { panic(types.ERROR_NOT_IMPL) }

// SetGroupRemark 设置群聊备注
func (m *QQClient) SetGroupRemark(groupuin uint64, remark string) error { panic(types.ERROR_NOT_IMPL) }

// SetGroupName 设置群聊名称
func (m *QQClient) SetGroupName(groupuin uint64, name string) error { panic(types.ERROR_NOT_IMPL) }

// SetGroupGlobalMute 群全员禁言
func (m *QQClient) SetGroupGlobalMute(groupuin uint64, isMute bool) error {
	panic(types.ERROR_NOT_IMPL)
}

// SetGroupMemberMute 禁言群成员
func (m *QQClient) SetGroupMemberMute(groupUin, uin uint64, duration uint32) error {
	panic(types.ERROR_NOT_IMPL)
}

// SetGroupLeave 退出群聊
func (m *QQClient) SetGroupLeave(groupuin uint64) error { panic(types.ERROR_NOT_IMPL) }

// SetGroupAdmin 设置群管理员
func (m *QQClient) SetGroupAdmin(groupUin, uin uint64, isAdmin bool) error {
	panic(types.ERROR_NOT_IMPL)
}

// SetGroupMemberName 设置群成员昵称
func (m *QQClient) SetGroupMemberName(groupUin, uin uint64, name string) error {
	panic(types.ERROR_NOT_IMPL)
}

// KickGroupMember 踢出群成员，可选是否拒绝加群请求
func (m *QQClient) KickGroupMember(groupUin, uin uint64, rejectAddRequest bool) error {
	panic(types.ERROR_NOT_IMPL)
}

// SetGroupMemberSpecialTitle 设置群成员专属头衔
func (m *QQClient) SetGroupMemberSpecialTitle(groupUin, uin uint64, title string) error {
	panic(types.ERROR_NOT_IMPL)
}

// SetGroupReaction 设置群消息表态
func (m *QQClient) SetGroupReaction(groupUin, sequence uint64, code string, isAdd bool) error {
	panic(types.ERROR_NOT_IMPL)
}

// GroupPoke 戳一戳群友
func (m *QQClient) GroupPoke(groupUin, uin uint64) error { panic(types.ERROR_NOT_IMPL) }

// FriendPoke 戳一戳好友
func (m *QQClient) FriendPoke(uin uint64) error { panic(types.ERROR_NOT_IMPL) }

// DeleteFriend 删除好友
func (m *QQClient) DeleteFriend(uin uint64, block bool) error { panic(types.ERROR_NOT_IMPL) }

// MarkPrivateMessageReaded 标记私聊消息已读
func (m *QQClient) MarkPrivateMessageReaded(uin uint64, timestamp int64, startSeq uint64) error {
	panic(types.ERROR_NOT_IMPL)
}

// MarkGroupMessageReaded 标记群消息已读
func (m *QQClient) MarkGroupMessageReaded(gin, seq uint64) error { panic(types.ERROR_NOT_IMPL) }

// QueryGroupImage 获取群图片
func (m *QQClient) QueryGroupImage(md5 []byte, fileUuid string) (*message.ImageElement, error) {
	panic(types.ERROR_NOT_IMPL)
}

// QueryFriendImage 获取私聊图片
func (m *QQClient) QueryFriendImage(md5 []byte, fileUuid string) (*message.ImageElement, error) {
	panic(types.ERROR_NOT_IMPL)
}

// FetchEssenceMessage 获取精华消息
func (m *QQClient) FetchEssenceMessage(groupuin uint64) ([]*message.GroupEssenceMessage, error) {
	panic(types.ERROR_NOT_IMPL)
}

// GetGroupHonorInfo 获取群荣誉信息
// reference https://github.com/Mrs4s/MiraiGo/blob/master/client/http_api.go
func (m *QQClient) GetGroupHonorInfo(groupuin uint64, honorType entity.HonorType) (*entity.GroupHonorInfo, error) {
	panic(types.ERROR_NOT_IMPL)
}

// GetGroupNotice 获取群公告
func (m *QQClient) GetGroupNotice(groupuin uint64) (l []*entity.GroupNoticeFeed, err error) {
	panic(types.ERROR_NOT_IMPL)
}

func (m *QQClient) uploadGroupNoticePic(bkn int, img []byte) (*entity.NoticeImage, error) {
	panic(types.ERROR_NOT_IMPL)
}

// AddGroupNoticeSimple 发群公告
func (m *QQClient) AddGroupNoticeSimple(groupuin uint64, text string) (noticeId string, err error) {
	return "", nil
}

// AddGroupNoticeWithPic 发群公告带图片
func (m *QQClient) AddGroupNoticeWithPic(groupuin uint64, text string, pic []byte) (noticeId string, err error) {
	return "", nil
}

// DelGroupNotice 删除群公告
func (m *QQClient) DelGroupNotice(groupuin uint64, fid string) error { panic(types.ERROR_NOT_IMPL) }

// SetAvatar 设置头像
func (m *QQClient) SetAvatar(avatar io.ReadSeeker) error { panic(types.ERROR_NOT_IMPL) }

// SetGroupAvatar 设置群头像
func (m *QQClient) SetGroupAvatar(groupuin uint64, avatar io.ReadSeeker) error {
	panic(types.ERROR_NOT_IMPL)
}

// SetEssenceMessage 设置群聊精华消息
func (m *QQClient) SetEssenceMessage(groupUin, seq, random uint64, isSet bool) error {
	panic(types.ERROR_NOT_IMPL)
}

// SendFriendLike 给好友点赞
func (m *QQClient) SendFriendLike(uin uint64, count uint32) error { panic(types.ERROR_NOT_IMPL) }

func (m *QQClient) GetPrivateMessages(uin uint64, timestamp int64, count uint32) ([]*message.PrivateMessage, error) {
	panic(types.ERROR_NOT_IMPL)
}

// FetchMarketFaceKey 获取魔法表情key
func (m *QQClient) FetchMarketFaceKey(faceIds ...string) ([]string, error) {
	panic(types.ERROR_NOT_IMPL)
}
