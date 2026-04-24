package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"

	pkt_msg "github.com/kernel-ai/koscore/client/packets/message"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/message/message_type"
	"github.com/kernel-ai/koscore/client/packets/oidb"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/highway"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/system"
	"github.com/kernel-ai/koscore/client/packets/websso"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/exception"
	"github.com/kernel-ai/koscore/utils/proto"
)

// 获取Rkey
func (m *QQClient) FetchRkey() (entity.RKeyMap, error) {
	pkt, e := oidb.BuildFetchRKeyPacket()
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseFetchRKeyPacket(pkt.Data)
}

// 设置在线状态
func (m *QQClient) SetOnlineStatus(status operation.SetStatus) error {
	data, _ := proto.Marshal(&status)
	pkt, err := m.sendOidbPacketAndWait(message_type.AttributeSetStatus.NewSsoPacket(m.Session().GetAndIncreaseSequence(), data))
	if err != nil {
		return err
	}
	rsp, err := proto.Unmarshal[operation.SetStatusResponse](pkt.Data)
	if err != nil {
		return err
	}
	if rsp.Message != "set status success" {
		return exception.NewFormat("set status failed: %s", rsp.Message)
	}
	return nil
}

// 获取单向好友列表
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/web.go#L23
func (m *QQClient) GetUnidirectionalFriendList() ([]*entity.User, error) {
	rsp, err := m.webSsoRequest("ti.qq.com", "OidbSvc.0xe17_0", fmt.Sprintf(`{"uint64_uin":%v,"uint64_top":0,"uint32_req_num":99,"bytes_cookies":""}`, m.session.Info.Uin))
	if err != nil {
		return nil, err
	}
	return websso.ParseUnidirectionalFriendsPacket(utils.S2B(rsp))
}

// 删除单向好友
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/web.go#L62
func (m *QQClient) DeleteUnidirectionalFriend(uin uint64) error {
	rsp, err := m.webSsoRequest("ti.qq.com", "OidbSvc.0x5d4_0", fmt.Sprintf(`{"uin_list":[%v]}`, uin))
	if err != nil {
		return err
	}
	webRsp := &struct {
		ErrorCode int32 `json:"ErrorCode"`
	}{}
	if err = json.Unmarshal(utils.S2B(rsp), webRsp); err != nil {
		return exception.NewUnmarshalJsonException(err, "web sso")
	}
	if webRsp.ErrorCode != 0 {
		return exception.NewFormat("web sso request error: %v", webRsp.ErrorCode)
	}
	return nil
}

// 获取对应群的群成员信息
func (m *QQClient) FetchGroupMember(group_uin, member_uin uint64) (*entity.GroupMember, error) {
	uid, err := m.GetUid(member_uin, group_uin)
	if err != nil {
		return nil, err
	}
	pkt, err := oidb.BuildFetchGroupMemberPacket(group_uin, uid)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	return oidb.ParseFetchGroupMemberPacket(pkt.Data)
}

// 发送群聊打卡消息
func (m *QQClient) SendGroupSign(group_uin uint64) (*oidb.BotGroupClockInResult, error) {
	pkt, e := oidb.BuildGroupSignPacket(m.session.Info.Uin, group_uin, m.version.CurrentVersion)
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseGroupSignResp(pkt.Data)
}

// 获取剩余@全员次数
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/group_msg.go#L68
func (m *QQClient) GetAtAllRemain(uin, group_uin uint64) (*oidb.AtAllRemainInfo, error) {
	pkt, err := oidb.BuildGetAtAllRemainRequest(uin, group_uin)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	return oidb.ParseGetAtAllRemainResponse(pkt.Data)
}

// 通过TX服务器检查URL安全性
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/security.go#L24
func (m *QQClient) CheckURLSafely(url string) (oidb.URLSecurityLevel, error) {
	pkt, err := oidb.BuildURLCheckRequest(m.session.Info.Uin, url)
	if err != nil {
		return oidb.URLSecurityLevelUnknown, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return oidb.URLSecurityLevelUnknown, err
	}
	return oidb.ParseURLCheckResponse(pkt.Data)
}

// 图片识别 有些域名的图可能无法识别，需要重新上传到tx服务器并获取图片下载链接
func (m *QQClient) ImageOcr(uri string) (*oidb.OcrResponse, error) {
	if uri == "" {
		return nil, errors.New("image url error")
	}
	pkt, e := oidb.BuildImageOcrRequestPacket(uri)
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseImageOcrResp(pkt.Data)
}

// 获取AI语音角色列表
func (m *QQClient) GetAiCharacters(gin uint64, chat_type entity.ChatType) (*entity.AiCharacterList, error) {
	if gin == 0 {
		gin = 42
	}
	pkt, e := oidb.BuildAiCharacterListService(gin, chat_type)
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	result, e := oidb.ParseAiCharacterListService(pkt.Data)
	if e != nil {
		return nil, e
	}
	result.Type = chat_type
	return result, nil
}

// 发送群AI语音
func (m *QQClient) SendGroupAiRecord(group_uin uint64, chat_type entity.ChatType, voice_id, text string) (*message.VoiceElement, error) {
	pkt, e := oidb.BuildGroupAiRecordService(group_uin, voice_id, text, chat_type, crypto.RandU32())
	if e != nil {
		return nil, e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return nil, e
	}
	return oidb.ParseGroupAiRecordService(pkt.Data)
}

// 处理好友请求
func (m *QQClient) SetFriendRequest(accept bool, target_uid string) error {
	pkt, err := oidb.BuildSetFriendRequestPacket(accept, target_uid)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 给好友点赞
func (m *QQClient) SendFriendLike(uin uint64, count uint32) error {
	uid, err := m.GetUid(uin)
	if err != nil {
		return err
	}
	if count > 20 {
		count = 20
	} else if count < 1 {
		count = 1
	}
	pkt, err := oidb.BuildFriendLikePacket(uid, count)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 删除好友
func (m *QQClient) DeleteFriend(uin uint64, block bool) error {
	uid, err := m.GetUid(uin)
	if err != nil {
		return err
	}
	pkt, err := oidb.BuildDeleteFriendPacket(uid, block)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 获取群公告
func (m *QQClient) GetGroupNotice(group_uin uint64) (l []*entity.GroupNoticeFeed, err error) {
	bkn, err := m.GetCsrfToken()
	if err != nil {
		return nil, err
	}
	var v url.Values
	v.Set("bkn", strconv.Itoa(bkn))
	v.Set("qid", strconv.FormatInt(int64(group_uin), 10))
	v.Set("ft", "23")
	v.Set("ni", "1")
	v.Set("n", "1")
	v.Set("i", "1")
	v.Set("log_read", "1")
	v.Set("platform", "1")
	v.Set("s", "-1")
	v.Set("n", "20")
	req, _ := http.NewRequest(http.MethodGet, "https://web.qun.qq.com/cgi-bin/announce/get_t_list?"+v.Encode(), nil)
	rsp, err := m.SendRequestWithCookie(req)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, exception.NewFormat("error resp code %d", rsp.StatusCode)
	}
	var r entity.GroupNoticeRsp
	if err = json.NewDecoder(rsp.Body).Decode(&r); err != nil {
		return
	}
	_ = rsp.Body.Close()
	o := make([]*entity.GroupNoticeFeed, 0, len(r.Feeds)+len(r.Inst))
	o = append(o, r.Feeds...)
	o = append(o, r.Inst...)
	return o, nil
}

// 发群公告
func (m *QQClient) AddGroupNoticeSimple(group_uin uint64, text string) (noticeId string, err error) {
	bkn, err := m.GetCsrfToken()
	if err != nil {
		return "", err
	}
	body := fmt.Sprintf(`qid=%v&bkn=%v&text=%v&pinned=0&type=1&settings={"is_show_edit_card":0,"tip_window_type":1,"confirm_required":1}`, group_uin, bkn, url.QueryEscape(text))
	req, err := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/add_qun_notice?bkn="+strconv.Itoa(bkn), strings.NewReader(body))
	if err != nil {
		return "", err
	}
	rsp, err := m.SendRequestWithCookie(req)
	if err != nil {
		return "", err
	}
	var res entity.NoticeSendResp
	if err = json.NewDecoder(rsp.Body).Decode(&res); err != nil {
		return "", err
	}
	_ = rsp.Body.Close()
	return res.NoticeId, nil
}

func (m *QQClient) uploadGroupNoticePic(bkn int, img []byte) (*entity.NoticeImage, error) {
	ret := &entity.NoticeImage{}
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	_ = w.WriteField("bkn", strconv.Itoa(bkn))
	_ = w.WriteField("source", "troopNotice")
	_ = w.WriteField("m", "0")
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="pic_up"; filename="temp_uploadFile.png"`)
	h.Set("Content-Type", "image/png")
	fw, _ := w.CreatePart(h)
	_, _ = fw.Write(img)
	_ = w.Close()
	req, err := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/upload_img", buf)
	if err != nil {
		return ret, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	rsp, err := m.SendRequestWithCookie(req)
	if err != nil {
		return ret, err
	}
	var res entity.NoticePicUpResponse
	if err = json.NewDecoder(rsp.Body).Decode(&res); err != nil {
		return ret, err
	}
	_ = rsp.Body.Close()
	if res.ErrorCode != 0 {
		return ret, errors.New(res.ErrorMessage)
	}
	if err = json.Unmarshal([]byte(html.UnescapeString(res.Id)), &ret); err != nil {
		return ret, err
	}
	return ret, nil
}

// 发群公告带图片
func (m *QQClient) AddGroupNoticeWithPic(group_uin uint64, text string, pic []byte) (noticeId string, err error) {
	bkn, e := m.GetCsrfToken()
	if e != nil {
		return "", e
	}
	img, e := m.uploadGroupNoticePic(bkn, pic)
	if e != nil {
		return "", e
	}
	body := fmt.Sprintf(`qid=%v&bkn=%v&text=%v&pinned=0&type=1&settings={"is_show_edit_card":0,"tip_window_type":1,"confirm_required":1}&pic=%v&imgWidth=%v&imgHeight=%v`, group_uin, bkn, url.QueryEscape(text), img.Id, img.Width, img.Height)
	req, e := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/add_qun_notice?bkn="+strconv.Itoa(bkn), strings.NewReader(body))
	if e != nil {
		return "", e
	}
	rsp, e := m.SendRequestWithCookie(req)
	if e != nil {
		return "", e
	}
	var res entity.NoticeSendResp
	if e = json.NewDecoder(rsp.Body).Decode(&res); e != nil {
		return "", e
	}
	_ = rsp.Body.Close()
	return res.NoticeId, nil
}

// 删除群公告
func (m *QQClient) DelGroupNotice(group_uin uint64, fid string) error {
	bkn, e := m.GetCsrfToken()
	if e != nil {
		return e
	}
	body := fmt.Sprintf(`fid=%s&qid=%v&bkn=%v&ft=23&op=1`, fid, group_uin, bkn)
	req, e := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/del_feed", strings.NewReader(body))
	if e != nil {
		return e
	}
	rsp, e := m.SendRequestWithCookie(req)
	if e != nil {
		return e
	}
	_ = rsp.Body.Close()
	return nil
}

// 禁言群成员
func (m *QQClient) SetGroupMemberMute(group_uin, uin uint64, duration uint32) error {
	uid, err := m.GetUid(uin, group_uin)
	if err != nil {
		return err
	}
	pkt, err := oidb.BuildSetGroupMemberMutePacket(group_uin, uid, duration)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.ParseSetGroupMemberMutePacket(pkt.Data)
}

// 设置群管理员
func (m *QQClient) SetGroupAdmin(group_uin, uin uint64, is_admin bool) error {
	uid, err := m.GetUid(uin, group_uin)
	if err != nil {
		return err
	}
	pkt, err := oidb.BuildSetGroupAdminPacket(group_uin, uid, is_admin)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	if err = oidb.CheckError(pkt.Data); err != nil {
		return err
	}
	if g := m.GetCachedMemberInfo(uin, group_uin); g != nil {
		g.Permission = entity.Admin
		m.cache.RefreshGroupMember(group_uin, g)
	}
	return nil
}

// 设置群头像
func (m *QQClient) SetGroupAvatar(group_uin uint64, avatar io.ReadSeeker) error {
	if avatar == nil {
		return errors.New("avatar is nil")
	}
	ext, e := proto.Marshal(&highway.GroupAvatarExtra{
		Type:     101,
		GroupUin: group_uin,
		Field3:   &highway.GroupAvatarExtraField3{Field1: 1},
		Field5:   3,
		Field6:   1,
	})
	if e != nil {
		return e
	}
	md5, size := crypto.ComputeMd5AndLength(avatar)
	return m.highwayUpload(3000, avatar, uint64(size), md5, ext)
}

// 设置群聊备注
func (m *QQClient) SetGroupRemark(group_uin uint64, remark string) error {
	pkt, err := oidb.BuildSetGroupRemarkPacket(group_uin, remark)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 踢出群成员，可选是否拒绝加群请求
func (m *QQClient) KickGroupMember(group_uin, uin uint64, reject_add_request bool) error {
	uid, err := m.GetUid(uin, group_uin)
	if err != nil {
		return err
	}
	pkt, err := oidb.BuildKickGroupMemberPacket(group_uin, uid, reject_add_request)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.ParseKickGroupMemberPacket(pkt.Data)
}

// 获取精华消息
func (m *QQClient) FetchEssenceMessage(group_uin uint64) ([]*message.GroupEssenceMessage, error) {
	bkn, err := m.GetCsrfToken()
	if err != nil {
		return nil, err
	}
	grp_info := m.GetCachedGroupInfo(group_uin)
	var essences []*message.GroupEssenceMessage
	page := 0
	for {
		uri := fmt.Sprintf("https://qun.qq.com/cgi-bin/group_digest/digest_list?random=7800&X-CROSS-ORIGIN=fetch&group_code=%d&page_start=%d&page_limit=20&bkn=%d", group_uin, page, bkn)
		req, err := http.NewRequest(http.MethodGet, uri, nil)
		if err != nil {
			return essences, err
		}
		rsp, err := m.SendRequestWithCookie(req)
		if err != nil {
			return essences, err
		}
		data, err := io.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		_ = rsp.Body.Close()
		if rsp.StatusCode != http.StatusOK {
			return essences, exception.NewFormat("error resp code %d", rsp.StatusCode)
		}
		rsp_json := gjson.ParseBytes(data)
		if rsp_json.Get("retcode").Int() != 0 {
			return essences, exception.NewFormat("error code %d, %s", rsp_json.Get("retcode").Int(), rsp_json.Get("retmsg").String())
		}
		for _, v := range rsp_json.Get("data").Get("msg_list").Array() {
			var elements []message.IMessageElement
			for _, e := range v.Get("msg_content").Array() {
				switch e.Get("msg_type").Int() {
				case 1:
					elements = append(elements, &message.TextElement{Content: e.Get("text").String()})
				case 2:
					elements = append(elements, &message.FaceElement{FaceId: uint32(e.Get("face_index").Int())})
				case 3:
					elements = append(elements, &message.ImageElement{URL: e.Get("image_url").String()})
				case 4:
					elements = append(elements, &message.FileElement{
						FileId:  e.Get("file_id").String(),
						FileURL: e.Get("file_thumbnail_url").String(),
					})
				}
			}
			sender_uin := uint64(v.Get("sender_uin").Int())
			sender_info := m.GetCachedMemberInfo(sender_uin, group_uin)
			essences = append(essences, &message.GroupEssenceMessage{
				OperatorUin:  uint64(v.Get("add_digest_uin").Int()),
				OperatorUid:  m.get_uid(uint64(v.Get("add_digest_uin").Int())),
				OperatorTime: uint64(v.Get("add_digest_time").Int()),
				CanRemove:    v.Get("can_be_removed").Bool(),
				Message: &message.GroupMessage{
					Message: &message.Message{
						Id:     uint64(v.Get("msg_seq").Int()),
						Random: uint64(v.Get("msg_random").Int()),
						Time:   v.Get("sender_time").Int(),
						Sender: message.Sender{
							Uin:      sender_uin,
							Uid:      m.get_uid(sender_uin, group_uin),
							Nickname: sender_info.Nickname,
							CardName: sender_info.MemberCard,
						},
						Elements: elements,
					},
					GroupUin:  grp_info.GroupUin,
					GroupName: grp_info.GroupName,
				},
			})
		}
		if rsp_json.Get("data").Get("is_end").Bool() {
			break
		}
	}
	return essences, nil
}

// 设置群聊精华消息
func (m *QQClient) SetEssenceMessage(group_uin, seq, random uint64, is_set bool) error {
	pkt, err := oidb.BuildSetEssenceMessagePacket(group_uin, seq, random, is_set)
	if err != nil {
		return err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 获取群荣誉信息
// reference https://github.com/Mrs4s/MiraiGo/blob/master/client/http_api.go
func (m *QQClient) GetGroupHonorInfo(group_uin uint64, honor_type entity.HonorType) (*entity.GroupHonorInfo, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://qun.qq.com/interactive/honorlist?gc=%d&type=%d", group_uin, honor_type), nil)
	if err != nil {
		return nil, err
	}
	rsp, err := m.SendRequestWithCookie(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	_ = rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, exception.NewFormat("error resp code %d", rsp.StatusCode)
	}
	matched := regexp.MustCompile(`window\.__INITIAL_STATE__\s*?=\s*?(\{.*})`).FindSubmatch(data)
	if len(matched) == 0 {
		return nil, errors.New("no matched data")
	}
	var ret entity.GroupHonorInfo
	if err = json.NewDecoder(bytes.NewReader(matched[1])).Decode(&ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// 标记私聊消息已读
func (m *QQClient) MarkPrivateMessageReaded(uin uint64, timestamp int64, start_seq uint64) error {
	uid, err := m.GetUid(uin)
	if err != nil {
		return err
	}
	pkt, err := m.sendOidbPacketAndWait(system.BuildPrivateSsoReadedReportPacket(uid, timestamp, start_seq))
	if err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 标记群消息已读
func (m *QQClient) MarkGroupMessageReaded(gin, start_seq uint64) error {
	pkt, err := m.sendOidbPacketAndWait(system.BuildGroupSsoReadedReportPacket(gin, start_seq))
	if err != nil {
		return err
	}
	return oidb.CheckError(pkt.Data)
}

// 获取魔法表情key
func (m *QQClient) FetchMarketFaceKey(face_ids ...string) ([]string, error) {
	pkt, e := m.sendOidbPacketAndWait(pkt_msg.BuildMarketFaceKeyPacket(face_ids...))
	if e != nil {
		return nil, e
	}
	return pkt_msg.ParseMarketFaceKeyPacket(pkt.Data)
}

// 设置头像
func (m *QQClient) SetAvatar(avatar io.ReadSeeker) error {
	if avatar == nil {
		return errors.New("avatar is nil")
	}
	md5, size := crypto.ComputeMd5AndLength(avatar)
	return m.highwayUpload(90, avatar, uint64(size), md5, nil)
}
