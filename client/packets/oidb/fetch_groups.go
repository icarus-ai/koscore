package oidb

import (
	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/internal/cache"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildFetchGroupsPacket() (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0xFE5, 2, &operation.FetchGroupsRequest{
		Config: &operation.FetchGroupsRequestConfig{
			Config1: &operation.FetchGroupsRequestConfig1{
				GroupOwner: proto.TRUE, MemberMax: proto.TRUE, MemberCount: proto.TRUE, GroupName: proto.TRUE, Question: proto.TRUE,
				Field2: proto.TRUE, Field8: proto.TRUE, Field9: proto.TRUE, Field10: proto.TRUE, Field11: proto.TRUE, Field12: proto.TRUE, Field13: proto.TRUE,
				Field14: proto.TRUE, Field15: proto.TRUE, Field16: proto.TRUE, Field17: proto.TRUE, Field18: proto.TRUE, Field20: proto.TRUE, Field22: proto.TRUE,
				Field23: proto.TRUE, Field24: proto.TRUE, Field25: proto.TRUE, Field26: proto.TRUE, Field27: proto.TRUE, Field28: proto.TRUE, Field29: proto.TRUE,
				Field30: proto.TRUE, Field31: proto.TRUE, Field32: proto.TRUE, Field5001: proto.TRUE, Field5002: proto.TRUE, Field5003: proto.TRUE,
			},
			Config2: &operation.FetchGroupsRequestConfig2{Field1: proto.TRUE, Field2: proto.TRUE, Field3: proto.TRUE, Field4: proto.TRUE, Field5: proto.TRUE, Field6: proto.TRUE, Field7: proto.TRUE, Field8: proto.TRUE},
			Config3: &operation.FetchGroupsRequestConfig3{Field5: proto.TRUE, Field6: proto.TRUE},
		},
	}, false, false)
}

func ParseFetchGroupsPacket(data []byte) ([]*entity.Group, error) {
	rsp, e := ParseOidbPacket[operation.FetchGroupsResponse](data)
	if e != nil {
		return nil, e
	}
	groups := make([]*entity.Group, len(rsp.Groups))
	for i, group := range rsp.Groups {
		groups[i] = &entity.Group{
			GroupUin:        uint64(group.GroupUin.Unwrap()),
			GroupName:       group.Info.GroupName.Unwrap(),
			MemberCount:     group.Info.MemberCount.Unwrap(),
			MaxMember:       group.Info.MemberMax.Unwrap(),
			GroupCreateTime: group.Info.CreatedTime.Unwrap(),
			Description:     group.Info.Description.Unwrap(),
			Question:        group.Info.Question.Unwrap(),
			Announcement:    group.Info.Announcement.Unwrap(),
		}
	}
	return groups, nil
}

func BuildFetchGroupMemberPacket(gin uint64, member_uid string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0xFE7, 4, &operation.FetchGroupMemberRequest{
		GroupUin: proto.Some(int64(gin)),
		Field2:   proto.Some[uint32](5),
		Field3:   proto.Some[uint32](2),
		Body: &operation.FetchGroupMembersRequestBody{
			MemberName:       proto.TRUE,
			MemberCard:       proto.TRUE,
			SpecialTitle:     proto.TRUE,
			Level:            proto.TRUE,
			JoinTimestamp:    proto.TRUE,
			LastMsgTimestamp: proto.TRUE,
			ShutUpTimestamp:  proto.TRUE,
			Permission:       proto.TRUE,
		},
		Params: &operation.OidbSvcTrpcScp0XFE7_4Params{Uid: proto.Some(member_uid)},
	}, false, false)
}

func ParseFetchGroupMemberPacket(data []byte) (*entity.GroupMember, error) {
	rsp, e := ParseOidbPacket[operation.FetchGroupMemberResponse](data)
	if e != nil {
		return nil, e
	}
	raw := rsp.Member
	return &entity.GroupMember{
		User: entity.User{
			Uin:      uint64(raw.Id.Uin.Unwrap()),
			Uid:      raw.Id.Uid.Unwrap(),
			Nickname: raw.MemberName.Unwrap(),
		},
		Permission: entity.GroupMemberPermission(raw.Permission.Unwrap()),
		GroupLevel: func() uint32 {
			if raw.Level == nil {
				return 0
			}
			return raw.Level.Level.Unwrap()
		}(),
		MemberCard:   raw.MemberCard.MemberCard.Unwrap(),
		SpecialTitle: raw.SpecialTitle.Unwrap(),
		JoinTime:     raw.JoinTimestamp.Unwrap(),
		LastMsgTime:  raw.LastMsgTimestamp.Unwrap(),
		ShutUpTime:   raw.ShutUpTimestamp.Unwrap(),
	}, nil
}

func BuildFetchGroupMembersPacket(gin uint64, cookie []byte) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0xFE7, 3, &operation.FetchGroupMembersRequest{
		GroupUin: proto.Some(int64(gin)),
		Field2:   proto.Some[uint32](5),
		Field3:   proto.Some[uint32](2),
		Body: &operation.FetchGroupMembersRequestBody{
			MemberName:       proto.TRUE,
			MemberCard:       proto.TRUE,
			SpecialTitle:     proto.TRUE,
			Level:            proto.TRUE,
			JoinTimestamp:    proto.TRUE,
			LastMsgTimestamp: proto.TRUE,
			ShutUpTimestamp:  proto.TRUE,
			Permission:       proto.TRUE,
		},
		Cookie: cookie,
	}, false, false)
}

func ParseFetchGroupMembersPacket(data []byte) ([]*entity.GroupMember, []byte, error) {
	rsp, e := ParseOidbPacket[operation.FetchGroupMembersResponse](data)
	if e != nil {
		return nil, nil, e
	}
	//lgrv2.c#
	//group := cache.GetGroupInfo(uint32(rsp.GroupUin.Unwrap()))
	//if group == nil { return nil, nil, exception.NewFormat("invalid target exception: %d", rsp.GroupUin.Unwrap()) }
	var ret []*entity.GroupMember
	for _, raw := range rsp.Members {
		ret = append(ret, &entity.GroupMember{
			User: entity.User{
				Uin:      uint64(raw.Id.Uin.Unwrap()),
				Uid:      raw.Id.Uid.Unwrap(),
				Nickname: raw.MemberName.Unwrap(),
			},
			Permission: entity.GroupMemberPermission(raw.Permission.Unwrap()),
			GroupLevel: func() uint32 {
				if raw.Level == nil {
					return 0
				}
				return raw.Level.Level.Unwrap()
			}(),
			MemberCard:   raw.MemberCard.MemberCard.Unwrap(),
			SpecialTitle: raw.SpecialTitle.Unwrap(),
			JoinTime:     raw.JoinTimestamp.Unwrap(),
			LastMsgTime:  raw.LastMsgTimestamp.Unwrap(),
			ShutUpTime:   raw.ShutUpTimestamp.Unwrap(),
		})
	}
	return ret, rsp.Cookie, nil
}

func BuildFetchGroupExtraPacket(gin uint64, strange bool) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x88D, utils.Ternary[uint32](strange, 14, 0), &operation.FetchGroupExtraRequest{
		Random: proto.Some(int64(crypto.RandU32())),
		Config: &operation.FetchGroupExtraRequestConfig{
			GroupUin: proto.Some(int64(gin)),
			Flags: &operation.FetchGroupExtraRequestConfigFlags{
				GroupName:             proto.Some(""),
				LatestMessageSequence: proto.TRUE,
			},
		},
	}, false, false)
}

func ParseFetchGroupExtraPacket(data []byte) (*operation.FetchGroupExtraResponseInfoResult, error) {
	rsp, e := ParseOidbPacket[operation.FetchGroupExtraResponse](data)
	if e != nil {
		return nil, e
	}
	return rsp.Info.Result, nil
}

func BuildFetchGroupNotificationsPacket(count, start uint64, isfiltered bool) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x10C0, utils.Ternary[uint32](isfiltered, 2, 1),
		&operation.FetchGroupNotificationsRequest{
			Count:         proto.Some(count),
			StartSequence: proto.Some(start),
		}, false, false)
}

func ParseFetchGroupNotificationsPacket(cache *cache.Cache, Isfiltered bool, data []byte, gins ...uint64) (*entity.GroupSystemMessages, error) {
	rsp, err := ParseOidbPacket[operation.FetchGroupNotificationsResponse](data)
	if err != nil {
		return nil, err
	}
	if rsp.GroupNotifications == nil {
		return nil, err
	}

	var requests entity.GroupSystemMessages
	var operatorUin, inviterUin uint64
	var operatorUid, inviterUid string
	var gin uint64
	for _, r := range rsp.GroupNotifications {
		if gin = uint64(r.Group.GroupUin.Unwrap()); len(gins) > 0 && gins[0] != gin {
			continue
		}
		if r.Operator == nil {
			operatorUid, operatorUin = "", 0
		} else {
			operatorUid = r.Operator.Uid.Unwrap()
			operatorUin = cache.GetUin(operatorUid)
		}
		if r.Inviter == nil {
			inviterUid, inviterUin = "", 0
		} else {
			inviterUid = r.Inviter.Uid.Unwrap()
			inviterUin = cache.GetUin(inviterUid)
		}

		base := entity.GroupNoticeBase{
			GroupUin:  gin,
			Sequence:  r.Sequence.Unwrap(),
			State:     entity.EventState(r.State.Unwrap()),
			EventType: entity.EventType(r.Type.Unwrap()),
			TargetUid: r.Target.Uid.Unwrap(),
		}
		base.TargetUin = cache.GetUin(base.TargetUid)
		base.Checked = base.State != entity.Unprocessed

		switch base.EventType {
		case entity.UserJoinRequest:
			requests.JoinRequests = append(requests.JoinRequests, &entity.UserJoinGroupRequest{
				GroupNoticeBase: base,
				OperatorUin:     operatorUin,
				OperatorUid:     operatorUid,
				Comment:         r.Comment.Unwrap(),
				IsFiltered:      Isfiltered,
			})
		case entity.UserInvited, entity.GroupInvited:
			requests.InvitedRequests = append(requests.InvitedRequests, &entity.GroupInvitedRequest{
				GroupNoticeBase: base,
				InvitorUin:      inviterUin,
				InvitorUid:      inviterUid,
				IsFiltered:      Isfiltered,
			})
		case entity.KickedToAdmin, entity.KickedToUser:
			requests.GroupKicks = append(requests.GroupKicks, &entity.GroupKickNotice{
				GroupNoticeBase: base,
				OperatorUin:     operatorUin,
				OperatorUid:     operatorUid,
			})
		case entity.AssignedAsAdmin:
			requests.GroupSetAdmins = append(requests.GroupSetAdmins, &entity.GroupSetAdminNotice{
				GroupNoticeBase: base,
				OperatorUin:     operatorUin,
				OperatorUid:     operatorUid,
			})
		case entity.ExitToAdmin:
			requests.GroupExits = append(requests.GroupExits, &entity.GroupExitNotice{GroupNoticeBase: base})
		case entity.RemoveAdminToUser, entity.RemoveAdminToAdmin:
			requests.GroupUnsetAdmins = append(requests.GroupUnsetAdmins, &entity.GroupUnsetAdminNotice{
				GroupNoticeBase: base,
				OperatorUin:     operatorUin,
				OperatorUid:     operatorUid,
			})
		}
	}
	return &requests, nil
}
