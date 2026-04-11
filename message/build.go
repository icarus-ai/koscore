package message

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/proto"
)

/*
import (
	"encoding/json"
	"fmt"
)
*/

func (m *TextElement) BuildElement() []*message.Elem {
	return []*message.Elem{{Text: &message.Text{TextMsg: proto.Some(m.Content)}}}
}

func (m *AtElement) BuildElement() []*message.Elem {
	reserveData, _ := proto.Marshal(&message.TextResvAttr{
		AtType:         proto.Some(utils.Ternary[uint32](m.TargetUin == 0, 1, 2)), // 1 for mention all
		AtMemberUin:    proto.Some[uint64](m.TargetUin),
		AtMemberTinyid: proto.Some[uint64](0),
		AtMemberUid:    proto.Some(m.TargetUid),
	})
	return []*message.Elem{{
		Text: &message.Text{
			TextMsg:   proto.Some(m.Display),
			PbReserve: reserveData,
		}}}
}

func (m *ReplyElement) BuildElement() []*message.Elem {
	reserveData, err := proto.Marshal(&message.SourceMsgResvAttr{
		OriMsgType:  proto.Some[uint32](2),
		SourceMsgId: proto.Some(m.SrcUid),
		SenderUid:   proto.Some(m.SenderUid),
		//ReceiverUid: proto.Some(m.SenderUid),
	})
	if err != nil {
		return nil
	}
	return []*message.Elem{{
		SrcMsg: &message.SourceMsg{
			OrigSeqs:  []uint64{uint64(m.ReplySeq)},
			SenderUin: proto.Some(m.SenderUin),
			Flag:      proto.Some[uint32](0), // intentional, force the client to fetch the original message
			Time:      proto.Some(m.Time),
			Elems:     PackElements(m.Elements),
			PbReserve: reserveData,
			//ToUin    : proto.Some[uint64](0),
		}}}
}

// GroupFileEntity
func (m *FileElement) BuildContent() []byte {
	extra, _ := proto.Marshal(&message.FileExtra{
		File: &message.NotOnlineFile{
			FileType:       proto.Some[uint32](0),
			FileUuid:       proto.Some(m.FileUUid),
			FileMd5:        m.FileMd5,
			FileName:       proto.Some(m.FileName),
			FileSize:       proto.Some(m.FileSize),
			SubCmd:         proto.Some[uint32](1),
			DangerLevel:    proto.Some[uint32](0),
			ExpireTime:     proto.Some(uint32(time.Now().Add(time.Hour * 24 * 7).Unix())),
			FileIdCrcMedia: proto.Some(m.FileHash),
		},
	})
	return extra
}

func (m *ImageElement) BuildElement() []*message.Elem {
	common, err := proto.Marshal(m.MsgInfo)
	if err != nil {
		return nil
	}
	msg := []*message.Elem{{}, {
		CommonElem: &message.CommonElem{
			ServiceType:  proto.Some[uint32](48),
			PbElem:       common,
			BusinessType: proto.Some(utils.Ternary[uint32](m.IsGroup, 20, 10)),
		}}}
	if m.CompatFace != nil {
		msg[0].CustomFace = m.CompatFace
	}
	if m.CompatImage != nil {
		msg[0].NotOnlineImage = m.CompatImage
	}
	return msg
}

func (m *ShortVideoElement) BuildElement() (ret []*message.Elem) {
	common, err := proto.Marshal(m.MsgInfo)
	if err != nil {
		return nil
	}
	if m.Compat != nil {
		ret = append(ret, &message.Elem{VideoFile: m.Compat})
	}
	return append(ret, &message.Elem{
		CommonElem: &message.CommonElem{
			ServiceType:  proto.Some[uint32](48),
			PbElem:       common,
			BusinessType: proto.Some(utils.Ternary[uint32](m.IsGroup, 21, 11)),
		},
	})
}

func (m *VoiceElement) BuildElement() []*message.Elem {
	common, err := proto.Marshal(m.MsgInfo)
	if err != nil {
		return nil
	}
	return []*message.Elem{{
		CommonElem: &message.CommonElem{
			ServiceType:  proto.Some[uint32](48),
			PbElem:       common,
			BusinessType: proto.Some(utils.Ternary[uint32](m.IsGroup, 22, 12)),
		}}}
}

func (m *LightAppElement) BuildElement() []*message.Elem {
	return []*message.Elem{{
		LightAppElem: &message.LightAppElem{
			BytesData: append([]byte{0x01}, binary.ZlibCompress([]byte(m.Content))...),
		}}}
}

func (m *ForwardMessage) BuildElement() []*message.Elem {
	var news []News
	var metaSource string
	nodes_size := len(m.Nodes)
	if nodes_size == 0 {
		news = []News{{Text: "转发消息"}}
		metaSource = "聊天记录"
	} else {
		news = make([]News, nodes_size)
		for i, node := range m.Nodes {
			news[i] = News{Text: fmt.Sprintf("%s: %s", node.SenderName, ToReadableString(node.Message))}
		}

		isSenderNameExist := make(map[string]bool)
		isContainSelf := false
		isCount := 0
		for _, v := range m.Nodes {
			if v.SenderId == m.SelfId && m.SelfId > 0 {
				isContainSelf = true
			}
			if _, ok := isSenderNameExist[v.SenderName]; !ok {
				isCount++
				isSenderNameExist[v.SenderName] = true
				if metaSource == "" {
					metaSource = v.SenderName
				} else {
					metaSource += fmt.Sprintf("和%s", v.SenderName)
				}
			}
		}
		if !isContainSelf || (isCount > 2 && isCount < 1) {
			metaSource = "群聊的聊天记录"
		} else {
			metaSource += "的聊天记录"
		}
	}

	guid := utils.NewUUID()
	data, _ := json.Marshal(&MultiMsgLightAppExtra{
		FileName: guid,
		Sum:      nodes_size,
	})

	data, _ = json.Marshal(&MultiMsgLightApp{
		App:    "com.tencent.multimsg",
		Desc:   "[聊天记录]",
		Prompt: "[聊天记录]",
		Ver:    "0.0.0.5",
		View:   "contact",
		Extra:  utils.B2S(data),
		Config: Config{
			Autosize: 1,
			Forward:  1,
			Round:    1,
			Type:     "normal",
			Width:    300,
		},
		Meta: Meta{Detail: Detail{
			News:    news,
			Resid:   m.ResId,
			Source:  metaSource,
			Summary: fmt.Sprintf("查看%d条转发消息", nodes_size),
			UniSeq:  guid,
		}},
	})
	return NewLightApp(utils.B2S(data)).BuildElement()
}

/*
func (e *FaceElement) BuildElement() []*message.Elem {
	if e.isLargeFace {
		business, resultid, name := int32(1), "", ""
		if        e.FaceId == 358 { business, resultid, name = 2, fmt.Sprint(e.ResultId), "/骰子"
		} else if e.FaceId == 359 { business, resultid, name = 2, fmt.Sprint(e.ResultId), "/包剪锤" }
		qFaceData, _ := proto.Marshal(&message.QFaceExtra{
			PackId     : proto.Some("1"),
			StickerId  : proto.Some("8"),
			Qsid       : proto.Some(int32(e.FaceId)),
			SourceType : proto.Some[int32](1),
			StickerType: proto.Some(business),
			ResultId   : proto.Some(resultid),
			Text       : proto.Some(name),
			RandomType : proto.Some[int32](1),
		})
		return []*message.Elem{{
			CommonElem: &message.CommonElem{
				ServiceType : 37,
				PbElem      : qFaceData,
				BusinessType: 1,
		} } }
	}
	return []*message.Elem {
		{ Face: &message.Face{Index: proto.Some(int32(e.FaceId))} },
	}
}

func (e *XMLElement) BuildElement() []*message.Elem {
	return []*message.Elem{{
		RichMsg: &message.RichMsg{
			ServiceId: proto.Some(int32(e.ServiceId)),
			Template1: append([]byte{0x01}, binary.ZlibCompress([]byte(e.Content))...),
	} }}
}

func (e *MarketFaceElement) BuildElement() []*message.Elem {
	reserve, _ := proto.Marshal(&message.MarketFacePbReserve{Field8: 1})
	return []*message.Elem { {
		MarketFace : &message.MarketFace{
		FaceName   : proto.String(e.Summary),
		ItemType   : proto.Uint32(e.ItemType),
		FaceInfo   : proto.Uint32(1),
		FaceId     : e.FaceId,
		TabId      : proto.Uint32(e.TabId),
		SubType    : proto.Uint32(e.SubType),
		Key        : e.EncryptKey,
		MediaType  : proto.Uint32(e.MediaType),
		ImageWidth : proto.Uint32(300),
		ImageHeight: proto.Uint32(300),
		MobileParam: utils.S2B(e.MagicValue),
		PbReserve  : reserve,
	} }, {
		Text       : &message.Text { Str: proto.Some(e.Summary) },
	} }
}
*/
