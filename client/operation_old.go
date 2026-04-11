package client

import (
	"io"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/message"
)

// SetOnlineStatus 设置在线状态
func (m *QQClient) SetOnlineStatus(status operation.SetStatus) error { return nil }

// SetGroupRemark 设置群聊备注
func (m *QQClient) SetGroupRemark(groupuin uint64, remark string) error { return nil }

// SetGroupName 设置群聊名称
func (m *QQClient) SetGroupName(groupuin uint64, name string) error { return nil }

// SetGroupGlobalMute 群全员禁言
func (m *QQClient) SetGroupGlobalMute(groupuin uint64, isMute bool) error { return nil }

// SetGroupMemberMute 禁言群成员
func (m *QQClient) SetGroupMemberMute(groupUin, uin uint64, duration uint32) error { return nil }

// SetGroupLeave 退出群聊
func (m *QQClient) SetGroupLeave(groupuin uint64) error { return nil }

// SetGroupAdmin 设置群管理员
func (m *QQClient) SetGroupAdmin(groupUin, uin uint64, isAdmin bool) error { return nil }

// SetGroupMemberName 设置群成员昵称
func (m *QQClient) SetGroupMemberName(groupUin, uin uint64, name string) error { return nil }

// KickGroupMember 踢出群成员，可选是否拒绝加群请求
func (m *QQClient) KickGroupMember(groupUin, uin uint64, rejectAddRequest bool) error { return nil }

// SetGroupMemberSpecialTitle 设置群成员专属头衔
func (m *QQClient) SetGroupMemberSpecialTitle(groupUin, uin uint64, title string) error { return nil }

// SetGroupReaction 设置群消息表态
func (m *QQClient) SetGroupReaction(groupUin uint64, sequence uint32, code string, isAdd bool) error {
	return nil
}

// GroupPoke 戳一戳群友
func (m *QQClient) GroupPoke(groupUin, uin uint64) error { return nil }

// FriendPoke 戳一戳好友
func (m *QQClient) FriendPoke(uin uint64) error { return nil }

// DeleteFriend 删除好友
func (m *QQClient) DeleteFriend(uin uint64, block bool) error { return nil }

// RecallFriendMessage 撤回私聊消息
func (m *QQClient) RecallFriendMessage(uin, seq, random, clientSeq uint64, timestamp int64) error {
	return nil
}

// RecallGroupMessage 撤回群聊消息
func (m *QQClient) RecallGroupMessage(gin, seq uint64) error { return nil }

// MarkPrivateMessageReaded 标记私聊消息已读
func (m *QQClient) MarkPrivateMessageReaded(uin uint64, timestamp, startSeq uint32) error { return nil }

// MarkGroupMessageReaded 标记群消息已读
func (m *QQClient) MarkGroupMessageReaded(gin uint64, seq uint32) error { return nil }

func (m *QQClient) GenFileNode(name, md5, sha1, uuid string, size uint32, isnt bool) *oidb.IndexNode {
	return nil
}

// QueryGroupImage 获取群图片
func (m *QQClient) QueryGroupImage(md5 []byte, fileUUid string) (*message.ImageElement, error) {
	return nil, nil
}

// QueryFriendImage 获取私聊图片
func (m *QQClient) QueryFriendImage(md5 []byte, fileUUid string) (*message.ImageElement, error) {
	return nil, nil
}

// FetchUserInfo 获取用户信息
func (m *QQClient) FetchUserInfo(uid string) (*entity.User, error) { return nil, nil }

// FetchUserInfoUin 通过uin获取用户信息
func (m *QQClient) FetchUserInfoUin(uin uint64) (*entity.User, error) { return nil, nil }

// FetchEssenceMessage 获取精华消息
func (m *QQClient) FetchEssenceMessage(groupuin uint64) ([]*message.GroupEssenceMessage, error) {
	return nil, nil
}

// GetGroupHonorInfo 获取群荣誉信息
// reference https://github.com/Mrs4s/MiraiGo/blob/master/client/http_api.go
func (m *QQClient) GetGroupHonorInfo(groupuin uint64, honorType entity.HonorType) (*entity.GroupHonorInfo, error) {
	return nil, nil
}

// GetGroupNotice 获取群公告
func (m *QQClient) GetGroupNotice(groupuin uint64) (l []*entity.GroupNoticeFeed, err error) {
	return nil, nil
}

func (m *QQClient) uploadGroupNoticePic(bkn int, img []byte) (*entity.NoticeImage, error) {
	return nil, nil
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
func (m *QQClient) DelGroupNotice(groupuin uint64, fid string) error { return nil }

// SetAvatar 设置头像
func (m *QQClient) SetAvatar(avatar io.ReadSeeker) error { return nil }

// SetGroupAvatar 设置群头像
func (m *QQClient) SetGroupAvatar(groupuin uint64, avatar io.ReadSeeker) error { return nil }

// SetEssenceMessage 设置群聊精华消息
func (m *QQClient) SetEssenceMessage(groupUin uint64, seq, random uint32, isSet bool) error {
	return nil
}

// SendFriendLike 给好友点赞
func (m *QQClient) SendFriendLike(uin uint64, count uint32) error { return nil }

func (m *QQClient) GetPrivateMessages(uin uint64, timestamp int64, count uint32) ([]*message.PrivateMessage, error) {
	return nil, nil
}

// GetGroupMessages 获取群聊历史消息
func (m *QQClient) GetGroupMessages(gin, start, end uint64) ([]*message.GroupMessage, error) {
	return nil, nil
}

// FetchMarketFaceKey 获取魔法表情key
func (m *QQClient) FetchMarketFaceKey(faceIds ...string) ([]string, error) { return nil, nil }
