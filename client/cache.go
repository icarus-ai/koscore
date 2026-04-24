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

// GetUid 获取缓存中对应uin的uid
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

// GetUin 获取缓存中对应的uin
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

// GetCachedFriendInfo 获取好友信息(缓存)
func (m *QQClient) GetCachedFriendInfo(uin uint64) *entity.User {
	if m.cache.FriendCacheIsEmpty() {
		if e := m.RefreshFriendCache(); e != nil {
			return nil
		}
	}
	if fr := m.cache.GetFriend(uin); fr != nil {
		return fr
	}
	_ = m.RefreshFriendCache()
	return m.cache.GetFriend(uin)
}

// GetCachedAllFriendsInfo 获取所有好友信息(缓存)
func (m *QQClient) GetCachedAllFriendsInfo() map[uint64]*entity.User {
	if m.cache.FriendCacheIsEmpty() {
		if e := m.RefreshFriendCache(); e != nil {
			return nil
		}
	}
	return m.cache.GetAllFriends()
}

// GetCachedGroupInfo 获取群信息(缓存)
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

// GetCachedAllGroupsInfo 获取所有群信息(缓存)
func (m *QQClient) GetCachedAllGroupsInfo() map[uint64]*entity.Group {
	if m.cache.GroupInfoCacheIsEmpty() {
		if err := m.RefreshAllGroupsInfo(); err != nil {
			return nil
		}
	}
	return m.cache.GetAllGroupsInfo()
}

// GetCachedMemberInfo 获取群成员信息(缓存)
func (m *QQClient) GetCachedMemberInfo(uin, groupUin uint64) *entity.GroupMember {
	if m.cache.GroupMemberCacheIsEmpty(groupUin) {
		if err := m.RefreshGroupMemberCache(groupUin, uin); err != nil {
			return nil
		}
	}
	if m := m.cache.GetGroupMember(uin, groupUin); m != nil {
		return m
	}
	_ = m.RefreshGroupMemberCache(uin, groupUin)
	return m.cache.GetGroupMember(uin, groupUin)
}

// GetCachedMembersInfo 获取指定群所有群成员信息(缓存)
func (m *QQClient) GetCachedMembersInfo(groupUin uint64) map[uint64]*entity.GroupMember {
	if m.cache.GroupMemberCacheIsEmpty(groupUin) {
		if err := m.RefreshGroupMembersCache(groupUin); err != nil {
			return nil
		}
	}
	if gm := m.cache.GetGroupMembers(groupUin); gm != nil {
		return gm
	}
	_ = m.RefreshGroupMembersCache(groupUin)
	return m.cache.GetGroupMembers(groupUin)
}

// ??? king GetCachedRkeyInfo 有问题

// GetCachedRkeyInfo 获取指定类型的RKey信息(缓存)
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

// GetCachedRkeyInfos 获取所有RKey信息(缓存)
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

// RefreshAllRkeyInfoCache 刷新RKey缓存
func (m *QQClient) RefreshAllRkeyInfoCache() error {
	info, e := m.FetchRkey()
	if e != nil {
		return e
	}
	m.cache.RefreshAllRKeyInfo(info)
	return nil
}

// RefreshFriendCache 刷新好友缓存
func (m *QQClient) RefreshFriendCache() error {
	friends, e := m.GetFriendsData()
	if e != nil {
		return e
	}
	m.cache.RefreshAllFriend(friends)
	unidirectionalFriends, e := m.GetUnidirectionalFriendList()
	if e != nil {
		return e
	}
	for _, f := range unidirectionalFriends {
		m.cache.RefreshFriend(f)
	}
	return nil
}

// OLD_CODE 刷新一个群的指定群成员缓存
func (m *QQClient) RefreshGroupMemberCache(groupUin, memberUin uint64) error {
	mem, e := m.FetchGroupMember(groupUin, memberUin)
	if e != nil {
		return e
	}
	m.cache.RefreshGroupMember(groupUin, mem)
	return nil
}

// 刷新指定群的所有群成员缓存
func (m *QQClient) RefreshGroupMembersCache(groupUin uint64) error {
	groupData, err := m.GetGroupMembersData(groupUin)
	if err != nil {
		return err
	}
	m.cache.RefreshGroupMembers(groupUin, groupData)
	return nil
}

// RefreshAllGroupMembersCache 刷新所有群的群成员缓存
func (m *QQClient) RefreshAllGroupMembersCache() error {
	groupsData, err := m.GetAllGroupsMembersData()
	if err != nil {
		return err
	}
	m.cache.RefreshAllGroupMembers(groupsData)
	return nil
}

// RefreshAllGroupsInfo 刷新所有群信息缓存
func (m *QQClient) RefreshAllGroupsInfo() error {
	groupsData, err := m.GetAllGroupsInfo()
	if err != nil {
		return err
	}
	m.cache.RefreshAllGroup(groupsData)
	return nil
}

// GetFriendsData 获取好友列表数据
func (m *QQClient) GetFriendsData() (map[uint64]*entity.User, error) {
	rsp, err := m.FetchFriends(nil)
	if err != nil {
		return nil, err
	}

	friends := make(map[uint64]*entity.User)
	for {
		rsp, err = m.FetchFriends(rsp.Cookie)
		if err != nil {
			return friends, err
		}
		for _, friend := range rsp.Friends {
			friends[friend.Uin] = friend
		}
		if len(rsp.Cookie) == 0 {
			break
		}
	}
	m.LOGD("获取%d个好友", len(friends))
	return friends, err
}

// GetGroupMembersData 获取指定群所有成员信息
func (m *QQClient) GetGroupMembersData(groupUin uint64) (map[uint64]*entity.GroupMember, error) {
	groupMembers := make(map[uint64]*entity.GroupMember)
	members, token, err := m.FetchGroupMembers(groupUin, nil)
	if err != nil {
		return groupMembers, err
	}
	for _, member := range members {
		groupMembers[member.Uin] = member
	}
	for token != nil {
		members, token, err = m.FetchGroupMembers(groupUin, token)
		if err != nil {
			return groupMembers, err
		}
		for _, member := range members {
			groupMembers[member.Uin] = member
		}
	}
	return groupMembers, err
}

// GetAllGroupsMembersData 获取所有群的群成员信息
func (m *QQClient) GetAllGroupsMembersData() (map[uint64]map[uint64]*entity.GroupMember, error) {
	groups, err := m.FetchGroups()
	if err != nil {
		return nil, err
	}
	groupsData := make(map[uint64]map[uint64]*entity.GroupMember, len(groups))
	for _, group := range groups {
		groupMembersData, err := m.GetGroupMembersData(group.GroupUin)
		if err != nil {
			return nil, err
		}
		groupsData[group.GroupUin] = groupMembersData
	}
	m.LOGD("获取%d个群的成员信息", len(groupsData))
	return groupsData, err
}

func (m *QQClient) GetAllGroupsInfo() (map[uint64]*entity.Group, error) {
	groupsInfo, err := m.FetchGroups()
	if err != nil {
		return nil, err
	}
	groupsData := make(map[uint64]*entity.Group, len(groupsInfo))
	for _, group := range groupsInfo {
		groupsData[group.GroupUin] = group
	}
	m.LOGD("获取%d个群信息", len(groupsData))
	return groupsData, err
}
