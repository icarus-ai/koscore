package oidb

import (
	"errors"
	"fmt"
	"time"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/common"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildFetchFriendsPacket(cookie []byte) (*sso_type.SsoPacket, error) {
	// BizData OidbNumber里面的东西代表你想要拿到的Property
	// 这些Property将会在返回的数据里面的Preserve的Field
	// 102   个性签名
	// 103   备注
	// 20002 昵称
	// 27394 QID
	return BuildOidbPacket(0xFD4, 1, &common.IncPullRequest{
		ReqCount: proto.Some[uint32](300),
		LocalSeq: proto.Some[uint32](13),
		Cookie:   cookie,
		Flag:     proto.Some[int32](1),
		ProxySeq: proto.Some[uint32](2147483647),
		RequestBiz: []*common.IncPullRequestBiz{
			{BizType: proto.Some[int32](1), BizData: &common.IncPullRequestBizBusi{ExtBusi: []int32{103, 102, 20002, 27394, 20009, 20037}}},
			{BizType: proto.Some[int32](4), BizData: &common.IncPullRequestBizBusi{ExtBusi: []int32{100, 101, 102}}},
		},
		ExtSnsFlagKey:       []uint32{13578, 13579, 13573, 13572, 13568},
		ExtPrivateIdListKey: []uint32{4051},
	}, false, false)
}

type FetchFriendsRsp struct {
	Friends  []*entity.User
	Category []*entity.UserCategory
	Cookie   []byte
}

func ParseFetchFriendsPacket(data []byte) (*FetchFriendsRsp, error) {
	rsp, e := ParseOidbPacket[common.IncPullResponse](data)
	if e != nil {
		return nil, e
	}

	friends := make([]*entity.User, len(rsp.FriendList))
	categories := make([]*entity.UserCategory, len(rsp.Category))
	category_map := make(map[int32]*entity.UserCategory)
	for i, raw := range rsp.Category {
		category := &entity.UserCategory{
			Id:     raw.CategoryId.Unwrap(),
			Name:   raw.CategoryName.Unwrap(),
			Count:  raw.CategoryMemberCount.Unwrap(),
			SortId: raw.CatogorySortId.Unwrap(),
		}
		categories[i], category_map[raw.CategoryId.Unwrap()] = category, category
	}
	for i, raw := range rsp.FriendList {
		friends[i] = &entity.User{
			Uin:          uint64(raw.Uin.Unwrap()),
			Uid:          raw.Uid.Unwrap(),
			Nickname:     raw.SubBiz[1].Data[20002],
			Remarks:      raw.SubBiz[1].Data[103],
			PersonalSign: raw.SubBiz[1].Data[102],
			Avatar:       entity.UserAvatar(uint64(raw.Uin.Unwrap())),
			QID:          raw.SubBiz[1].Data[27394],
			Age:          uint32(raw.SubBiz[1].NumData[20037]),
			Sex:          entity.GenderInfo(raw.SubBiz[1].NumData[20009]),
			Category:     category_map[raw.CategoryId.Unwrap()],
		}
	}
	return &FetchFriendsRsp{
		Friends:  friends,
		Category: categories,
		Cookie:   rsp.Cookie,
	}, nil
}

var fetch_strange_keys []*operation.FetchStrangerRequestKey

func BuildFetchStrangerPacket[T uint64 | string](id T, sub uint32) (*sso_type.SsoPacket, error) {
	if len(fetch_strange_keys) == 0 {
		for _, v := range []uint64{
			101,   // avatar 头像
			102,   // sign 简介/签名
			103,   // remark 备注
			104,   // tag
			105,   // level 等级
			107,   // business 业务列表
			20002, // nickname 昵称
			20003, // country 国家
			20004, // city
			20005,
			20006, // home city
			20009, // gender; 1 Male 2 Female 255 Unknown 性别
			20011, // eMail
			20016, // desensitized mobile phone number
			20020, // municipal district 城市
			20021, // school 学校
			20022, 20023, 20024,
			20026, // registration time; Only year, hour, minute, second 注册时间
			20031, // birthday 生日
			20037, // age 年龄
			24002,
			24007, //?游戏标签
			27037, 27049,
			27372, // 状态
			27394, // QId
			27406, // 自定义状态文本
			41756, 41757, 42257, 42315, 42362, 42432, 45160, 45161, 62026,
		} {
			fetch_strange_keys = append(fetch_strange_keys, &operation.FetchStrangerRequestKey{Key: proto.Some(v)})
		}
	}

	switch o := any(id).(type) {
	case uint64:
		if sub != 2 {
			sub = 2
		}
		return BuildOidbPacket(0xFE1, sub, &operation.FetchStrangerByUinRequest{
			Uin:  proto.Some(int64(o)),
			Keys: fetch_strange_keys,
		}, false, false)
	case string:
		if sub != 2 && sub != 8 {
			sub = 2
		}
		return BuildOidbPacket(0xFE1, sub, &operation.FetchStrangerByUidRequest{
			Uid:  proto.Some(o),
			Keys: fetch_strange_keys,
		}, false, true)
	default:
		return nil, nil
	}
}

func ParseFetchStrangerPacket(data []byte) (*entity.User, error) {
	rsp, e := ParseOidbPacket[operation.FetchStrangerResponse](data)
	if e != nil {
		return nil, e
	}

	byt_map := make(map[uint64][]byte)
	num_map := make(map[uint64]uint64)
	for _, property := range rsp.Body.Properties.BytesProperties {
		byt_map[property.Key.Unwrap()] = property.Value
	}
	for _, propetry := range rsp.Body.Properties.NumberProperties {
		num_map[propetry.Key.Unwrap()] = propetry.Value.Unwrap()
	}

	nickname_bytes, ok := byt_map[20002]
	if !ok {
		return nil, errors.New("operation exception: Stranger not found")
	}

	// can't not get uid
	ret := &entity.User{
		Uin:          uint64(rsp.Body.Uin.Unwrap()),
		Nickname:     string(nickname_bytes),
		PersonalSign: string(byt_map[102]),
		Remarks:      string(byt_map[103]),
		Level:        uint32(num_map[105]),
		Sex:          entity.GenderInfo(num_map[20009]),
		Registration: int64(num_map[20026]),
		Age:          uint32(num_map[20037]),
		QID:          string(byt_map[27394]),
		Country:      string(byt_map[20003]),
		City:         string(byt_map[20004]),
		School:       string(byt_map[20021]),
	}

	// 生日 07d00b1d
	if data = byt_map[20031]; len(data) >= 4 {
		date := fmt.Sprintf("%04d/%02d/%02d", (int(data[0])<<8|int(data[1]))+1, data[2], data[3])
		uxt, _ := time.ParseInLocation("2006/01/02", date, time.FixedZone("UTC+8", 8*3600))
		ret.Birthday = uxt.Unix()
	}
	if data = byt_map[27406]; len(data) > 0 {
		if customs, e := proto.Unmarshal[operation.CustomStatus](data); e == nil {
			statusId := num_map[27372]
			mask := uint32((268435455 - statusId) >> 31)
			ret.Status.StatusId = uint32(statusId - uint64(268435456&mask))
			ret.Status.FaceId = customs.FaceId
			ret.Status.Msg = customs.Msg.Unwrap()
		}
	}
	if data = byt_map[101]; len(data) > 0 {
		if avatar, e := proto.Unmarshal[operation.Avatar](data); e == nil {
			ret.Avatar = avatar.Url.Unwrap() + "640"
		}
	}
	if data = byt_map[107]; len(data) > 0 {
		if business, e := proto.Unmarshal[operation.Business](data); e == nil {
			for _, v := range business.Body.Lists {
				ret.Business = append(ret.Business, entity.BusinessCustom{
					Type:  entity.BusinessType(v.Type),
					Level: v.Level,
					Icon: func() string {
						if x, ok := v.Icon.(*operation.BusinessList_Icon1); ok {
							return x.Icon1
						}
						if x, ok := v.Icon.(*operation.BusinessList_Icon2); ok {
							return x.Icon2
						}
						return ""
					}(),
					IsPro:  v.IsPro,
					IsYear: v.IsYear,
				})
			}
		}
	}
	return ret, nil
}
