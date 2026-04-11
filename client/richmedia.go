package client

import (
	"errors"

	pkt_msg "github.com/kernel-ai/koscore/client/packets/message"
	pb_msg "github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	pb_hw "github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"

	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
)

func (m *QQClient) UploadGroupFile(gin uint64, file *message.FileElement, parent_dir string) (*message.FileElement, error) {
	if file == nil || file.FileStream == nil {
		return nil, errors.New("element type is not group file")
	}

	file.FileName = resolveFileName(file.FileStream, file.FileName)
	pkt, err := pkt_msg.BuildGroupFSUploadPacket(gin, file.FileStream, file.FileSize, file.FileName, parent_dir, file.FileSha1, file.FileMd5)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := pkt_msg.ParseGroupFSUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	if !upload.FileExist.Unwrap() {
		ext, err := build_highway_file_ex(m.UIN(), gin, 0,
			&pb_hw.ExcitingFileEntry{
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
	if pkt, err = pkt_msg.BuildGroupFileSendPacket(gin, upload.FileId.Unwrap(), crypto.RandU32()); err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	if err = pkt_msg.ParseGroupFileSendPacket(pkt.Data); err != nil {
		return nil, err
	}
	// king ??? lglv2 ret upload.FileId.Unwrap()
	return file, nil
}

func (m *QQClient) UploadPrivateFile(uin uint64, file *message.FileElement) (*message.FileElement, error) {
	if file == nil || file.FileStream == nil {
		return nil, errors.New("element type is not file")
	}
	pkt, err := pkt_msg.BuildPrivateFSUploadPacket(m.Uid(), m.GetUid(uin), file)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := pkt_msg.ParsePrivateFSUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	if !upload.FileExist.Unwrap() {
		ext, err := build_highway_file_ex(m.UIN(), 0, 1,
			&pb_hw.ExcitingFileEntry{
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
	file.FileUUid = upload.Uuid.Unwrap()
	return file, nil
}

/*

   public async Task<string> GroupFSDownload(long groupUin, string fileId)
   {
       var request = new GroupFSDownloadEventReq(groupUin, fileId);
       var response = await context.EventContext.SendEvent<GroupFSDownloadEventResp>(request);
       return response.FileUrl;
   }

   public async Task GroupFSMove(long groupUin, string fileId, string parentDirectory, string targetDirectory) => await context.EventContext.SendEvent<GroupFSMoveEventResp>(new GroupFSMoveEventReq(groupUin, fileId, parentDirectory, targetDirectory));

   public async Task GroupFSDelete(long groupUin, string fileId) => await context.EventContext.SendEvent<GroupFSDeleteEventResp>(new GroupFSDeleteEventReq(groupUin, fileId));

*/

// *****

func build_highway_file_ex(bot_uin, gin uint64, field200 int32, entry *pb_hw.ExcitingFileEntry, file_name string, ip string, port uint32) ([]byte, error) {
	return proto.Marshal(&pb_hw.FileUploadExt{
		Unknown1: proto.Some[int32](100),
		Unknown2: proto.Some[int32](1),
		//Unknown200: field200,
		Entry: &pb_hw.FileUploadEntry{
			BusiBuff: &pb_hw.ExcitingBusiInfo{
				SenderUin:   proto.Some(int64(bot_uin)),
				ReceiverUin: proto.Some(int64(gin)),
				GroupCode:   proto.Some(int64(gin)),
			},
			ClientInfo: &pb_hw.ExcitingClientInfo{
				ClientType:   proto.Some[int32](3),
				AppId:        proto.Some("100"),
				TerminalType: proto.Some[int32](3),
				ClientVer:    proto.Some("1.1.1"),
				Unknown:      proto.Some[int32](4),
			},
			FileEntry:    entry,
			FileNameInfo: &pb_hw.ExcitingFileNameInfo{FileName: proto.Some(file_name)},
			Host: &pb_hw.ExcitingHostConfig{
				Hosts: []*pb_hw.ExcitingHostInfo{{
					Port: proto.Some[uint32](port),
					Url:  &pb_hw.ExcitingUrlInfo{Unknown: proto.Some[int32](1), Host: proto.Some(ip)},
				}}}},
	})
}

// ***** ***** ***** ***** *****

func (m *QQClient) UploadGroupImage(gin uint64, image *message.ImageElement) (*message.ImageElement, error) {
	if image == nil || image.Stream == nil {
		return nil, errors.New("image is nil")
	}
	defer utils.CloseIO(image.Stream)
	image.IsGroup = true
	pkt, err := pkt_msg.BuildGroupImageUploadPacket(gin, image)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := pkt_msg.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}
	if md5, ext := pkt_msg.NTV2RICH_MEDIA_IMAGE.CommonGenerateHighwayExt(upload, image.Stream, image.Size, nil); ext != nil {
		//m.LOGD("group image upload ukey: %s", ukey)
		if err = m.highwayUpload(1004, image.Stream, uint64(image.Size), md5, ext); err != nil {
			return nil, err
		}
	}
	image.CompatFace, _ = proto.Unmarshal[pb_msg.CustomFace](upload.CompatQMsg)
	image.MsgInfo = upload.MsgInfo
	image.FileUUid = upload.MsgInfo.MsgInfoBody[0].Index.FileUuid.Unwrap()
	return image, nil
}

func (m *QQClient) UploadPrivateImage(uin uint64, image *message.ImageElement) (*message.ImageElement, error) {
	if image == nil || image.Stream == nil {
		return nil, errors.New("image is nil")
	}
	defer utils.CloseIO(image.Stream)
	image.IsGroup = false
	pkt, err := pkt_msg.BuildPrivateImageUploadPacket(m.GetUid(uin), image)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := pkt_msg.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}
	if md5, ext := pkt_msg.NTV2RICH_MEDIA_IMAGE.CommonGenerateHighwayExt(upload, image.Stream, image.Size, nil); ext != nil {
		//m.LOGD("private image upload ukey: %s", ukey)
		if err = m.highwayUpload(1003, image.Stream, uint64(image.Size), md5, ext); err != nil {
			return nil, err
		}
	}
	image.CompatFace, _ = proto.Unmarshal[pb_msg.CustomFace](upload.CompatQMsg)
	image.MsgInfo = upload.MsgInfo
	image.FileUUid = upload.MsgInfo.MsgInfoBody[0].Index.FileUuid.Unwrap()
	return image, nil
}

func (m *QQClient) UploadGroupShortVideo(gin uint64, video *message.ShortVideoElement) (*message.ShortVideoElement, error) {
	if video == nil || video.Stream == nil {
		return nil, errors.New("video is nil")
	}
	if video.Thumb == nil || video.Thumb.Stream == nil {
		return nil, errors.New("video thumb is nil")
	}
	defer utils.CloseIO(video.Stream)
	defer utils.CloseIO(video.Thumb.Stream)
	pkt, err := pkt_msg.BuildGroupVideoUploadPacket(gin, video)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := pkt_msg.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	// video
	if md5, ext := pkt_msg.NTV2RICH_MEDIA_VIdEO.CommonGenerateHighwayExt(upload, video.Stream, video.Size, nil); ext != nil {
		//m.LOGD("group video upload ukey: %s", ukey)
		if err = m.highwayUpload(1005, video.Stream, uint64(video.Size), md5, ext); err != nil {
			return nil, err
		}
	}
	// thumb
	if md5, ext := pkt_msg.NTV2RICH_MEDIA_IMAGE.CommonGenerateHighwayExt(upload, video.Thumb.Stream, video.Thumb.Size, upload.SubFileInfos[0]); ext != nil {
		//m.LOGD("group video.thumb upload ukey: %s", ukey)
		if err = m.highwayUpload(1006, video.Thumb.Stream, uint64(video.Thumb.Size), md5, ext); err != nil {
			return nil, err
		}
	}

	video.Compat, _ = proto.Unmarshal[pb_msg.VideoFile](upload.CompatQMsg)
	video.MsgInfo = upload.MsgInfo
	video.Name = video.Compat.FileName.Unwrap()
	video.UUid = video.Compat.FileUuid.Unwrap()
	return video, nil
}

func (m *QQClient) UploadPrivateShortVideo(uin uint64, video *message.ShortVideoElement) (*message.ShortVideoElement, error) {
	if video == nil || video.Stream == nil {
		return nil, errors.New("video is nil")
	}
	if video.Thumb == nil || video.Thumb.Stream == nil {
		return nil, errors.New("video thumb is nil")
	}
	defer utils.CloseIO(video.Stream)
	defer utils.CloseIO(video.Thumb.Stream)
	pkt, err := pkt_msg.BuildPrivateVideoUploadPacket(m.GetUid(uin), video)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := pkt_msg.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	// video
	if md5, ext := pkt_msg.NTV2RICH_MEDIA_VIdEO.CommonGenerateHighwayExt(upload, video.Stream, video.Size, nil); ext != nil {
		//m.LOGD("pivate video upload ukey: %s", ukey)
		if err = m.highwayUpload(1001, video.Stream, uint64(video.Size), md5, ext); err != nil {
			return nil, err
		}
	}
	// thumb
	if md5, ext := pkt_msg.NTV2RICH_MEDIA_IMAGE.CommonGenerateHighwayExt(upload, video.Thumb.Stream, video.Thumb.Size, upload.SubFileInfos[0]); ext != nil {
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
	video.UUid = video.Compat.FileUuid.Unwrap()
	return video, nil
}

func (m *QQClient) UploadGroupRecord(gin uint64, voice *message.VoiceElement) (*message.VoiceElement, error) {
	if voice == nil || voice.Stream == nil {
		return nil, errors.New("voice is nil")
	}
	defer utils.CloseIO(voice.Stream)
	pkt, err := pkt_msg.BuildGroupRecordUploadPacket(gin, voice)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := pkt_msg.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	if md5, ext := pkt_msg.NTV2RICH_MEDIA_RECORD.CommonGenerateHighwayExt(upload, voice.Stream, voice.Size, nil); ext != nil {
		//m.LOGD("group record upload ukey: %s", ukey)
		if err = m.highwayUpload(1008, voice.Stream, uint64(voice.Size), md5, ext); err != nil {
			return nil, err
		}
	}

	voice.MsgInfo = upload.MsgInfo
	voice.UUid = upload.MsgInfo.MsgInfoBody[0].Index.FileUuid.Unwrap()
	voice.Compat = upload.CompatQMsg
	return voice, nil
}

func (m *QQClient) UploadPrivateRecord(uin uint64, voice *message.VoiceElement) (*message.VoiceElement, error) {
	if voice == nil || voice.Stream == nil {
		return nil, errors.New("voice is nil")
	}
	defer utils.CloseIO(voice.Stream)
	pkt, err := pkt_msg.BuildPrivateRecordUploadPacket(m.GetUid(uin), voice)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	upload, err := pkt_msg.ParseNTv2RichMediaUploadPacket(pkt.Data)
	if err != nil {
		return nil, err
	}

	if md5, ext := pkt_msg.NTV2RICH_MEDIA_RECORD.CommonGenerateHighwayExt(upload, voice.Stream, voice.Size, nil); ext != nil {
		//m.LOGD("group record upload ukey: %s", ukey)
		if err = m.highwayUpload(1007, voice.Stream, uint64(voice.Size), md5, ext); err != nil {
			return nil, err
		}
	}

	voice.MsgInfo = upload.MsgInfo
	voice.UUid = upload.MsgInfo.MsgInfoBody[0].Index.FileUuid.Unwrap()
	voice.Compat = upload.CompatQMsg
	return voice, nil
}
