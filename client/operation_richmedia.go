package client

import (
	"errors"

	pb_msg "github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	pb_oidb "github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"

	"github.com/kernel-ai/koscore/client/packets/oidb"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
)

// 上传群聊图片
func (m *QQClient) UploadGroupImage(gin uint64, image *message.ImageElement) (*message.ImageElement, error) {
	if image == nil || image.Stream == nil {
		return nil, errors.New("image is nil")
	}
	defer utils.CloseIO(image.Stream)
	image.IsGroup = true
	pkt, err := oidb.BuildGroupImageUploadPacket(gin, image)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := oidb.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}
	if md5, ext := oidb.NTV2RICH_MEDIA_IMAGE.CommonGenerateHighwayExt(upload, image.Stream, image.Size, nil); ext != nil {
		//m.LOGD("group image upload ukey: %s", upload.UKey.Unwrap())
		if err = m.highwayUpload(1004, image.Stream, uint64(image.Size), md5, ext); err != nil {
			return nil, err
		}
	}
	image.CompatFace, _ = proto.Unmarshal[pb_msg.CustomFace](upload.CompatQMsg)
	image.MsgInfo = upload.MsgInfo
	image.FileUuid = upload.MsgInfo.MsgInfoBody[0].Index.FileUuid.Unwrap()
	return image, nil
}

// 上传私聊图片
func (m *QQClient) UploadPrivateImage(uin uint64, image *message.ImageElement) (*message.ImageElement, error) {
	if image == nil || image.Stream == nil {
		return nil, errors.New("image is nil")
	}
	defer utils.CloseIO(image.Stream)
	image.IsGroup = false
	uid, err := m.GetUid(uin)
	if err != nil {
		return nil, err
	}
	pkt, err := oidb.BuildPrivateImageUploadPacket(uid, image)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := oidb.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}
	if md5, ext := oidb.NTV2RICH_MEDIA_IMAGE.CommonGenerateHighwayExt(upload, image.Stream, image.Size, nil); ext != nil {
		//m.LOGD("private image upload ukey: %s", ukey)
		if err = m.highwayUpload(1003, image.Stream, uint64(image.Size), md5, ext); err != nil {
			return nil, err
		}
	}
	image.CompatImage, _ = proto.Unmarshal[pb_msg.NotOnlineImage](upload.CompatQMsg)
	image.MsgInfo = upload.MsgInfo
	image.FileUuid = upload.MsgInfo.MsgInfoBody[0].Index.FileUuid.Unwrap()
	return image, nil
}

// 上传群聊视频
func (m *QQClient) UploadGroupShortVideo(gin uint64, video *message.ShortVideoElement) (*message.ShortVideoElement, error) {
	if video == nil || video.Stream == nil {
		return nil, errors.New("video is nil")
	}
	if video.Thumb == nil || video.Thumb.Stream == nil {
		return nil, errors.New("video thumb is nil")
	}
	defer utils.CloseIO(video.Stream)
	defer utils.CloseIO(video.Thumb.Stream)
	video.IsGroup = true
	pkt, err := oidb.BuildGroupVideoUploadPacket(gin, video)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := oidb.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	// video
	if md5, ext := oidb.NTV2RICH_MEDIA_VIdEO.CommonGenerateHighwayExt(upload, video.Stream, video.Size, nil); ext != nil {
		//m.LOGD("group video upload ukey: %s", ukey)
		if err = m.highwayUpload(1005, video.Stream, uint64(video.Size), md5, ext); err != nil {
			return nil, err
		}
	}
	// thumb
	if md5, ext := oidb.NTV2RICH_MEDIA_IMAGE.CommonGenerateHighwayExt(upload, video.Thumb.Stream, video.Thumb.Size, upload.SubFileInfos[0]); ext != nil {
		//m.LOGD("group video.thumb upload ukey: %s", ukey)
		if err = m.highwayUpload(1006, video.Thumb.Stream, uint64(video.Thumb.Size), md5, ext); err != nil {
			return nil, err
		}
	}

	video.Compat, _ = proto.Unmarshal[pb_msg.VideoFile](upload.CompatQMsg)
	video.MsgInfo = upload.MsgInfo
	video.Name = video.Compat.FileName.Unwrap()
	video.Uuid = video.Compat.FileUuid.Unwrap()
	return video, nil
}

// 上传私聊视频
func (m *QQClient) UploadPrivateShortVideo(uin uint64, video *message.ShortVideoElement) (*message.ShortVideoElement, error) {
	if video == nil || video.Stream == nil {
		return nil, errors.New("video is nil")
	}
	if video.Thumb == nil || video.Thumb.Stream == nil {
		return nil, errors.New("video thumb is nil")
	}
	defer utils.CloseIO(video.Stream)
	defer utils.CloseIO(video.Thumb.Stream)
	video.IsGroup = false
	uid, err := m.GetUid(uin)
	if err != nil {
		return nil, err
	}
	pkt, err := oidb.BuildPrivateVideoUploadPacket(uid, video)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := oidb.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	// video
	if md5, ext := oidb.NTV2RICH_MEDIA_VIdEO.CommonGenerateHighwayExt(upload, video.Stream, video.Size, nil); ext != nil {
		//m.LOGD("pivate video upload ukey: %s", ukey)
		if err = m.highwayUpload(1001, video.Stream, uint64(video.Size), md5, ext); err != nil {
			return nil, err
		}
	}
	// thumb
	if md5, ext := oidb.NTV2RICH_MEDIA_IMAGE.CommonGenerateHighwayExt(upload, video.Thumb.Stream, video.Thumb.Size, upload.SubFileInfos[0]); ext != nil {
		//m.LOGD("pivate video.thumb upload ukey: %s", ukey)
		if err = m.highwayUpload(1002, video.Thumb.Stream, uint64(video.Thumb.Size), md5, ext); err != nil {
			return nil, err
		}
	}

	if video.Compat, err = proto.Unmarshal[pb_msg.VideoFile](upload.CompatQMsg); err != nil {
		return nil, err
	}
	video.MsgInfo = upload.MsgInfo
	video.Name = video.Compat.FileName.Unwrap()
	video.Uuid = video.Compat.FileUuid.Unwrap()
	return video, nil
}

// 上传群聊语音片
func (m *QQClient) UploadGroupRecord(gin uint64, voice *message.VoiceElement) (*message.VoiceElement, error) {
	if voice == nil || voice.Stream == nil {
		return nil, errors.New("voice is nil")
	}
	defer utils.CloseIO(voice.Stream)
	voice.IsGroup = true
	pkt, err := oidb.BuildGroupRecordUploadPacket(gin, voice)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := oidb.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	if md5, ext := oidb.NTV2RICH_MEDIA_RECORD.CommonGenerateHighwayExt(upload, voice.Stream, voice.Size, nil); ext != nil {
		//m.LOGD("group record upload ukey: %s", ukey)
		if err = m.highwayUpload(1008, voice.Stream, uint64(voice.Size), md5, ext); err != nil {
			return nil, err
		}
	}

	voice.MsgInfo = upload.MsgInfo
	voice.Uuid = upload.MsgInfo.MsgInfoBody[0].Index.FileUuid.Unwrap()
	voice.Compat = upload.CompatQMsg
	return voice, nil
}

// 上传私聊语音
func (m *QQClient) UploadPrivateRecord(uin uint64, voice *message.VoiceElement) (*message.VoiceElement, error) {
	if voice == nil || voice.Stream == nil {
		return nil, errors.New("voice is nil")
	}
	defer utils.CloseIO(voice.Stream)
	voice.IsGroup = false
	uid, err := m.GetUid(uin)
	if err != nil {
		return nil, err
	}
	pkt, err := oidb.BuildPrivateRecordUploadPacket(uid, voice)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := oidb.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	if md5, ext := oidb.NTV2RICH_MEDIA_RECORD.CommonGenerateHighwayExt(upload, voice.Stream, voice.Size, nil); ext != nil {
		//m.LOGD("group record upload ukey: %s", ukey)
		if err = m.highwayUpload(1007, voice.Stream, uint64(voice.Size), md5, ext); err != nil {
			return nil, err
		}
	}

	voice.MsgInfo = upload.MsgInfo
	voice.Uuid = upload.MsgInfo.MsgInfoBody[0].Index.FileUuid.Unwrap()
	voice.Compat = upload.CompatQMsg
	return voice, nil
}

// 获取私聊图片下载链接
func (m *QQClient) GetPrivateImageURL(node *pb_oidb.IndexNode) (string, error) {
	pkt, e := oidb.BuildPrivateImageDownloadPacket(m.session.Info.Uid, node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return oidb.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取群聊图片下载链接
func (m *QQClient) GetGroupImageURL(gin uint64, node *pb_oidb.IndexNode) (string, error) {
	pkt, e := oidb.BuildGroupImageDownloadPacket(gin, node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return oidb.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取私聊语音下载链接
func (m *QQClient) GetPrivateRecordURL(node *pb_oidb.IndexNode) (string, error) {
	pkt, e := oidb.BuildPrivateRecordDownloadPacket(m.session.Info.Uid, node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return oidb.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取群聊语音下载链接
func (m *QQClient) GetGroupRecordURL(gin uint64, node *pb_oidb.IndexNode) (string, error) {
	pkt, e := oidb.BuildGroupRecordDownloadPacket(gin, node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return oidb.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取私聊视频下载链接
func (m *QQClient) GetPrivateVideoURL(node *pb_oidb.IndexNode) (string, error) {
	pkt, e := oidb.BuildPrivateVideoDownloadPacket(m.session.Info.Uid, node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return oidb.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 获取群聊视频下载链接
func (m *QQClient) GetGroupVideoURL(gin uint64, node *pb_oidb.IndexNode) (string, error) {
	pkt, e := oidb.BuildGroupVideoDownloadPacket(gin, node)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return oidb.ParseNTv2RichMediaRspDownload(pkt.Data)
}

// 上传群聊文件
func (m *QQClient) UploadGroupFile(gin uint64, file *message.FileElement, parent_dir string) (*message.FileElement, error) {
	if file == nil || file.FileStream == nil {
		return nil, errors.New("element type is not group file")
	}

	file.FileName = resolveFileName(file.FileStream, file.FileName)
	pkt, err := oidb.BuildGroupFSUploadPacket(gin, file.FileStream, file.FileSize, file.FileName, parent_dir, file.FileSha1, file.FileMd5)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := oidb.ParseGroupFSUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	if !upload.FileExist.Unwrap() {
		ext, err := build_highway_file_ex(m.session.Info.Uin, gin, 0,
			&operation.ExcitingFileEntry{
				FileSize:  proto.Some(int64(file.FileSize)),
				Md5:       file.FileMd5,
				CheckKey:  upload.CheckKey,
				Md510M:    file.FileMd5,
				FileId:    upload.FileId,
				UploadKey: upload.FileKey,
			}, file.FileName, upload.UploadIp.Unwrap(), upload.UploadPort.Unwrap())
		if err != nil {
			return nil, err
		}
		if err = m.highwayUpload(71, file.FileStream, file.FileSize, file.FileMd5, ext); err != nil {
			return nil, err
		}
	}
	if pkt, err = oidb.BuildGroupFileSendPacket(gin, upload.FileId.Unwrap(), crypto.RandU32()); err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	if err = oidb.ParseGroupFileSendPacket(pkt.Data); err != nil {
		return nil, err
	}
	// king ??? lglv2 ret upload.FileId.Unwrap()
	return file, nil
}

// 上传私聊文件
func (m *QQClient) UploadPrivateFile(uin uint64, file *message.FileElement) (*message.FileElement, error) {
	if file == nil || file.FileStream == nil {
		return nil, errors.New("element type is not file")
	}
	uid, err := m.GetUid(uin)
	if err != nil {
		return nil, err
	}
	pkt, err := oidb.BuildPrivateFSUploadPacket(m.session.Info.Uid, uid, file)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := oidb.ParsePrivateFSUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	if !upload.FileExist.Unwrap() {
		ext, err := build_highway_file_ex(m.session.Info.Uin, 0, 1,
			&operation.ExcitingFileEntry{
				FileSize:  proto.Some(int64(file.FileSize)),
				Md5:       file.FileMd5,
				CheckKey:  file.FileSha1,
				Md510M:    file.FileMd5,
				FileId:    upload.Uuid,
				UploadKey: upload.MediaPlatformUploadKey,
			}, file.FileName, upload.UploadIp.Unwrap(), upload.UploadPort.Unwrap())
		if err != nil {
			return nil, err
		}
		if err = m.highwayUpload(95, file.FileStream, file.FileSize, file.FileMd5, ext); err != nil {
			return nil, err
		}
	}

	file.FileHash = upload.FileIdCrc.Unwrap()
	file.FileUuid = upload.Uuid.Unwrap()
	return file, nil
}

func build_highway_file_ex(bot_uin, gin uint64, field200 int32, entry *operation.ExcitingFileEntry, file_name string, ip string, port uint32) ([]byte, error) {
	return proto.Marshal(&operation.FileUploadExt{
		Unknown1: proto.Some[int32](100),
		Unknown2: proto.Some[int32](1),
		//Unknown200: field200,
		Entry: &operation.FileUploadEntry{
			BusiBuff: &operation.ExcitingBusiInfo{
				SenderUin:   proto.Some(int64(bot_uin)),
				ReceiverUin: proto.Some(int64(gin)),
				GroupCode:   proto.Some(int64(gin)),
			},
			ClientInfo: &operation.ExcitingClientInfo{
				ClientType:   proto.Some[int32](3),
				AppId:        proto.Some("100"),
				TerminalType: proto.Some[int32](3),
				ClientVer:    proto.Some("1.1.1"),
				Unknown:      proto.Some[int32](4),
			},
			FileEntry:    entry,
			FileNameInfo: &operation.ExcitingFileNameInfo{FileName: proto.Some(file_name)},
			Host: &operation.ExcitingHostConfig{
				Hosts: []*operation.ExcitingHostInfo{{
					Port: proto.Some[uint32](port),
					Url:  &operation.ExcitingUrlInfo{Unknown: proto.Some[int32](1), Host: proto.Some(ip)},
				}}}},
	})
}

// 获取私聊文件下载链接
func (m *QQClient) GetPrivateFileURL(file_uuid string, file_hash string) (string, error) {
	pkt, e := oidb.BuildPrivateFSDownloadPacket(m.session.Info.Uid, file_uuid, file_hash)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return oidb.ParsePrivateFSDownloadPacket(pkt.Data)
}

// 获取群聊文件下载链接
func (m *QQClient) GetGroupFileURL(groupUin uint64, file_id string) (string, error) {
	pkt, e := oidb.BuildGroupFSDownloadPacket(groupUin, file_id)
	if e != nil {
		return "", e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return "", e
	}
	return oidb.ParseGroupFSDownloadPacket(pkt.Data)
}

// 移动群文件
func (m *QQClient) MoveGroupFile(gin uint64, file_id string, parent_dir, target_dir string) error {
	pkt, e := oidb.BuildGroupFSMovePacket(gin, file_id, parent_dir, target_dir)
	if e != nil {
		return e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return e
	}
	return oidb.ParseGroupFSMovePacket(pkt.Data)
}

// 删除群文件
func (m *QQClient) DeleteGroupFile(gin uint64, file_id string) error {
	pkt, e := oidb.BuildGroupFSDeletePacket(gin, file_id)
	if e != nil {
		return e
	}
	if pkt, e = m.sendOidbPacketAndWait(pkt); e != nil {
		return e
	}
	return oidb.ParseGroupFSDeletePacket(pkt.Data)
}
