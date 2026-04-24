package oidb

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/message"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/utils/exception"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildGroupFileSendPacket(gin uint64, fileid string, random uint32) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x6D9, 4, &oidb.D6D9ReqBody{
		FeedsInfoReq: &oidb.FeedsReqBody{
			GroupCode: proto.Some(gin),
			AppId:     proto.Some[uint32](2),
			FeedsInfoList: []*message.FeedsInfo{{
				BusId:     proto.Some[uint32](102),
				FileId:    proto.Some(fileid),
				MsgRandom: proto.Some[uint32](random),
				FeedFlag:  proto.Some[uint32](1),
			}},
		},
	}, false, false)
}

func ParseGroupFileSendPacket(data []byte) error {
	rsp, e := ParseOidbPacket[oidb.D6D9RspBody](data)
	if e != nil {
		return e
	}
	if rsp.FeedsInfoRsp.RetCode.Unwrap() == 0 {
		return nil
	}
	return exception.NewFormat("%s (%d)", rsp.FeedsInfoRsp.RetMsg.Unwrap(), rsp.FeedsInfoRsp.RetCode.Unwrap())
}

func BuildGroupFSUploadPacket(gin uint64, fstream io.ReadSeeker, file_size uint64, file_name, parent_dir string, sha1, md5 []byte) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x6D6, 0, &oidb.D6D6ReqBody{
		UploadFileReq: &oidb.UploadFileReqBody{
			GroupCode:      proto.Some(gin),
			AppId:          proto.Some[uint32](7),
			BusId:          proto.Some[uint32](102),
			Entrance:       proto.Some[uint32](6),
			ParentFolderId: proto.Some(parent_dir),
			FileName:       proto.Some(file_name),
			LocalPath:      proto.Some("/" + file_name),
			FileSize:       proto.Some(file_size),
			Sha:            sha1,
			// king ?? Sha3: crypto.ComputeBlockSha1(fstream, highway.FileSize),
			Md5: md5,
		},
	}, false, false)
}

func ParseGroupFSUploadPacket(data []byte) (*oidb.UploadFileRspBody, error) {
	rsp, e := ParseOidbPacket[oidb.D6D6RspBody](data)
	if e != nil {
		return nil, e
	}
	result := rsp.UploadFileRsp
	if result.RetCode.Unwrap() != 0 {
		return nil, exception.NewOperationExceptionCode(result.RetCode.Unwrap(), result.RetMsg.Unwrap())
	}
	return result, nil
}

func BuildGroupFSDownloadPacket(gin uint64, fileid string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x6D6, 2, &oidb.D6D6ReqBody{
		DownloadFileReq: &oidb.DownloadFileReqBody{
			GroupCode: proto.Some(gin),
			BusId:     proto.Some[uint32](102),
			AppId:     proto.Some[uint32](7),
			FileId:    proto.Some(fileid),
		},
	}, false, false)
}

func ParseGroupFSDownloadPacket(data []byte) (string, error) {
	rsp, e := ParseOidbPacket[oidb.D6D6RspBody](data)
	if e != nil {
		return "", e
	}
	result := rsp.DownloadFileRsp
	if result.RetCode.Unwrap() != 0 {
		return "", exception.NewOperationExceptionCode(result.RetCode.Unwrap(), result.RetMsg.Unwrap())
	}
	return fmt.Sprintf("https://%s:443/ftn_handler/%s/?fname=", result.DownloadIp.Unwrap(), hex.EncodeToString(result.DownloadUrl)), nil
}

func BuildGroupFSDeletePacket(gin uint64, fileid string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x6d6, 3, &oidb.D6D6ReqBody{
		DeleteFileReq: &oidb.DeleteFileReqBody{
			GroupCode: proto.Some(gin),
			BusId:     proto.Some[uint32](102),
			AppId:     proto.Some[uint32](7),
			FileId:    proto.Some(fileid),
		},
	}, false, false)
}

func ParseGroupFSDeletePacket(data []byte) error {
	rsp, e := ParseOidbPacket[oidb.D6D6RspBody](data)
	if e != nil {
		return e
	}
	result := rsp.DeleteFileRsp
	if result.RetCode.Unwrap() != 0 {
		return exception.NewOperationExceptionCode(result.RetCode.Unwrap(), result.RetMsg.Unwrap())
	}
	return nil
}

func BuildGroupFSMovePacket(gin uint64, fileid string, parent_dir, target_dir string) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x6d6, 5, &oidb.D6D6ReqBody{
		MoveFileReq: &oidb.MoveFileReqBody{
			GroupCode:      proto.Some(gin),
			BusId:          proto.Some[uint32](102),
			AppId:          proto.Some[uint32](7),
			FileId:         proto.Some(fileid),
			ParentFolderId: proto.Some(parent_dir),
			DestFolderId:   proto.Some(target_dir),
		},
	}, false, false)
}

func ParseGroupFSMovePacket(data []byte) error {
	rsp, e := ParseOidbPacket[oidb.D6D6RspBody](data)
	if e != nil {
		return e
	}
	result := rsp.MoveFileRsp
	if result.RetCode.Unwrap() != 0 {
		return exception.NewOperationExceptionCode(result.RetCode.Unwrap(), result.RetMsg.Unwrap())
	}
	return nil
}
