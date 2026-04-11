package entity

type (
	EventState          uint32
	EventType           uint32
	GroupRequestOperate uint8 // 处理加群请求
)

const (
	GroupRequestOperateAllow  GroupRequestOperate = 1 // 同意
	GroupRequestOperateDeny   GroupRequestOperate = 2 // 拒绝
	GroupRequestOperateIgnore GroupRequestOperate = 3 // 忽略
)

const (
	/*
		1:未处理
		2:已同意
		3:已拒绝
		4:已忽略
		5:已处理
	*/
	NoNeed EventState = iota
	Unprocessed
	Processed
)

/*
1: "用户昵称"申请加入"群名称"群
2: "用户昵称"邀请您加入"群名称"群
3: 成为管理员, 通知全员
6: UIN被T出群, 通知管理员
7: UIN被T出群, 通知UIN
10: 您的好友“用户昵称”拒绝加入“群名称”群。拒绝理由：XXX
11: "群名称"群管理员“用户昵称”拒绝了您的加群请求。拒绝理由：XXX
12:您的好友“用户昵称”已经同意加入“群名称”群。
13:UIN退群，通知管理员。
15:管理员身份被取消，通知被取消人
16：管理员身份被取消，通知其他管理员
20:(同2,已废弃): 您的好友“用户昵称”邀请您加入“群名称”群。
21:(同12,已废弃): 您的好友“用户昵称”已经同意加入“群名称”群。附加信息：正在等待管理员验证。
22:“用户昵称”申请加入“群名称”群。附加信息：来自群成员XXX的邀请。
23:(同10，已废弃): 您的好友“用户昵称”拒绝加入“群名称”群。拒绝理由：XXX。
35:群“群名称”管理员已同意您的加群申请
*/
const (
	UserJoinRequest    EventType = 1  // 用户申请加群
	GroupInvited       EventType = 2  // 被邀请加群
	AssignedAsAdmin    EventType = 3  // 被设置为管理员
	KickedToAdmin      EventType = 6  // 被踢出群聊 通知管理员
	KickedToUser       EventType = 7  // 被踢出群聊 通知UIN
	ExitToAdmin        EventType = 13 // UIN退群，通知管理员
	RemoveAdminToUser  EventType = 15 // 被取消管理员 通知被取消人
	RemoveAdminToAdmin EventType = 16 // 被取消管理员 通知其他管理员
	UserInvited        EventType = 22 // 群员邀请其他人
)

type Group struct {
	GroupUin        uint64
	GroupName       string
	GroupOwner      uint32
	GroupCreateTime uint32
	//GroupMemo       string
	GroupLevel  uint32
	MemberCount uint32
	MaxMember   uint32
	LastMsgSeq  uint32

	Description  string
	Question     string
	Announcement string
}

type (
	GroupNoticeBase struct {
		GroupUin   uint64
		GroupName  string
		Sequence   uint64
		State      EventState
		EventType  EventType
		TargetUin  uint64
		TargetUid  string
		TargetNick string
		Checked    bool
	}

	UserJoinGroupRequest struct {
		GroupNoticeBase
		InvitorUin  uint64 `json:"invitor_uin"`
		InvitorUid  string `json:"-"`
		InvitorNick string `json:"-"`
		OperatorUin uint64 `json:"actor"`
		OperatorUid string `json:"-"`
		Comment     string `json:"message"`
		IsFiltered  bool   `json:"is_filtered"`
	}

	GroupInvitedRequest struct {
		GroupNoticeBase
		InvitorUin  uint64 `json:"invitor_uin"`
		InvitorUid  string `json:"-"`
		InvitorNick string `json:"-"`
		IsFiltered  bool   `json:"is_filtered"`
	}

	GroupSetAdminNotice struct {
		GroupNoticeBase
		OperatorUin uint64
		OperatorUid string
	}

	GroupKickNotice struct {
		GroupNoticeBase
		OperatorUin uint64
		OperatorUid string
	}

	GroupExitNotice struct{ GroupNoticeBase }

	GroupUnsetAdminNotice struct {
		GroupNoticeBase
		OperatorUin uint64
		OperatorUid string
	}

	GroupSystemMessages struct {
		InvitedRequests  []*GroupInvitedRequest  `json:"invited_requests"`
		JoinRequests     []*UserJoinGroupRequest `json:"join_requests"`
		GroupSetAdmins   []*GroupSetAdminNotice
		GroupKicks       []*GroupKickNotice
		GroupExits       []*GroupExitNotice
		GroupUnsetAdmins []*GroupUnsetAdminNotice
	}
)

func (g *Group) Avatar() string { return GroupAvatar(g.GroupUin) }
