package client

import (
	"errors"
	"fmt"

	pkt_msg "github.com/kernel-ai/koscore/client/packets/message"
	pb_msg "github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
)

func (m *QQClient) SendRawMessage(route *pb_msg.SendRoutingHead, body *pb_msg.MessageBody) (rsp *pb_msg.PbSendMsgResp, seq, random uint32, err error) {
	var pkt *sso_type.SsoPacket
	seq, random = m.session.GetSequence(), crypto.RandU32()
	if pkt, err = pkt_msg.BuildRawMessage(seq, random, route, body); err != nil {
		return
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return
	}
	if rsp, err = pkt_msg.ParseMessagePacket(pkt.Data); err != nil {
		return
	}
	if rsp.Result.Unwrap() != 0 {
		return nil, 0, 0, fmt.Errorf("operation exception: %d", rsp.Result.Unwrap())
	}
	return
}

func (m *QQClient) SendGroupMessage(gin uint64, elements []message.IMessageElement, needPreprocess ...bool) (*message.GroupMessage, error) {
	if needPreprocess == nil || needPreprocess[0] {
		elements = m.preProcessGroupMessage(gin, elements)
	}
	rsp, _, random, err := m.SendRawMessage(&pb_msg.SendRoutingHead{
		Group: &pb_msg.Grp{GroupUin: proto.Some(int64(gin))},
	}, message.PackElementsToBody(elements))
	if err != nil {
		return nil, err
	}
	if rsp.Sequence.Unwrap() == 0 && rsp.ClientSequence.Unwrap() == 0 {
		return nil, errors.New("ret group sequence 0")
	}
	return &message.GroupMessage{
		GroupUin:  gin,
		GroupName: m.GetCachedGroupInfo(gin).GroupName,
		Message: &message.Message{
			Id:        rsp.Sequence.Unwrap(),
			Random:    uint64(random),
			ClientSeq: rsp.ClientSequence.Unwrap(),
			Time:      rsp.SendTime.Unwrap(),
			Elements:  elements,
			Sender: message.Sender{
				Uin:      m.UIN(),
				Uid:      m.Uid(),
				Nickname: m.Nick(),
				CardName: m.GetCachedMemberInfo(m.UIN(), gin).MemberCard,
				IsFriend: false,
			},
		},
	}, nil
}

func (m *QQClient) SendPrivateMessage(uin uint64, elements []message.IMessageElement, needPreprocess ...bool) (*message.PrivateMessage, error) {
	if needPreprocess == nil || needPreprocess[0] {
		elements = m.preProcessPrivateMessage(uin, elements)
	}
	rsp, _, random, err := m.SendRawMessage(&pb_msg.SendRoutingHead{
		C2C: &pb_msg.C2C{PeerUin: proto.Some(int64(uin)), PeerUid: proto.Some(m.GetUid(uin))},
	}, message.PackElementsToBody(elements))
	if err != nil {
		return nil, err
	}
	if rsp.Sequence.Unwrap() == 0 && rsp.ClientSequence.Unwrap() == 0 {
		return nil, errors.New("ret private sequence 0")
	}

	return &message.PrivateMessage{
		Message: &message.Message{
			Id:        rsp.Sequence.Unwrap(),
			Random:    uint64(random),
			ClientSeq: rsp.ClientSequence.Unwrap(),
			Time:      rsp.SendTime.Unwrap(),
			Elements:  elements,
			Sender: message.Sender{
				Uin:      m.UIN(),
				Uid:      m.Uid(),
				Nickname: m.Nick(),
				IsFriend: true,
			},
		},
	}, nil
}

// 发送私聊文件
func (m *QQClient) SendPrivateFile(uin uint64, localFilePath, filename string) error {
	fsem, e := message.NewLocalFile(localFilePath, filename)
	if e != nil {
		return e
	}
	fsem, e = m.UploadPrivateFile(uin, fsem)
	if e != nil {
		return e
	}
	rsp, _, _, e := m.SendRawMessage(&pb_msg.SendRoutingHead{
		Trans0X211: &pb_msg.Trans0X211{
			ToUin: proto.Some(int64(uin)),
			CcCmd: proto.Some[uint32](4),
			Uid:   proto.Some(m.GetUid(uin)),
		},
	}, message.PackElementsToBody([]message.IMessageElement{fsem}))
	if e != nil {
		return e
	}
	if rsp.Sequence.Unwrap() == 0 {
		return errors.New("ret private file sequence 0")
	}
	if rsp.ClientSequence.Unwrap() == 0 {
		return errors.New("ret private file client sequence 0")
	}
	return nil
}

// 发送群文件
func (m *QQClient) SendGroupFile(gin uint64, local_path, filename, target_dir string) error {
	fsem, e := message.NewLocalFile(local_path, filename)
	if e != nil {
		return e
	}
	if _, e = m.UploadGroupFile(gin, fsem, target_dir); e != nil {
		return e
	}
	return nil
}

// make a fake message
func (m *QQClient) BuildFakeMessage(nodes []*message.ForwardNode) []*pb_msg.CommonMessage {
	body := make([]*pb_msg.CommonMessage, len(nodes))
	seq := crypto.RandU32()
	for idx, node := range nodes {
		body[idx] = &pb_msg.CommonMessage{
			RoutingHead: &pb_msg.RoutingHead{
				FromUid: proto.Some(m.GetUid(node.SenderId)),
				FromUin: proto.Some(int64(node.SenderId)),
			},
			ContentHead: &pb_msg.ContentHead{
				Type:           proto.Some(int32(node.Type)),
				Random:         proto.Some(seq),
				Sequence:       proto.Some(uint64(seq) + uint64(idx)),
				Time:           proto.Some(int64(node.Time)),
				ClientSequence: proto.Some[uint64](1),
				MsgUid:         proto.Some[uint64](0),
			}}
		if node.GroupId != 0 {
			body[idx].RoutingHead.Group = &pb_msg.CommonGroup{
				GroupCode:     proto.Some(int64(node.GroupId)),
				GroupCard:     proto.Some(node.SenderName),
				GroupCardType: proto.Some[int32](2),
			}
			m.preProcessGroupMessage(node.GroupId, node.Message)
		} else {
			body[idx].RoutingHead.CommonC2C = &pb_msg.CommonC2C{Name: proto.Some(node.SenderName)}
			body[idx].RoutingHead.ToUid = proto.Some(m.Uid())
			body[idx].RoutingHead.ToUin = proto.Some(int64(m.UIN()))
			body[idx].ContentHead.SubType = proto.Some[int32](4)
			body[idx].ContentHead.DivSeq = proto.Some[int32](4)
			m.preProcessPrivateMessage(m.UIN(), node.Message)
		}
		body[idx].MessageBody = message.PackElementsToBody(node.Message)
	}
	return body
}

// *****

func (m *QQClient) preProcessGroupMessage(groupUin uint64, elements []message.IMessageElement) []message.IMessageElement {
	for _, element := range elements {
		switch elem := element.(type) {
		case *message.AtElement:
			if mem := m.GetCachedMemberInfo(elem.TargetUin, groupUin); mem != nil {
				elem.TargetUid = mem.Uid
				elem.Display = "@" + mem.DisplayName()
			}
		case *message.ImageElement:
			if elem.MsgInfo != nil {
				continue
			}
			if _, err := m.UploadGroupImage(groupUin, elem); err != nil {
				m.LOGE("%v", err)
			} else if elem.MsgInfo == nil {
				m.LOGE("UploadGroupImage failed")
			}
		case *message.VoiceElement:
			if elem.MsgInfo != nil {
				continue
			}
			if _, err := m.UploadGroupRecord(groupUin, elem); err != nil {
				m.LOGE("%v", err)
			} else if elem.MsgInfo == nil {
				m.LOGE("UploadGroupRecord failed")
			}
		case *message.ShortVideoElement:
			if elem.MsgInfo != nil {
				continue
			}
			if _, err := m.UploadGroupShortVideo(groupUin, elem); err != nil {
				m.LOGE("%v", err)
			} else if elem.MsgInfo == nil {
				m.LOGE("UploadGroupVideo failed")
			}
		case *message.ForwardMessage:
			if elem.ResId != "" && len(elem.Nodes) == 0 {
				forward, _ := m.FetchForwardMsg(elem.ResId, true)
				elem.IsGroup = true
				elem.Nodes = forward.Nodes
			}
			if elem.ResId == "" && len(elem.Nodes) != 0 {
				if _, err := m.UploadForwardMsg(elem, groupUin); err != nil {
					m.LOGE("%v", err)
				}
			}
		}
	}
	return elements
}

func (m *QQClient) preProcessPrivateMessage(targetUin uint64, elements []message.IMessageElement) []message.IMessageElement {
	for _, element := range elements {
		switch elem := element.(type) {
		case *message.ImageElement:
			if elem.MsgInfo != nil {
				continue
			}
			if _, err := m.UploadPrivateImage(targetUin, elem); err != nil {
				m.LOGE("%v", err)
			} else if elem.MsgInfo == nil {
				m.LOGE("UploadPrivateImage failed")
			}
		case *message.VoiceElement:
			if elem.MsgInfo != nil {
				continue
			}
			if _, err := m.UploadPrivateRecord(targetUin, elem); err != nil {
				m.LOGE("%v", err)
			} else if elem.MsgInfo == nil {
				m.LOGE("UploadPrivateRecord failed")
			}
		case *message.ShortVideoElement:
			if elem.MsgInfo != nil {
				continue
			}
			if _, err := m.UploadPrivateShortVideo(targetUin, elem); err != nil {
				m.LOGE("%v", err)
			} else if elem.MsgInfo == nil {
				m.LOGE("UploadPrivateVideo failed")
			}
		case *message.ForwardMessage:
			if elem.ResId != "" && len(elem.Nodes) == 0 {
				forward, _ := m.FetchForwardMsg(elem.ResId, false)
				elem.SelfId = m.UIN()
				elem.Nodes = forward.Nodes
			}
			if elem.ResId == "" && len(elem.Nodes) != 0 {
				if _, err := m.UploadForwardMsg(elem, 0); err != nil {
					m.LOGE("%v", err)
				}
			}
		}
	}
	return elements
}
