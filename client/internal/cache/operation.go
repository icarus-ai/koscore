package cache

import "github.com/kernel-ai/koscore/client/entity"

func (c *Cache) RefreshAll(
	friendCache map[uint64]*entity.User,
	groupCache map[uint64]*entity.Group,
	groupMemberCache map[uint64]map[uint64]*entity.GroupMember,
	rkeyCache entity.RKeyMap,
) {
	c.RefreshAllFriend(friendCache)
	c.RefreshAllGroup(groupCache)
	c.RefreshAllGroupMembers(groupMemberCache)
	c.RefreshAllRKeyInfo(rkeyCache)
}

// 刷新一个好友的缓存
func (c *Cache) RefreshFriend(friend *entity.User) { setCacheOf(c, friend.Uin, friend) }

// 刷新所有好友缓存
func (c *Cache) RefreshAllFriend(friendCache map[uint64]*entity.User) {
	refreshAllCacheOf(c, friendCache)
}

// 刷新指定群的一个群成员缓存
func (c *Cache) RefreshGroupMember(groupUin uint64, groupMember *entity.GroupMember) {
	group, ok := getCacheOf[Cache](c, groupUin)
	if !ok {
		group = &Cache{}
		setCacheOf(c, groupUin, group)
	}
	setCacheOf(group, groupMember.Uin, groupMember)
}

// 刷新一个群内的所有群成员缓存
func (c *Cache) RefreshGroupMembers(groupUin uint64, groupMembers map[uint64]*entity.GroupMember) {
	newc := &Cache{}
	refreshAllCacheOf(newc, groupMembers)
	setCacheOf(c, groupUin, newc)
}

// 刷新所有群的群员缓存
func (c *Cache) RefreshAllGroupMembers(groupMemberCache map[uint64]map[uint64]*entity.GroupMember) {
	newc := make(map[uint64]*Cache, len(groupMemberCache)*2)
	for groupUin, v := range groupMemberCache {
		group := &Cache{}
		refreshAllCacheOf(group, v)
		newc[groupUin] = group
	}
	refreshAllCacheOf(c, newc)
}

// 刷新一个群的群信息缓存
func (c *Cache) RefreshGroup(group *entity.Group) { setCacheOf(c, group.GroupUin, group) }

// 刷新所有群的群信息缓存
func (c *Cache) RefreshAllGroup(groupCache map[uint64]*entity.Group) {
	refreshAllCacheOf(c, groupCache)
}

// 刷新所有RKey缓存
func (c *Cache) RefreshAllRKeyInfo(rkeyCache entity.RKeyMap) { refreshAllCacheOf(c, rkeyCache) }
