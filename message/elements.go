package message

// from https://github.com/Mrs4s/MiraiGo/blob/master/message/elements.go

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	_ "embed"

	"github.com/tidwall/gjson"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/audio"
	"github.com/kernel-ai/koscore/utils/crypto"
)

//go:embed default_thumb.jpg
var DefaultThumb []byte

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
		ReplySeq  uint32
		SenderUin uint64
		SenderUid string
		GroupUin  uint64 // 私聊回复群聊时
		Time      uint32
		// v2
		SrcUid      uint64
		SrcSequence uint64 // ReplySeq
		Elements    []IMessageElement
		SourceUin   int64 // only for storage, not used in protocol
	}

	VoiceElement struct {
		Name string
		UUid string
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
		FileUUid string // only in new protocol photo
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
		FileUUid string // private
		FileHash string

		// send
		FileStream io.ReadSeeker
		FileSha1   []byte
	}

	ShortVideoElement struct {
		Name     string
		UUid     string
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

const AtTypeGroupMember = 0 // At群成员

func NewText(s string) *TextElement {
	return &TextElement{Content: s}
}

func NewAt(target uint64, display ...string) *AtElement {
	dis := "@" + strconv.FormatInt(int64(target), 10)
	if target == 0 {
		dis = "@全体成员"
	}
	if len(display) != 0 {
		dis = display[0]
	}
	return &AtElement{
		TargetUin: target,
		Display:   dis,
	}
}

/*
func NewGroupReply(m *GroupMessage) *ReplyElement {
	return &ReplyElement{
		ReplySeq : m.Id,
		SenderUin: m.Sender.Uin,
		Time     : m.Time,
		Elements : m.Elements,
	}
}

func NewPrivateReply(m *PrivateMessage) *ReplyElement {
	return &ReplyElement{
		ReplySeq : m.Id,
		SenderUin: m.Sender.Uin,
		Time     : m.Time,
		Elements : m.Elements,
	}
}
*/

func NewRecord(data []byte, summary ...string) *VoiceElement {
	return NewStreamRecord(bytes.NewReader(data), summary...)
}

func NewStreamRecord(r io.ReadSeeker, summary ...string) *VoiceElement {
	md5, sha1, length := crypto.ComputeMd5AndSha1AndLength(r)
	return &VoiceElement{
		Stream: r,
		Md5:    md5,
		Sha1:   sha1,
		Size:   uint32(length),
		Summary: func() string {
			if len(summary) == 0 {
				return ""
			}
			return summary[0]
		}(),
		Duration: func() uint32 {
			if info, err := audio.Decode(r); err == nil {
				return uint32(info.Time)
			}
			return uint32(length)
		}(),
	}
}

func NewFileRecord(path string, summary ...string) (*VoiceElement, error) {
	voice, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamRecord(voice, summary...), nil
}

func NewImage(data []byte, summary ...string) *ImageElement {
	return NewStreamImage(bytes.NewReader(data), summary...)
}

func NewStreamImage(r io.ReadSeeker, summary ...string) *ImageElement {
	md5, sha1, length := crypto.ComputeMd5AndSha1AndLength(r)
	return &ImageElement{
		Stream: r,
		Md5:    md5,
		Sha1:   sha1,
		Size:   uint32(length),
		Summary: func() string {
			if len(summary) == 0 {
				return ""
			}
			return summary[0]
		}(),
	}
}

func NewFileImage(path string, summary ...string) (*ImageElement, error) {
	img, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamImage(img, summary...), nil
}

func NewVideo(data, thumb []byte, summary ...string) *ShortVideoElement {
	return NewStreamVideo(bytes.NewReader(data), bytes.NewReader(thumb), summary...)
}

func NewStreamVideo(r io.ReadSeeker, thumb io.ReadSeeker, summary ...string) *ShortVideoElement {
	md5, sha1, length := crypto.ComputeMd5AndSha1AndLength(r)
	return &ShortVideoElement{
		Md5:    md5,
		Sha1:   sha1,
		Stream: r,
		Compat: &message.VideoFile{},
		Size:   uint32(length),
		Thumb:  NewVideoThumb(thumb),
		Summary: func() string {
			if len(summary) == 0 {
				return ""
			}
			return summary[0]
		}(),
	}
}

func NewFileVideo(path string, thumb []byte, summary ...string) (*ShortVideoElement, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamVideo(file, bytes.NewReader(thumb), summary...), nil
}

func NewVideoThumb(r io.ReadSeeker) *VideoThumb {
	width, height := uint32(1920), uint32(1080)
	md5, sha1, size := crypto.ComputeMd5AndSha1AndLength(r)
	_, imgSize, err := utils.ImageResolve(r)
	if err == nil {
		width, height = uint32(imgSize.Width), uint32(imgSize.Height)
	}
	return &VideoThumb{
		Stream: r,
		Size:   uint32(size),
		Md5:    md5,
		Sha1:   sha1,
		Width:  width,
		Height: height,
	}
}

func NewFile(data []byte, fileName string) *FileElement {
	return NewStreamFile(bytes.NewReader(data), fileName)
}

func NewStreamFile(r io.ReadSeeker, fileName string) *FileElement {
	md5, sha1, length := crypto.ComputeMd5AndSha1AndLength(r)
	return &FileElement{
		FileName:   fileName,
		FileSize:   length,
		FileStream: r,
		FileMd5:    md5,
		FileSha1:   sha1,
	}
}

func NewLocalFile(path string, name ...string) (*FileElement, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamFile(file,
		utils.LazyTernary(len(name) == 0,
			func() string { return filepath.Base(file.Name()) },
			func() string { return name[0] },
		)), nil
}

func NewLightApp(content string) *LightAppElement {
	return &LightAppElement{
		AppName: gjson.Get(content, "app").Str,
		Content: content,
	}
}

func NewXML(content string) *XMLElement { return NewXMLWithId(35, content) }
func NewXMLWithId(id int, content string) *XMLElement {
	return &XMLElement{ServiceId: id, Content: content}
}

func NewForward(resid string, nodes []*ForwardNode) *ForwardMessage {
	return &ForwardMessage{ResId: resid, Nodes: nodes}
}
func NewForwardWithResId(resid string) *ForwardMessage         { return NewForward(resid, nil) }
func NewForwardWithNodes(nodes []*ForwardNode) *ForwardMessage { return NewForward("", nodes) }

func NewFace(id uint32) *FaceElement {
	return &FaceElement{FaceId: id}
}

// key: FetchMarketFaceKey(emojiId) 获取的值
func NewMarketFace(emojiPackId uint32, emojiId []byte, key, summary, value string) *MarketFaceElement {
	return &MarketFaceElement{
		Summary:    summary,
		ItemType:   6,
		FaceId:     emojiId,
		TabId:      emojiPackId,
		SubType:    3,
		EncryptKey: utils.S2B(key),
		MediaType:  0,
		MagicValue: value,
	}
}

func (m *MarketFaceElement) FaceIdString() string {
	if m.MediaType == 2 {
		return utils.B2S(m.FaceId)
	}
	return fmt.Sprintf("%x", m.FaceId)
}

func NewDice(value uint32) *FaceElement {
	if value > 6 {
		value = crypto.RandU32()%3 + 1
	}
	return &FaceElement{
		FaceId:      358,
		ResultId:    value,
		isLargeFace: true,
	}
}

type FingerGuessingType uint32

const (
	FingerGuessingRock     FingerGuessingType = 3 // 石头
	FingerGuessingScissors FingerGuessingType = 2 // 剪刀
	FingerGuessingPaper    FingerGuessingType = 1 // 布
)

func (m FingerGuessingType) String() string {
	switch m {
	case FingerGuessingRock:
		return "石头"
	case FingerGuessingScissors:
		return "剪刀"
	case FingerGuessingPaper:
		return "布"
	}
	return fmt.Sprint(int(m))
}

func NewFingerGuessing(value FingerGuessingType) *FaceElement {
	return &FaceElement{
		FaceId:      359,
		ResultId:    uint32(value),
		isLargeFace: true,
	}
}

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
