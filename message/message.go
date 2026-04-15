package message

import (
	"encoding/asn1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kernel-ai/koscore/client/packets/message/message_type"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
	"github.com/kernel-ai/koscore/utils/proto"
)

func parse_message_resolves(content_head *message.ContentHead, routing_head *message.RoutingHead) *Message {
	msg := &Message{
		Id:        content_head.Sequence.Unwrap(),
		Random:    uint64(content_head.Random.Unwrap()),
		ClientSeq: content_head.ClientSequence.Unwrap(),
		Time:      content_head.Time.Unwrap(),
		Sender: Sender{
			Uin:      uint64(routing_head.FromUin.Unwrap()),
			Uid:      routing_head.FromUid.Unwrap(),
			AppId:    routing_head.FromAppId.Unwrap(),
			IsFriend: false,
		},
		Target: Sender{
			Uin: uint64(routing_head.ToUin.Unwrap()),
			Uid: routing_head.ToUid.Unwrap(),
		},
		MsgUid: content_head.MsgUid.Unwrap(), // MsgUid & 0xFFFFFFFF are the same to random
	}

	switch content_head.Type.Unwrap() {
	case message_type.PRIVATE_MESSAGE:
		//msg.Sender.Nickname = "" // 后续通过缓存查询
		msg.Sender.IsFriend = true
	case message_type.GROUP_MESSAGE:
		msg.Sender.Nickname = routing_head.Group.GroupCard.Unwrap()
		//msg.Sender.IsFriend = await _context.CacheContext.ResolveFriend(msg.Sender.Uin)
		msg.Sender.CardName = routing_head.Group.GroupCard.Unwrap()
	case message_type.TEMP_MESSAGE:
		msg.Sender.Nickname = routing_head.CommonC2C.Name.Unwrap()
		//msg.Sender.TinyId = routing_head.CommonC2C.FromTinyId.Unwrap()
		//await _context.CacheContext.ResolveStranger(routingHead.ToUid)).CloneWithSource(routingHead.CommonC2C.FromTinyId);
	default: // throw new NotImplementedException();
	}
	return msg
}

type (
	innerSequence struct {
		A       int64
		B       int64
		BotUin  int64
		FileKey []byte
		TimeUTC int64
		C       []byte
	}
	outerSequence struct {
		Version int
		Inner   *innerSequence
		Empty   []byte
	}
)

func parse_ptt_rich_text(uin uint64, body *message.MessageBody, isGroup bool) (res []IMessageElement) {
	if body.RichText.Ptt != nil {
		rich := body.RichText
		inner := &innerSequence{
			A: 1, B: 0,
			BotUin:  int64(uin),
			FileKey: nil,
			TimeUTC: time.Now().UTC().Unix(),
			C: binary.NewBuilder().
				WriteLengthString("filetype", prefix.Int32|prefix.LengthOnly).
				WriteLengthString("0", prefix.Int32|prefix.LengthOnly).
				WriteLengthString("codec", prefix.Int32|prefix.LengthOnly).
				WriteLengthString("1", prefix.Int32|prefix.LengthOnly).
				ToBytes(),
		}
		if uin >= 0x80000000 {
			inner.BotUin -= 0x100000000
		}
		if len(rich.Ptt.GroupFileKey) > 0 {
			inner.FileKey = rich.Ptt.GroupFileKey
		} else if len(rich.Ptt.FileUuid) > 0 {
			inner.FileKey = rich.Ptt.FileUuid
		}

		data, e := asn1.Marshal(&outerSequence{
			Version: 1,
			Inner:   inner,
			Empty:   binary.Empty,
		})
		if e == nil {
			v := &VoiceElement{
				URL:      fmt.Sprintf("https://grouptalk.c2c.qq.com/?ver=2&rkey=%X&voice_codec=1&filetype=0", data),
				Name:     rich.Ptt.FileName.Unwrap(),
				Size:     rich.Ptt.FileSize.Unwrap(),
				Md5:      rich.Ptt.FileMd5,
				Uuid:     string(rich.Ptt.FileUuid),
				Duration: rich.Ptt.Time.Unwrap(),
			}
			if isGroup {
				if rich.Ptt.FileId.Unwrap() != 0 {
					v.Node = &oidb.IndexNode{FileUuid: rich.Ptt.FileKey}
				}
			} else {
				v.Node = &oidb.IndexNode{FileUuid: proto.Some(string(rich.Ptt.FileUuid))}
			}
			res = append(res, v)
		}
	}

	if body.MsgContent != nil {
		extra, err := proto.Unmarshal[message.FileExtra](body.MsgContent)
		if err != nil {
			return res
		}
		if extra.File.FileSize.IsSome() && extra.File.FileName.IsSome() && extra.File.FileMd5 != nil && extra.File.FileUuid.IsSome() && extra.File.FileIdCrcMedia.IsSome() {
			res = append(res, &FileElement{
				FileSize: extra.File.FileSize.Unwrap(),
				FileName: extra.File.FileName.Unwrap(),
				FileMd5:  extra.File.FileMd5,
				FileUuid: extra.File.FileUuid.Unwrap(),
				FileHash: extra.File.FileIdCrcMedia.Unwrap(),
			})
		}
	}
	return
}

func ParseGroupMessage(bot_uin uint64, msg *message.CommonMessage) *GroupMessage {
	ret := &GroupMessage{
		GroupName: msg.RoutingHead.Group.GroupName.Unwrap(),
		GroupUin:  uint64(msg.RoutingHead.Group.GroupCode.Unwrap()),
		Message:   parse_message_resolves(msg.ContentHead, msg.RoutingHead),
	}
	if msg.MessageBody != nil {
		ret.Elements = ParseMessageElements(msg.MessageBody.RichText.Elems)
		ret.Elements = append(ret.Elements, parse_ptt_rich_text(bot_uin, msg.MessageBody, true)...)
	}
	return ret
}

func ParsePrivateMessage(bot_uin uint64, msg *message.CommonMessage) *PrivateMessage {
	ret := parse_message_resolves(msg.ContentHead, msg.RoutingHead)
	if msg.MessageBody != nil {
		ret.Elements = ParseMessageElements(msg.MessageBody.RichText.Elems)
		ret.Elements = append(ret.Elements, parse_ptt_rich_text(bot_uin, msg.MessageBody, false)...)
	}
	return &PrivateMessage{Message: ret}
}

func ParseTempMessage(bot_uin uint64, msg *message.CommonMessage) *TempMessage {
	ret := parse_message_resolves(msg.ContentHead, msg.RoutingHead)
	if msg.MessageBody != nil {
		ret.Elements = ParseMessageElements(msg.MessageBody.RichText.Elems)
		ret.Elements = append(ret.Elements, parse_ptt_rich_text(bot_uin, msg.MessageBody, false)...)
	}
	return &TempMessage{Message: ret}
}

func ParseMessageElements(msg []*message.Elem) (res []IMessageElement) {
	skipNext := false

	for _, elem := range msg {
		if skipNext {
			skipNext = false
			continue
		}

		// ReplyEntity
		if elem.SrcMsg != nil && len(elem.SrcMsg.OrigSeqs) != 0 {
			resvAttr, e := proto.Unmarshal[message.SourceMsgResvAttr](elem.SrcMsg.PbReserve)
			if e != nil {
				continue
			}
			var elements []*message.Elem
			for _, v := range elem.SrcMsg.Elems {
				//if len(v) > 0 { if _elem, e := proto.Unmarshal[message.Elem](v); e == nil { elements = append(elements, _elem) } }
				elements = append(elements, v)
			}
			res = append(res, &ReplyElement{
				ReplySeq:  elem.SrcMsg.OrigSeqs[0],
				Time:      elem.SrcMsg.Time.Unwrap(),
				SenderUin: elem.SrcMsg.SenderUin.Unwrap(),
				GroupUin:  elem.SrcMsg.ToUin.Unwrap(),
				Elements:  ParseMessageElements(elements),
				SourceUin: int64(elem.SrcMsg.SenderUin.Unwrap()),
				SrcUid:    resvAttr.SourceMsgId.Unwrap(),
			})
		}

		// TextEntity
		if elem.Text != nil {
			if len(elem.Text.Attr6Buf) > 0 {
				att6 := binary.NewReader(elem.Text.Attr6Buf)
				att6.SkipBytes(7)
				at := NewAt(uint64(att6.ReadU32()), elem.Text.TextMsg.Unwrap())
				at.SubType = AtTypeGroupMember
				// v2
				if attr, e := proto.Unmarshal[message.TextResvAttr](elem.Text.PbReserve); e == nil {
					at.TargetUin = attr.AtMemberUin.Unwrap()
					if attr.AtType.Unwrap() == 2 {
						at.TargetUid = attr.AtMemberUid.Unwrap()
					}
				}
				res = append(res, at)
			} else {
				s := elem.Text.TextMsg.Unwrap()
				if strings.Contains(s, "\r") && !strings.Contains(s, "\r\n") {
					s = strings.ReplaceAll(s, "\r", "\r\n")
				}
				res = append(res, NewText(s))
			}
		}

		// ***** HAS_OLD_CODE BIGIN *****
		if elem.Face != nil {
			if len(elem.Face.Old) > 0 {
				if elem.Face.Index.IsSome() {
					res = append(res, &FaceElement{FaceId: uint32(elem.Face.Index.Unwrap())})
				}
			} else if elem.CommonElem != nil && len(elem.CommonElem.PbElem) > 0 {
				switch elem.CommonElem.ServiceType.Unwrap() {
				case 37:
					if qFace, err := proto.Unmarshal[message.QFaceExtra](elem.CommonElem.PbElem); err == nil {
						if qFace.Qsid.IsSome() {
							res = append(res, &FaceElement{FaceId: uint32(qFace.Qsid.Unwrap()), isLargeFace: true})
						}
					}
				case 33:
					if qFace, err := proto.Unmarshal[message.QSmallFaceExtra](elem.CommonElem.PbElem); err == nil {
						res = append(res, &FaceElement{FaceId: qFace.FaceId.Unwrap(), isLargeFace: false})
					}
				}
			}
		}
		// ***** HAS_OLD_CODE END *****

		// ***** HAS_OLD_CODE BIGIN *****
		if elem.VideoFile != nil {
			video := elem.VideoFile
			res = append(res, &ShortVideoElement{
				Name: video.FileName.Unwrap(),
				Uuid: video.FileUuid.Unwrap(),
				Size: uint32(video.FileSize.Unwrap()),
				Md5:  video.FileMd5,
				Node: &oidb.IndexNode{
					Info: &oidb.FileInfo{
						FileName: video.FileName,
						FileSize: proto.Some(uint32(video.FileSize.Unwrap())),
						FileHash: proto.Some(hex.EncodeToString(video.FileMd5)),
					},
					FileUuid: video.FileUuid,
				},
				Thumb: &VideoThumb{
					Size: uint32(elem.VideoFile.ThumbFileSize.Unwrap()),
					Md5:  elem.VideoFile.ThumbFileMd5,
				},
			})
		}
		// ***** HAS_OLD_CODE END *****

		// ***** HAS_OLD_CODE BIGIN *****
		if elem.CustomFace != nil {
			if len(elem.CustomFace.Md5) == 0 {
				continue
			}

			uri := elem.CustomFace.OrigUrl.Unwrap()
			if strings.Contains(uri, "rkey") {
				uri = "https://multimedia.nt.qq.com.cn" + uri
			} else {
				uri = "http://gchat.qpic.cn" + uri
			}

			res = append(res, func() *ImageElement {
				img := &ImageElement{
					ImageId: elem.CustomFace.FilePath.Unwrap(),
					Size:    elem.CustomFace.Size.Unwrap(),
					Width:   uint32(elem.CustomFace.Width.Unwrap()),
					Height:  uint32(elem.CustomFace.Height.Unwrap()),
					URL:     uri,
					Md5:     elem.CustomFace.Md5,
				}
				if elem.CustomFace.PbReserve != nil {
					img.SubType = elem.CustomFace.PbReserve.SubType.Unwrap()
					img.Summary = elem.CustomFace.PbReserve.Summary.Unwrap()
				}
				return img
			}())
		}
		// ***** HAS_OLD_CODE END *****

		// ***** HAS_OLD_CODE BIGIN *****
		if elem.NotOnlineImage != nil {
			if len(elem.NotOnlineImage.PicMd5) == 0 {
				continue
			}

			url := elem.NotOnlineImage.OrigUrl.Unwrap()
			if strings.Contains(url, "rkey") {
				url = "https://multimedia.nt.qq.com.cn" + url
			} else {
				url = "http://gchat.qpic.cn" + url
			}

			img := &ImageElement{
				ImageId: string(elem.NotOnlineImage.FilePath),
				Size:    elem.NotOnlineImage.FileLen.Unwrap(),
				Width:   elem.NotOnlineImage.PicWidth.Unwrap(),
				Height:  elem.NotOnlineImage.PicHeight.Unwrap(),
				URL:     url,
				Md5:     elem.NotOnlineImage.PicMd5,
			}
			// king ???
			/*
				if elem.NotOnlineImage.PbReserve != nil {
					img.SubType = elem.NotOnlineImage.PbReserve.SubType
					img.Summary = elem.NotOnlineImage.PbReserve.Summary
				}
			*/
			res = append(res, img)
		}

		// ImageEntity
		// RecordEntity
		// VideoEntity
		// new protocol image
		if elem.CommonElem != nil {
			switch elem.CommonElem.ServiceType.Unwrap() {
			case 48:
				extra, err := proto.Unmarshal[oidb.MsgInfo](elem.CommonElem.PbElem)
				if err != nil || len(extra.MsgInfoBody) == 0 {
					continue
				} // 不合理的合并转发会导致越界
				index := extra.MsgInfoBody[0].Index

				switch elem.CommonElem.BusinessType.Unwrap() {
				case 10, 20: // img
					res = append(res, &ImageElement{
						ImageId:  index.Info.FileName.Unwrap(),
						FileUuid: index.FileUuid.Unwrap(),
						SubType:  int32(extra.ExtBizInfo.Pic.BizType.Unwrap()),
						Summary:  utils.Ternary(extra.ExtBizInfo.Pic.TextSummary.Unwrap() == "", "[图片]", extra.ExtBizInfo.Pic.TextSummary.Unwrap()),
						Md5:      utils.MustParseHexStr(index.Info.FileHash.Unwrap()),
						Sha1:     utils.MustParseHexStr(index.Info.FileSha1.Unwrap()),
						Width:    index.Info.Width.Unwrap(),
						Height:   index.Info.Height.Unwrap(),
						Size:     index.Info.FileSize.Unwrap(),
						MsgInfo:  extra,
					})
				case 12, 22: // record 22 for Group
					res = append(res, &VoiceElement{
						Name:     index.Info.FileName.Unwrap(),
						Uuid:     index.FileUuid.Unwrap(),
						Md5:      utils.MustParseHexStr(index.Info.FileHash.Unwrap()),
						Sha1:     utils.MustParseHexStr(index.Info.FileSha1.Unwrap()),
						Duration: index.Info.Time.Unwrap(),
						Node:     index,
						MsgInfo:  extra,
					})
				case 11, 21: // video
					video := &ShortVideoElement{
						Name:     index.Info.FileName.Unwrap(),
						Uuid:     index.FileUuid.Unwrap(),
						Md5:      utils.MustParseHexStr(index.Info.FileHash.Unwrap()),
						Sha1:     utils.MustParseHexStr(index.Info.FileSha1.Unwrap()),
						Size:     index.Info.FileSize.Unwrap(),
						Node:     index,
						Duration: index.Info.Time.Unwrap(),
						MsgInfo:  extra,
					}
					if len(extra.MsgInfoBody) > 1 {
						info := extra.MsgInfoBody[1].Index
						video.Thumb = &VideoThumb{
							Size:   info.Info.FileSize.Unwrap(),
							Width:  info.Info.Width.Unwrap(),
							Height: info.Info.Height.Unwrap(),
							Md5:    utils.MustParseHexStr(info.Info.FileHash.Unwrap()),
							Sha1:   utils.MustParseHexStr(info.Info.FileSha1.Unwrap()),
						}
					}
					res = append(res, video)
				}
				// ***** HAS_OLD_CODE BIGIN *****
			case 3: // 闪照
				skipNext = true
				reader := binary.NewReader(elem.CommonElem.PbElem[1:])
				length, _ := reader.ReadUvarint()
				img, err := proto.Unmarshal[message.NotOnlineImage](reader.ReadBytes(int(length)))
				if err != nil {
					continue
				}
				res = append(res, &ImageElement{
					ImageId: string(img.FilePath),
					Md5:     img.PicMd5,
					Size:    img.FileLen.Unwrap(),
					// king ??? SubType: img.PbRes.SubType,
					Flash:   true,
					Summary: "[闪照]",
					Width:   img.PicWidth.Unwrap(),
					Height:  img.PicHeight.Unwrap(),
					URL:     fmt.Sprintf("http://gchat.qpic.cn/gchatpic_new/0/0-0-%X/0", img.PicMd5),
				})
			case 33:
				newSysFaceMsg, err := proto.Unmarshal[message.QSmallFaceExtra](elem.CommonElem.PbElem)
				if err != nil {
					continue
				}
				res = append(res, NewFace(newSysFaceMsg.FaceId.Unwrap()))
			case 37:
				skipNext = true
				faceExtra, err := proto.Unmarshal[message.QFaceExtra](elem.CommonElem.PbElem)
				if err != nil {
					continue
				}
				result, _ := strconv.ParseInt(faceExtra.ResultId.Unwrap(), 10, 32)
				res = append(res, &FaceElement{
					FaceId:      uint32(faceExtra.Qsid.Unwrap()),
					ResultId:    uint32(result),
					isLargeFace: true,
				}) // sticker 永远为单独消息
				// ***** HAS_OLD_CODE END *****
			}
		}

		// GroupFileEntity
		if elem.TransElemInfo != nil && elem.TransElemInfo.ElemType.Unwrap() == 24 {
			payload := binary.NewReader(elem.TransElemInfo.ElemValue)
			payload.SkipBytes(1)
			data := payload.ReadLengthBytes(prefix.Int16 | prefix.LengthOnly)
			extra, err := proto.Unmarshal[message.GroupFileExtra](data)
			if err != nil {
				continue
			}
			res = append(res, &FileElement{
				FileId:   extra.Inner.Info.FileId.Unwrap(),
				FileName: extra.Inner.Info.FileName.Unwrap(),
				FileSize: uint64(extra.Inner.Info.FileSize.Unwrap()),
				// king ??? FileMd5 : extra.Inner.Info.FileMd5.Unwrap(),
			})
		}

		// MultiMsgEntity
		if elem.RichMsg != nil && elem.RichMsg.ServiceId.Unwrap() == 35 {
			if elem.RichMsg.BytesTemplate1 != nil {
				xmlData := binary.ZlibUncompress(elem.RichMsg.BytesTemplate1[1:])
				var multimsg MultiMessage
				if err := xml.Unmarshal(xmlData, &multimsg); err == nil {
					res = append(res, NewForwardWithResID(multimsg.ResId))
				} else {
					res = append(res, &XMLElement{ServiceId: 35, Content: utils.B2S(xmlData)})
				}
			}
		}

		// LightAppEntity
		if elem.LightAppElem != nil && len(elem.LightAppElem.BytesData) > 1 {
			var content []byte
			switch elem.LightAppElem.BytesData[0] {
			case 0:
				content = elem.LightAppElem.BytesData[1:]
			case 1:
				content = binary.ZlibUncompress(elem.LightAppElem.BytesData[1:])
			}
			// 解析出错 or 非法内容
			if len(content) > 0 && len(content) < 1024*1024*1024 {
				res = append(res, NewLightApp(utils.B2S(content)))
			}
		}

		// ***** HAS_OLD_CODE BIGIN *****
		if elem.MarketFace != nil {
			res = append(res, &MarketFaceElement{
				Summary:    elem.MarketFace.FaceName.Unwrap(),
				ItemType:   elem.MarketFace.ItemType.Unwrap(),
				FaceInfo:   elem.MarketFace.FaceInfo.Unwrap(),
				FaceId:     elem.MarketFace.FaceId,
				TabId:      elem.MarketFace.TabId.Unwrap(),
				SubType:    elem.MarketFace.SubType.Unwrap(),
				EncryptKey: elem.MarketFace.Key,
				MediaType:  elem.MarketFace.MediaType.Unwrap(),
				MagicValue: utils.B2S(elem.MarketFace.MobileParam),
			})
		}
		// ***** HAS_OLD_CODE END *****
	}
	return
}

func ToReadableString(m []IMessageElement) string {
	sb := new(strings.Builder)
	for _, elem := range m {
		sb.WriteString(ToReadableStringEle(elem))
	}
	return sb.String()
}

func ToReadableStringEle(elem IMessageElement) string {
	switch elem.Type() {
	case Text:
		return elem.(*TextElement).Content
	case At:
		return elem.(*AtElement).Display
	case Image:
		return "[图片]"
	case Reply:
		return "[回复]" // [Optional] + ToReadableString(e.Elements), 这里不破坏原义不添加
	case Face:
		return "[表情]"
	case Voice:
		return "[语音]"
	case Video:
		return "[视频]"
	case Service:
		return "[XML]"
	case LightApp:
		return "[卡片消息]"
	case Forward:
		return "[转发消息]"
	case MarketFace:
		return "[魔法表情]"
	case File:
		return "[文件]"
	case RedBag:
		return "[红包]"
	default:
		return "[暂不支持该消息类型]"
	}
}

/*
func NewSendingMessage() *SendingMessage { return &SendingMessage{} }

func (msg *SendingMessage) GetElems() []IMessageElement { return msg.Elements }

// Append 要传入msg的引用
func (msg *SendingMessage) Append(e IMessageElement) *SendingMessage {
	v := reflect.ValueOf(e)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		msg.Elements = append(msg.Elements, e)
	}
	return msg
}

func (msg *SendingMessage) FirstOrNil(f func(element IMessageElement) bool) IMessageElement {
	for _, elem := range msg.Elements {
		if f(elem) { return elem }
	}
	return nil
}
*/
