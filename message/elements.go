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
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/audio"
	"github.com/kernel-ai/koscore/utils/crypto"
)

//go:embed default_thumb.jpg
var DefaultThumb []byte

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

func NewXML(content string) *XMLElement { return NewXMLWithID(35, content) }
func NewXMLWithID(id int, content string) *XMLElement {
	return &XMLElement{ServiceId: id, Content: content}
}

func NewForward(resid string, nodes []*ForwardNode) *ForwardMessage {
	return &ForwardMessage{ResId: resid, Nodes: nodes}
}
func NewForwardWithResID(resid string) *ForwardMessage         { return NewForward(resid, nil) }
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
