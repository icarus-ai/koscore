package client

import (
	"time"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/utils/exception"
)

func (m *QQClient) GetUid(uin uint64, gin ...uint64) (string, error) {
	uid := m.get_uid(uin, gin...)
	if uid == "" {
		return "", exception.NewInvalidTargetException(uin, gin...)
	}
	return uid, nil
}

// 获取缓存中对应uin的uid
func (m *QQClient) get_uid(uin uint64, gin ...uint64) string {
	if uin == 0 {
		return ""
	}
	if len(gin) == 0 {
		if m.cache.FriendCacheIsEmpty() {
			if e := m.RefreshFriendCache(); e != nil {
				return ""
			}
		}
	} else if m.cache.GroupMemberCacheIsEmpty(gin[0]) {
		if e := m.RefreshGroupMembersCache(gin[0]); e != nil {
			return ""
		}
	}
	if uid := m.cache.GetUid(uin, gin...); uid != "" {
		return uid
	}
	if len(gin) == 0 {
		_ = m.RefreshFriendCache()
	} else {
		_ = m.RefreshGroupMembersCache(gin[0])
	}
	return m.cache.GetUid(uin, gin...)
}

// 获取缓存中对应的uin
func (m *QQClient) GetUin(uid string, gin ...uint64) uint64 {
	if uid == "" {
		return 0
	}
	if len(gin) == 0 {
		if m.cache.FriendCacheIsEmpty() {
			if e := m.RefreshFriendCache(); e != nil {
				return 0
			}
		}
	} else if m.cache.GroupMemberCacheIsEmpty(gin[0]) {
		if e := m.RefreshGroupMembersCache(gin[0]); e != nil {
			return 0
		}
	}
	if uin := m.cache.GetUin(uid, gin...); uin != 0 {
		return uin
	}
	if len(gin) == 0 {
		_ = m.RefreshFriendCache()
	} else {
		_ = m.RefreshGroupMembersCache(gin[0])
	}
	return m.cache.GetUin(uid, gin...)
}

// 获取好友信息(缓存)
func (m *QQClient) GetCachedFriendInfo(uin uint64, cache ...bool) *entity.User {
	if m.cache.FriendCacheIsEmpty() {
		if e := m.RefreshFriendCache(); e != nil {
			return nil
		}
	}
	if fr := m.cache.GetFriend(uin); fr != nil {
		return fr
	}
	if len(cache) > 0 && cache[0] {
		return nil
	}
	_ = m.RefreshFriendCache()
	return m.cache.GetFriend(uin)
}

// 获取所有好友信息(缓存)
func (m *QQClient) GetCachedAllFriendsInfo() map[uint64]*entity.User {
	if m.cache.FriendCacheIsEmpty() {
		if e := m.RefreshFriendCache(); e != nil {
			return nil
		}
	}
	return m.cache.GetAllFriends()
}

// 获取群信息(缓存)
func (m *QQClient) GetCachedGroupInfo(gin uint64) *entity.Group {
	if m.cache.GroupInfoCacheIsEmpty() {
		if e := m.RefreshAllGroupsInfo(); e != nil {
			return nil
		}
	}
	if g := m.cache.GetGroupInfo(gin); g != nil {
		return g
	}
	_ = m.RefreshAllGroupsInfo()
	return m.cache.GetGroupInfo(gin)
}

// 获取所有群信息(缓存)
func (m *QQClient) GetCachedAllGroupsInfo() map[uint64]*entity.Group {
	if m.cache.GroupInfoCacheIsEmpty() {
		if err := m.RefreshAllGroupsInfo(); err != nil {
			return nil
		}
	}
	return m.cache.GetAllGroupsInfo()
}

// 获取群成员信息(缓存)
func (m *QQClient) GetCachedMemberInfo(uin, gin uint64, cache ...bool) *entity.GroupMember {
	if m.cache.GroupMemberCacheIsEmpty(gin) {
		if err := m.RefreshGroupMemberCache(gin, uin); err != nil {
			return nil
		}
	}
	if m := m.cache.GetGroupMember(uin, gin); m != nil {
		return m
	}
	if len(cache) > 0 && cache[0] {
		return nil
	}
	_ = m.RefreshGroupMemberCache(uin, gin)
	return m.cache.GetGroupMember(uin, gin)
}

// 获取指定群所有群成员信息(缓存)
func (m *QQClient) GetCachedMembersInfo(gin uint64) map[uint64]*entity.GroupMember {
	if m.cache.GroupMemberCacheIsEmpty(gin) {
		if err := m.RefreshGroupMembersCache(gin); err != nil {
			return nil
		}
	}
	if gm := m.cache.GetGroupMembers(gin); gm != nil {
		return gm
	}
	_ = m.RefreshGroupMembersCache(gin)
	return m.cache.GetGroupMembers(gin)
}

// 获取指定类型的RKey信息(缓存)
// ??? king GetCachedRkeyInfo 有问题
func (m *QQClient) GetCachedRkeyInfo(rkeyType entity.RKeyType) (*entity.RKeyInfo, error) {
	for {
		if !m.cache.RkeyInfoCacheIsEmpty() {
			inf := m.cache.GetRKeyInfo(rkeyType)
			if int64((inf.ExpireTime+inf.CreateTime)/2) > time.Now().Unix() {
				return inf, nil
			}
		}
		if e := m.RefreshAllRkeyInfoCache(); e != nil {
			return nil, e
		}
	}
}

// 获取所有RKey信息(缓存)
func (m *QQClient) GetCachedRkeyInfos() map[entity.RKeyType]*entity.RKeyInfo {
	for {
		if !m.cache.RkeyInfoCacheIsEmpty() {
			ok, inf := true, m.cache.GetAllRkeyInfo()
			for _, v := range inf {
				if int64((v.ExpireTime+v.CreateTime)/2) <= time.Now().Unix() {
					ok = false
					break
				}
			}
			if ok {
				return inf
			}
		}
		if e := m.RefreshAllRkeyInfoCache(); e != nil {
			return nil
		}
	}
	/*
	   refresh := m.cache.RkeyInfoCacheIsEmpty()

	   	for {
	   		if refresh {
	   			if e := m.RefreshAllRkeyInfoCache(); e != nil { return nil }
	   			refresh = false
	   		}
	   		inf := m.cache.GetAllRkeyInfo()
	   		for _, v := range inf {
	   			if v.ExpireTime <= uint64(time.Now().Unix()) {
	   				refresh = true
	   				break
	   		} }
	   		if refresh { continue }
	   		return inf
	   	}
	*/
}

// 刷新RKey缓存
func (m *QQClient) RefreshAllRkeyInfoCache() error {
	info, e := m.FetchRkey()
	if e != nil {
		return e
	}
	m.cache.RefreshAllRKeyInfo(info)
	return nil
}

// 刷新好友缓存
func (m *QQClient) RefreshFriendCache() error {
	friends, e := m.GetFriendsData()
	if e != nil {
		return e
	}
	m.cache.RefreshAllFriend(friends)
	unfriends, e := m.GetUnidirectionalFriendList()
	if e != nil {
		return e
	}
	for _, item := range unfriends {
		m.cache.RefreshFriend(item)
	}
	return nil
}

// OLD_CODE 刷新一个群的指定群成员缓存
func (m *QQClient) RefreshGroupMemberCache(gin, member_uin uint64) error {
	mem, e := m.FetchGroupMember(gin, member_uin)
	if e != nil {
		return e
	}
	m.cache.RefreshGroupMember(gin, mem)
	return nil
}

// 刷新指定群的所有群成员缓存
func (m *QQClient) RefreshGroupMembersCache(gin uint64) error {
	data, err := m.GetGroupMembersData(gin)
	if err != nil {
		return err
	}
	m.cache.RefreshGroupMembers(gin, data)
	return nil
}

// 刷新所有群的群成员缓存
func (m *QQClient) RefreshAllGroupMembersCache() error {
	data, err := m.GetAllGroupsMembersData()
	if err != nil {
		return err
	}
	m.cache.RefreshAllGroupMembers(data)
	return nil
}

// 刷新所有群信息缓存
func (m *QQClient) RefreshAllGroupsInfo() error {
	data, err := m.GetAllGroupsInfo()
	if err != nil {
		return err
	}
	m.cache.RefreshAllGroup(data)
	return nil
}

// 获取好友列表数据
func (m *QQClient) GetFriendsData() (map[uint64]*entity.User, error) {
	rsp, err := m.FetchFriends(nil)
	if err != nil {
		return nil, err
	}
	data := make(map[uint64]*entity.User)
	for {
		if rsp, err = m.FetchFriends(rsp.Cookie); err != nil {
			return nil, err
		}
		for _, friend := range rsp.Friends {
			data[friend.Uin] = friend
		}
		if len(rsp.Cookie) == 0 {
			break
		}
	}
	m.LOGD("获取%d个好友", len(data))
	return data, err
}

// 获取指定群所有成员信息
func (m *QQClient) GetGroupMembersData(gin uint64) (map[uint64]*entity.GroupMember, error) {
	members, token, err := m.FetchGroupMembers(gin, nil)
	if err != nil {
		return nil, err
	}
	data := make(map[uint64]*entity.GroupMember)
	for _, item := range members {
		data[item.Uin] = item
	}
	for token != nil {
		if members, token, err = m.FetchGroupMembers(gin, token); err != nil {
			return nil, err
		}
		for _, item := range members {
			data[item.Uin] = item
		}
	}
	return data, err
}

// 获取所有群的群成员信息
func (m *QQClient) GetAllGroupsMembersData() (map[uint64]map[uint64]*entity.GroupMember, error) {
	groups, err := m.FetchGroups()
	if err != nil {
		return nil, err
	}
	data := make(map[uint64]map[uint64]*entity.GroupMember)
	for _, item := range groups {
		members, err := m.GetGroupMembersData(item.GroupUin)
		if err != nil {
			return nil, err
		}
		data[item.GroupUin] = members
	}
	m.LOGD("获取%d个群的成员信息", len(data))
	return data, err
}

func (m *QQClient) GetAllGroupsInfo() (map[uint64]*entity.Group, error) {
	groups, err := m.FetchGroups()
	if err != nil {
		return nil, err
	}
	data := make(map[uint64]*entity.Group)
	for _, item := range groups {
		data[item.GroupUin] = item
	}
	m.LOGD("获取%d个群信息", len(data))
	return data, err
}

// ***** add date 20260514 *****

// 刷新表情信息
func (m *QQClient) RefreshFaceDetails() (int, error) {
	details, e := m.FaceDetails()
	if e != nil {
		return 0, e
	}
	m.face_details.Set(details)
	m.LOGD("获取%d个表情信息", len(details))
	return len(details), nil
}

// 获取表情信息
func (m *QQClient) GetFaceDetail(qsid string) (*entity.BotFaceDetail, error) {
	if m.face_details.IsEmpty() {
		if _, e := m.RefreshFaceDetails(); e != nil {
			return nil, e
		}
	}
	return m.face_details.Get(qsid), nil
}
