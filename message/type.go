package message

import "github.com/kernel-ai/koscore/client/packets/pb/v2/message"

type IMessage interface {
	GetElements() []IMessageElement
	Chat() int64
	ToString() string
	Texts() []string
}

type ElementType int

const (
	Text       ElementType = iota // 文本
	Image                         // 图片
	Face                          // 表情
	At                            // 艾特
	Reply                         // 回复
	Service                       // 服务
	Forward                       // 转发
	File                          // 文件
	Voice                         // 语音
	Video                         // 视频
	LightApp                      // 轻应用
	RedBag                        // 红包
	MarketFace                    // 魔法表情
)

type (
	Sender struct {
		Uin          uint64
		Uid          string
		Nickname     string
		CardName     string
		IsFriend     bool
		AppId        int32
		OriginalByte []byte
		//Original *message.PushMsgBody
	}

	Message struct {
		Id        uint64
		Random    uint64
		Time      int64
		Sender    Sender
		Target    Sender
		Elements  []IMessageElement
		ClientSeq uint64
		MsgUid    uint64
	}

	PrivateMessage struct{ *Message }

	TempMessage struct{ *Message }

	GroupMessage struct {
		*Message
		GroupUin  uint64
		GroupName string
	}

	// SendingMessage struct { Elements []IMessageElement }

	IMessageElement interface{ Type() ElementType }
)

// build

type MsgContentBuilder interface {
	BuildContent() []byte
}

type ElementBuilder interface {
	BuildElement() []*message.Elem
}
