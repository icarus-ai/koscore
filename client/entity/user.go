package entity

import "fmt"

type (
	OnlineStatus struct {
		StatusId uint32
		FaceId   uint32
		Msg      string
	}

	BusinessType uint32

	BusinessCustom struct {
		Type   BusinessType
		Level  uint32
		Icon   string
		IsYear bool
		IsPro  bool
	}

	GenderInfo uint8

	UserCategory struct {
		Id     int32
		Name   string
		Count  int32
		SortId int32
	}

	User struct {
		Uin          uint64
		Uid          string
		Nickname     string
		Remarks      string
		PersonalSign string
		Avatar       string
		Age          uint32
		Sex          GenderInfo // 1男 2女 255不可见
		Level        uint32
		Source       string // 好友来源

		QID string

		Country string
		City    string
		School  string

		VipLevel uint32

		Registration int64
		Birthday     int64
		Status       OnlineStatus
		Business     []BusinessCustom

		Category *UserCategory
	}
)

func (m GenderInfo) String() string {
	switch m {
	case 0:
		return "⚧"
	case 1:
		return "♂"
	case 2:
		return "♀"
	case 255:
		return "?"
	default:
		return ""
	}
}

func (m BusinessCustom) String() string {
	v := m.Type.String()
	if m.IsYear {
		v = "年费" + v
	}
	if m.IsPro {
		v = "超级" + v
	}
	return fmt.Sprintf("%s lv%d", v, m.Level)
}

func (m OnlineStatus) String() string {
	if m.StatusId == 13633281 {
		return fmt.Sprintf("自定义状态: %s", m.Msg)
	}
	if v, ok := k_status_map[m.StatusId]; ok {
		return v + " " + m.Msg
	}
	return fmt.Sprintf("%d %s", m.StatusId, m.Msg)
}

func (m BusinessType) String() string {
	switch m {
	case 1:
		return "QQ会员"
	case 4:
		return "腾讯视频"
	case 101:
		return "红钻"
	case 102:
		return "黄钻"
	case 103:
		return "绿钻"
	case 104:
		return "情侣个性钻"
	case 105:
		return "微云会员"
	case 107:
		return "SVIP+腾讯视频"
	case 108:
		return "大王超级会员"
	case 113:
		return "QQ大会员"
	case 115:
		return "cf游戏特权"
	case 117:
		return "QQ集卡"
	case 118:
		return "蓝钻"
	case 119:
		return "情侣会员"
	default:
		return "未知"
	}
}

var k_status_map = map[uint32]string{
	1: "在线",
	3: "离开",
	4: "隐身/离线",
	5: "忙碌",
	6: "Q我吧",
	7: "请勿打扰",

	263169: "听歌中",
	394241: "今日天气",
	197633: "timi中",
	525313: "熬夜中",

	1770497: "恋爱中",
	3081217: "好运锦鲤",
	2098177: "嗨到飞起",
	2229249: "元气满满",
	2556929: "一言难尽",
	7931137: "emo中",
	2491393: "我太难了",
	1836033: "我没事",
	2425857: "想静静",
	2294785: "悠哉哉",
	1312001: "摸鱼中",
	2360321: "无聊中",

	11600897: "水逆退散",
	13698817: "难得糊涂",
	14485249: "我想开了",
	15926017: "信号弱",
	16253697: "睡觉中",
	14419713: "肝作业",
	16384769: "学习中",
	15140609: "搬砖中",
	15205121: "我的电量",
	15271681: "一起元梦",
	15337217: "求星搭子",
	16581377: "追剧中",
	16713473: "做好事",
	13829889: "出去浪",
	14616321: "去旅行",
	14550785: "被掏空",
	14747393: "今日步数",
	14878465: "我crush了",
	14026497: "爱你",

	13633281: "自定义状态",
}
