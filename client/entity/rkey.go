package entity

type RKeyType uint64

const (
	FriendRKey RKeyType = 10
	GroupRKey  RKeyType = 20
)

type RKeyInfo struct {
	RKeyType   RKeyType
	RKey       string
	CreateTime uint64
	ExpireTime uint64
}

type RKeyMap map[RKeyType]*RKeyInfo
