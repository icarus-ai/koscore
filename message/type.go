package message

import (
	"io"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
)

type (
	TextElement struct {
		Content string
	}

	AtElement struct {
		TargetUin uint64
		TargetUid string
		Display   string
		SubType   AtType
	}

	FaceElement struct {
		FaceId      uint32
		ResultId    uint32 // 猜拳和骰子的值
		isLargeFace bool
	}

	ReplyElement struct {
		SenderUin uint64
		SenderUid string
		GroupUin  uint64 // 私聊回复群聊时
		Time      uint32
		// v2
		SrcUid    uint64
		ReplySeq  uint64 // SrcSequence
		Elements  []IMessageElement
		SourceUin int64 // only for storage, not used in protocol
	}

	VoiceElement struct {
		Name string
		Uuid string
		Size uint32
		URL  string
		Md5  []byte
		Sha1 []byte
		Node *oidb.IndexNode

		// --- sending ---
		MsgInfo  *oidb.MsgInfo
		Compat   []byte
		Duration uint32
		Stream   io.ReadSeeker
		Summary  string

		IsGroup bool
	}

	ImageElement struct {
		ImageId  string
		FileUuid string // only in new protocol photo
		Size     uint32
		Width    uint32
		Height   uint32
		URL      string
		SubType  int32

		// EffectId show pic effect id.
		EffectId int32 // deprecated
		Flash    bool

		// send & receive
		Summary string
		Md5     []byte // only in old protocol photo
		IsGroup bool

		Sha1        []byte
		MsgInfo     *oidb.MsgInfo
		Stream      io.ReadSeeker
		CompatFace  *message.CustomFace     // GroupImage
		CompatImage *message.NotOnlineImage // FriendImage
	}

	FileElement struct {
		FileSize uint64
		FileName string
		FileMd5  []byte
		FileURL  string
		FileId   string // group
		FileUuid string // private
		FileHash string

		// send
		FileStream io.ReadSeeker
		FileSha1   []byte
	}

	ShortVideoElement struct {
		Name     string
		Uuid     string
		Size     uint32
		URL      string
		Duration uint32
		Node     *oidb.IndexNode

		// send
		Thumb   *VideoThumb
		Summary string
		Md5     []byte
		Sha1    []byte
		Stream  io.ReadSeeker
		MsgInfo *oidb.MsgInfo
		Compat  *message.VideoFile

		IsGroup bool
	}

	VideoThumb struct {
		Stream io.ReadSeeker
		Size   uint32
		Md5    []byte
		Sha1   []byte
		Width  uint32
		Height uint32
	}

	LightAppElement struct {
		AppName string
		Content string
	}

	XMLElement struct {
		ServiceId int
		Content   string
	}

	ForwardMessage struct {
		IsGroup bool
		SelfId  uint64
		ResId   string
		Nodes   []*ForwardNode
	}

	MarketFaceElement struct {
		Summary    string
		ItemType   uint32
		FaceInfo   uint32
		FaceId     []byte // decoded = mediaType == 2 ? string(FaceId) : hex.EncodeToString(FaceId).toLower().trimSpace(); download url param?
		TabId      uint32
		SubType    uint32 // image type, 0 -> None 1 -> Magic Face 2 -> GIF 3 -> PNG
		EncryptKey []byte // tea + xor, see EMosmUtils.class::a maybe useful?
		MediaType  uint32 // 1 -> Voice Face 2 -> dynamic face
		MagicValue string
	}

	AtType int
)

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

func (m *TextElement) Type() ElementType       { return Text }
func (m *AtElement) Type() ElementType         { return At }
func (m *FaceElement) Type() ElementType       { return Face }
func (m *ReplyElement) Type() ElementType      { return Reply }
func (m *VoiceElement) Type() ElementType      { return Voice }
func (m *ImageElement) Type() ElementType      { return Image }
func (m *FileElement) Type() ElementType       { return File }
func (m *ShortVideoElement) Type() ElementType { return Video }
func (m *LightAppElement) Type() ElementType   { return LightApp }
func (m *XMLElement) Type() ElementType        { return Service }
func (m *ForwardMessage) Type() ElementType    { return Forward }
func (m *MarketFaceElement) Type() ElementType { return MarketFace }

func (m ElementType) String() string {
	switch m {
	case Text:
		return "文本"
	case Image:
		return "图片"
	case Face:
		return "表情"
	case At:
		return "AT"
	case Reply:
		return "回复"
	case Service:
		return "XML"
	case Forward:
		return "合并"
	case File:
		return "文件"
	case Voice:
		return "语音"
	case Video:
		return "视频"
	case LightApp:
		return "JSON"
	case RedBag:
		return "红包"
	case MarketFace:
		return "魔法表情"
	}
	return "unknown"
}

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

type IMessage interface {
	GetElements() []IMessageElement
	Chat() int64
	ToString() string
	Texts() []string
}

func (msg *GroupMessage) ToString() string               { return ToReadableString(msg.Elements) }
func (msg *GroupMessage) GetElements() []IMessageElement { return msg.Elements }
func (msg *GroupMessage) Chat() int64                    { return int64(msg.Id) }
func (msg *GroupMessage) Texts() []string {
	texts := make([]string, 0, len(msg.Elements))
	for _, elem := range msg.Elements {
		texts = append(texts, ToReadableStringEle(elem))
	}
	return texts
}

func (msg *PrivateMessage) ToString() string               { return ToReadableString(msg.Elements) }
func (msg *PrivateMessage) GetElements() []IMessageElement { return msg.Elements }
func (msg *PrivateMessage) Chat() int64                    { return int64(msg.Id) }
func (msg *PrivateMessage) Texts() []string {
	texts := make([]string, 0, len(msg.Elements))
	for _, elem := range msg.Elements {
		texts = append(texts, ToReadableStringEle(elem))
	}
	return texts
}

func (msg *TempMessage) ToString() string               { return ToReadableString(msg.Elements) }
func (msg *TempMessage) GetElements() []IMessageElement { return msg.Elements }
func (msg *TempMessage) Chat() int64                    { return int64(msg.Id) }
func (msg *TempMessage) Texts() []string {
	texts := make([]string, 0, len(msg.Elements))
	for _, elem := range msg.Elements {
		texts = append(texts, ToReadableStringEle(elem))
	}
	return texts
}

// build

type MsgContentBuilder interface {
	BuildContent() []byte
}

type ElementBuilder interface {
	BuildElement() []*message.Elem
}

func ElementsHasType(elems []IMessageElement, t ElementType) bool {
	for _, elem := range elems {
		if elem.Type() == t {
			return true
		}
	}
	return false
}

func PackElementsToBody(elems []IMessageElement) (body *message.MessageBody) {
	body = &message.MessageBody{
		RichText: &message.RichText{Elems: PackElements(elems)},
	}
	for _, elem := range elems {
		if bd, ok := elem.(MsgContentBuilder); ok {
			body.MsgContent = bd.BuildContent()
		}
	}
	return
}

func PackElements(elems []IMessageElement) []*message.Elem {
	if len(elems) == 0 {
		return nil
	}
	ret := make([]*message.Elem, 0, len(elems))
	for _, elem := range elems {
		if bd, ok := elem.(ElementBuilder); ok {
			ret = append(ret, bd.BuildElement()...)
		}
	}
	return ret
}
