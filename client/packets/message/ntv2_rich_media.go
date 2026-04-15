package message

import (
	"encoding/hex"
	"errors"
	"io"

	"github.com/kernel-ai/koscore/client/internal/highway"
	pkt_oidb "github.com/kernel-ai/koscore/client/packets/oidb"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/audio"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
)

func (typ ntv2_rich_media) common_ConvertIPv4(ipv4s []*oidb.IPv4) *oidb.NTHighwayNetwork {
	ret := &oidb.NTHighwayNetwork{}
	for _, item := range ipv4s {
		ret.IPv4S = append(ret.IPv4S, &oidb.NTHighwayIPv4{
			Domain: &oidb.NTHighwayDomain{
				IsEnable: proto.TRUE,
				IP:       proto.Some(binary.UInt32ToIPV4Address(item.OutIP.Unwrap())),
			},
			Port: item.OutPort,
		})
	}
	return ret
}

func (typ ntv2_rich_media) CommonGenerateHighwayExt(upload *oidb.UploadResp, stream io.ReadSeeker, io_size uint32, subFileInfo *oidb.SubFileInfo) (md5, ext []byte) {
	if upload.UKey.Unwrap() == "" {
		return
	}
	if subFileInfo == nil {
		hash := &oidb.NTHighwayHash{}
		index := upload.MsgInfo.MsgInfoBody[0].Index
		if typ == NTV2RICH_MEDIA_VIdEO {
			io_size = highway.BlockSize
			hash.FileSha1 = crypto.ComputeBlockSha1(stream, highway.BlockSize)
		} else {
			sha, _ := hex.DecodeString(index.Info.FileSha1.Unwrap())
			hash.FileSha1 = [][]byte{sha}
		}
		md5, _ = hex.DecodeString(index.Info.FileHash.Unwrap())
		ext, _ = proto.Marshal(&oidb.NTV2RichMediaHighwayExt{
			FileUuid:    index.FileUuid,
			UKey:        upload.UKey,
			Network:     typ.common_ConvertIPv4(upload.IPv4S),
			MsgInfoBody: upload.MsgInfo.MsgInfoBody,
			BlockSize:   proto.Some(uint32(highway.BlockSize)),
			Hash:        hash,
		})
		return
	}
	if subFileInfo.UKey.Unwrap() != "" {
		return
	}
	index := upload.MsgInfo.MsgInfoBody[1].Index
	sha, _ := hex.DecodeString(index.Info.FileSha1.Unwrap())
	md5, _ = hex.DecodeString(index.Info.FileHash.Unwrap())
	ext, _ = proto.Marshal(&oidb.NTV2RichMediaHighwayExt{
		FileUuid:    index.FileUuid,
		UKey:        subFileInfo.UKey,
		Network:     typ.common_ConvertIPv4(subFileInfo.IPv4S),
		MsgInfoBody: upload.MsgInfo.MsgInfoBody,
		BlockSize:   proto.Some(uint32(highway.BlockSize)),
		Hash:        &oidb.NTHighwayHash{FileSha1: [][]byte{sha}},
	})
	return
}

type ntv2_rich_media uint8

const (
	NTV2RICH_MEDIA_IMAGE  ntv2_rich_media = 1
	NTV2RICH_MEDIA_VIdEO  ntv2_rich_media = 2
	NTV2RICH_MEDIA_RECORD ntv2_rich_media = 3
)

func (typ ntv2_rich_media) build_head(cmd uint32, id any) *oidb.MultiMediaReqHead {
	return &oidb.MultiMediaReqHead{
		Common: &oidb.CommonHead{
			RequestId: proto.Some[uint32](1),
			Command:   proto.Some(cmd),
		},
		Scene:  typ.build_scene(id),
		Client: &oidb.ClientMeta{AgentType: proto.Some[uint32](2)},
	}
}

func (typ ntv2_rich_media) build_scene(id any) *oidb.SceneInfo {
	scene := &oidb.SceneInfo{RequestType: proto.Some[uint32](2)}
	switch typ {
	case NTV2RICH_MEDIA_IMAGE:
		scene.BusinessType = proto.Some[uint32](1)
	case NTV2RICH_MEDIA_VIdEO:
		scene.BusinessType = proto.Some[uint32](2)
	case NTV2RICH_MEDIA_RECORD:
		scene.BusinessType = proto.Some[uint32](3)
	}
	switch o := id.(type) {
	case uint64:
		scene.SceneType = proto.Some[uint32](2)
		scene.Group = &oidb.GroupInfo{GroupUin: proto.Some(int64(o))}
	case string:
		scene.SceneType = proto.Some[uint32](1)
		scene.C2C = &oidb.C2CUserInfo{
			TargetUid:   proto.Some(o),
			AccountType: proto.Some[uint32](1),
		}
	}
	return scene
}

func (typ ntv2_rich_media) build_file(fs fstream) (*oidb.FileInfo, error) {
	md5str := hex.EncodeToString(fs.md5)

	info := &oidb.FileInfo{
		FileSize: proto.Some(fs.size),
		FileHash: proto.Some(md5str),
		FileSha1: proto.Some(hex.EncodeToString(fs.sha1)),
	}
	switch typ {
	case NTV2RICH_MEDIA_IMAGE:
		format, size, err := utils.ImageResolve(fs.stream)
		if err != nil {
			return nil, err
		}
		info.Type = &oidb.FileType{
			Type:      proto.Some[uint32](1),
			PicFormat: proto.Some(uint32(format)),
		}
		info.Width = proto.Some(uint32(size.Width))
		info.Height = proto.Some(uint32(size.Height))
		info.FileName = proto.Some(md5str + "." + format.String())
		info.Original = proto.Some[uint32](1)
	case NTV2RICH_MEDIA_VIdEO:
		// unable to determine video type, skip
		info.Type = &oidb.FileType{Type: proto.Some[uint32](2)}
		info.FileName = proto.Some(md5str + ".mp4") // default to mp4
	case NTV2RICH_MEDIA_RECORD:
		audio_info, err := audio.Decode(fs.stream)
		if err != nil {
			return nil, err
		}
		info.Type = &oidb.FileType{
			Type:        proto.Some[uint32](3),
			VideoFormat: proto.Some[uint32](1),
		}
		info.Time = proto.Some(uint32(audio_info.Time + 0.5))
		info.FileName = proto.Some(md5str + "." + audio_info.Type.String())
	}
	return info, nil
}

type fstream struct {
	stream   io.ReadSeeker
	size     uint32
	md5      []byte
	sha1     []byte
	sub_type uint32
}

func (typ ntv2_rich_media) build_upload(id any, ext *oidb.ExtBizInfo, fsinfo ...fstream) (*oidb.NTV2RichMediaReq, error) {
	var files []*oidb.UploadInfo
	file, err := typ.build_file(fsinfo[0])
	if err != nil {
		return nil, err
	}
	files = append(files, &oidb.UploadInfo{FileInfo: file, SubFileType: proto.Some(fsinfo[0].sub_type)})
	if len(fsinfo) == 2 { // video thumb sub_file_type 100
		if file, err = NTV2RICH_MEDIA_IMAGE.build_file(fsinfo[1]); err != nil {
			return nil, err
		}
		files = append(files, &oidb.UploadInfo{FileInfo: file, SubFileType: proto.Some(fsinfo[1].sub_type)})
	}

	head := typ.build_head(100, id)
	return &oidb.NTV2RichMediaReq{
		ReqHead: head,
		Upload: &oidb.UploadReq{
			UploadInfo:             files,
			TryFastUploadCompleted: proto.Some(true),
			SrvSendMsg:             proto.Some(false),
			ClientRandomId:         proto.Some(uint64(crypto.RandU32())),
			CompatQMsgSceneType:    head.Scene.SceneType,
			ExtBizInfo:             ext,
			ClientSeq:              proto.Some[uint32](10),
			NoNeedCompatMsg:        proto.Some(false),
		}}, nil
}

func (typ ntv2_rich_media) build_download(id any, node *oidb.IndexNode) (*oidb.NTV2RichMediaReq, error) {
	if node == nil {
		return nil, errors.New("node is null")
	}
	return &oidb.NTV2RichMediaReq{
		ReqHead: typ.build_head(100, id),
		Download: &oidb.DownloadReq{
			Node: node,
			Download: &oidb.DownloadExt{
				Pic:   &oidb.PicDownloadExt{},
				Video: &oidb.VideoDownloadExt{},
				Ptt:   &oidb.PttDownloadExt{},
			}}}, nil
}

// *****

func BuildGroupImageUploadPacket(gin uint64, image *message.ImageElement) (*sso_type.SsoPacket, error) {
	if image.Stream == nil {
		return nil, errors.New("image stream data is null")
	}
	body, e := NTV2RICH_MEDIA_IMAGE.build_upload(gin, &oidb.ExtBizInfo{
		Pic: &oidb.PicExtBizInfo{
			//BizType    : proto.Some(uint32(image.SubType)),
			TextSummary:       proto.Some(utils.Ternary(image.Summary == "" && image.SubType == 1, "[动画表情]", image.Summary)),
			BytesPbReserveC2C: []byte{0x08, 0x00, 0x18, 0x00, 0x20, 0x00, 0x4A, 0x00, 0x50, 0x00, 0x62, 0x00, 0x92, 0x01, 0x00, 0x9A, 0x01, 0x00, 0xAA, 0x01, 0x0C, 0x08, 0x00, 0x12, 0x00, 0x18, 0x00, 0x20, 0x00, 0x28, 0x00, 0x3A, 0x00},
		},
		Video: &oidb.VideoExtBizInfo{BytesPbReserve: binary.Empty},
		Ptt:   &oidb.PttExtBizInfo{BytesReserve: binary.Empty, BytesPbReserve: binary.Empty, BytesGeneralFlags: binary.Empty},
	}, fstream{stream: image.Stream, size: image.Size, md5: image.Md5, sha1: image.Sha1})
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x11C4, 100, body, false, false)
}

func BuildPrivateImageUploadPacket(self_uid string, image *message.ImageElement) (*sso_type.SsoPacket, error) {
	if image.Stream == nil {
		return nil, errors.New("image stream data is null")
	}
	body, e := NTV2RICH_MEDIA_IMAGE.build_upload(self_uid, &oidb.ExtBizInfo{
		Pic: &oidb.PicExtBizInfo{
			//BizType    : proto.Some(uint32(image.SubType)),
			TextSummary:       proto.Some(utils.Ternary(image.Summary == "" && image.SubType == 1, "[动画表情]", image.Summary)),
			BytesPbReserveC2C: []byte{0x08, 0x00, 0x18, 0x00, 0x20, 0x00, 0x42, 0x00, 0x50, 0x00, 0x62, 0x00, 0x92, 0x01, 0x00, 0x9A, 0x01, 0x00, 0xA2, 0x01, 0x0C, 0x08, 0x00, 0x12, 0x00, 0x18, 0x00, 0x20, 0x00, 0x28, 0x00, 0x3A, 0x00},
		},
		Video: &oidb.VideoExtBizInfo{BytesPbReserve: binary.Empty},
		Ptt:   &oidb.PttExtBizInfo{BytesReserve: binary.Empty, BytesPbReserve: binary.Empty, BytesGeneralFlags: binary.Empty},
	}, fstream{stream: image.Stream, size: image.Size, md5: image.Md5, sha1: image.Sha1})
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x11C5, 100, body, false, false)
}

func BuildGroupVideoUploadPacket(gin uint64, video *message.ShortVideoElement) (*sso_type.SsoPacket, error) {
	if video.Stream == nil {
		return nil, errors.New("video stream data is null")
	}
	if video.Thumb.Stream == nil {
		return nil, errors.New("video thumb stream data is null")
	}
	body, e := NTV2RICH_MEDIA_VIdEO.build_upload(gin, &oidb.ExtBizInfo{
		Pic: &oidb.PicExtBizInfo{
			BizType:     proto.Some[uint32](0),
			TextSummary: proto.Some(""),
		},
		Video: &oidb.VideoExtBizInfo{BytesPbReserve: []byte{0x80, 0x01, 0x00}},
		Ptt:   &oidb.PttExtBizInfo{BytesReserve: binary.Empty, BytesPbReserve: binary.Empty, BytesGeneralFlags: binary.Empty},
	},
		fstream{stream: video.Stream, size: video.Size, md5: video.Md5, sha1: video.Sha1},
		fstream{stream: video.Thumb.Stream, size: video.Thumb.Size, md5: video.Thumb.Md5, sha1: video.Thumb.Sha1, sub_type: 100},
	)
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x11EA, 100, body, false, false)
}

func BuildPrivateVideoUploadPacket(self_uid string, video *message.ShortVideoElement) (*sso_type.SsoPacket, error) {
	if video.Stream == nil {
		return nil, errors.New("video stream data is null")
	}
	if video.Thumb.Stream == nil {
		return nil, errors.New("video thumb stream data is null")
	}
	body, e := NTV2RICH_MEDIA_VIdEO.build_upload(self_uid, &oidb.ExtBizInfo{
		Pic: &oidb.PicExtBizInfo{
			BizType:     proto.Some[uint32](0),
			TextSummary: proto.Some(""),
		},
		Video: &oidb.VideoExtBizInfo{BytesPbReserve: []byte{0x80, 0x01, 0x00}},
		Ptt:   &oidb.PttExtBizInfo{BytesReserve: binary.Empty, BytesPbReserve: binary.Empty, BytesGeneralFlags: binary.Empty},
	},
		fstream{stream: video.Stream, size: video.Size, md5: video.Md5, sha1: video.Sha1},
		fstream{stream: video.Thumb.Stream, size: video.Thumb.Size, md5: video.Thumb.Md5, sha1: video.Thumb.Sha1, sub_type: 100},
	)
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x11E9, 100, body, false, false)
}

func BuildGroupRecordUploadPacket(gin uint64, voice *message.VoiceElement) (*sso_type.SsoPacket, error) {
	if voice.Stream == nil {
		return nil, errors.New("voice stream data is null")
	}
	body, e := NTV2RICH_MEDIA_RECORD.build_upload(gin, &oidb.ExtBizInfo{
		Pic: &oidb.PicExtBizInfo{
			//BizType    : proto.Some[uint32](0),
			TextSummary: proto.Some(""),
		},
		Video: &oidb.VideoExtBizInfo{BytesPbReserve: binary.Empty},
		Ptt: &oidb.PttExtBizInfo{
			BytesReserve:      []byte{0x08, 0x00, 0x38, 0x00},
			BytesPbReserve:    binary.Empty,
			BytesGeneralFlags: []byte{0x9a, 0x01, 0x07, 0xaa, 0x03, 0x04, 0x08, 0x08, 0x12, 0x00},
		},
	}, fstream{stream: voice.Stream, size: voice.Size, md5: voice.Md5, sha1: voice.Sha1})
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x126E, 100, body, false, false)
}

func BuildPrivateRecordUploadPacket(self_uid string, voice *message.VoiceElement) (*sso_type.SsoPacket, error) {
	if voice.Stream == nil {
		return nil, errors.New("voice stream data is null")
	}
	body, e := NTV2RICH_MEDIA_RECORD.build_upload(self_uid, &oidb.ExtBizInfo{
		Pic: &oidb.PicExtBizInfo{
			//BizType    : proto.Some[uint32](0),
			TextSummary: proto.Some(""),
		},
		Video: &oidb.VideoExtBizInfo{BytesPbReserve: binary.Empty},
		Ptt: &oidb.PttExtBizInfo{
			BytesReserve:      []byte{0x08, 0x00, 0x38, 0x00},
			BytesPbReserve:    binary.Empty,
			BytesGeneralFlags: []byte{0x9a, 0x01, 0x0b, 0xaa, 0x03, 0x08, 0x08, 0x04, 0x12, 0x04, 0x00, 0x00, 0x00, 0x00},
		},
	}, fstream{stream: voice.Stream, size: voice.Size, md5: voice.Md5, sha1: voice.Sha1})
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x126D, 100, body, false, false)
}

// new ImageGroupUploadEventResp(response.Upload.MsgInfo, response.Upload.CompatQMsg, Common.GenerateExt(response.Upload)))
// new ImageUploadEventResp     (response.Upload.MsgInfo, response.Upload.CompatQMsg, Common.GenerateExt(response.Upload)))
// new VideoUploadEventResp     (response.Upload.MsgInfo, response.Upload.CompatQMsg, Common.GenerateExt(response.Upload), Common.GenerateExt(response.Upload, response.Upload.SubFileInfos[0])));
// new VideoGroupUploadEventResp(response.Upload.MsgInfo, response.Upload.CompatQMsg, Common.GenerateExt(response.Upload), Common.GenerateExt(response.Upload, response.Upload.SubFileInfos[0])));
// new RecordUploadEventResp     (response.Upload.MsgInfo, response.Upload.CompatQMsg, Common.GenerateExt(response.Upload)));
// new RecordGroupUploadEventResp(response.Upload.MsgInfo, response.Upload.CompatQMsg, Common.GenerateExt(response.Upload)));

func ParseNTv2RichMediaUploadPacket(data []byte) (*oidb.UploadResp, error) {
	rsp, e := pkt_oidb.ParseOidbPacket[oidb.NTV2RichMediaResp](data)
	if e != nil {
		return nil, e
	}
	if rsp.Upload == nil {
		return nil, errors.New("NTV2RichMediaResp.Upload nil")
	}
	return rsp.Upload, nil
}

// *****

func BuildGroupImageDownloadPacket(gin uint64, node *oidb.IndexNode) (*sso_type.SsoPacket, error) {
	if node == nil {
		return nil, errors.New("image node data is null")
	}
	body, e := NTV2RICH_MEDIA_IMAGE.build_download(gin, node)
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x11C4, 200, body, false, false)
}

func BuildPrivateImageDownloadPacket(self_uid string, node *oidb.IndexNode) (*sso_type.SsoPacket, error) {
	if node == nil {
		return nil, errors.New("image node data is null")
	}
	body, e := NTV2RICH_MEDIA_IMAGE.build_download(self_uid, node)
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x11C5, 200, body, false, false)
}

func BuildGroupVideoDownloadPacket(gin uint64, node *oidb.IndexNode) (*sso_type.SsoPacket, error) {
	if node == nil {
		return nil, errors.New("video node data is null")
	}
	body, e := NTV2RICH_MEDIA_VIdEO.build_download(gin, node)
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x11EA, 200, body, false, false)
}

func BuildPrivateVideoDownloadPacket(self_uid string, node *oidb.IndexNode) (*sso_type.SsoPacket, error) {
	if node == nil {
		return nil, errors.New("video node data is null")
	}
	body, e := NTV2RICH_MEDIA_VIdEO.build_download(self_uid, node)
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x11E9, 200, body, false, false)
}

func BuildGroupRecordDownloadPacket(gin uint64, node *oidb.IndexNode) (*sso_type.SsoPacket, error) {
	if node == nil {
		return nil, errors.New("voice node data is null")
	}
	body, e := NTV2RICH_MEDIA_RECORD.build_download(gin, node)
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x126E, 200, body, false, false)
}

func BuildPrivateRecordDownloadPacket(self_uid string, node *oidb.IndexNode) (*sso_type.SsoPacket, error) {
	if node == nil {
		return nil, errors.New("voice node data is null")
	}
	body, e := NTV2RICH_MEDIA_RECORD.build_download(self_uid, node)
	if e != nil {
		return nil, e
	}
	return pkt_oidb.BuildOidbPacket(0x126D, 200, body, false, false)
}

func ParseNTv2RichMediaRspDownload(data []byte) (string, error) {
	rsp, e := pkt_oidb.ParseOidbPacket[oidb.NTV2RichMediaResp](data)
	if e != nil {
		return "", e
	}
	body := rsp.Download
	return "https://" + body.Info.Domain.Unwrap() + body.Info.UrlPath.Unwrap() + body.RKeyParam.Unwrap(), nil
}
