package message

import (
	"errors"
	"fmt"

	pkt_oidb "github.com/kernel-ai/koscore/client/packets/oidb"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/message"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildPrivateFSUploadPacket(self_uid, target_uid string, file *message.FileElement) (*sso_type.SsoPacket, error) {
	md510MCheckSum, _ := crypto.ComputeMd5AndLengthWithLimit(file.FileStream, 10*1024*1024)
	return pkt_oidb.BuildOidbPacket(0xE37, 1700, &operation.OfflineFileUploadRequest{
		Command: proto.Some[uint32](1700),
		Seq:     proto.Some[int32](0),
		Upload: &operation.ApplyUploadReqV3{
			SenderUid:      proto.Some(self_uid),
			ReceiverUid:    proto.Some(target_uid),
			FileSize:       proto.Some(uint32(file.FileSize)),
			FileName:       proto.Some(file.FileName),
			Md510MCheckSum: md510MCheckSum,
			Sha1CheckSum:   file.FileSha1,
			LocalPath:      proto.Some("/"),
			Md5CheckSum:    file.FileMd5,
			//Sha3CheckSum  : []byte{},
			// king ?? Sha3: crypto.ComputeBlockSha1(fstream, highway.FileSize),
		},
		BusinessId:               proto.Some[int32](3),
		ClientType:               proto.Some[int32](1),
		FlagSupportMediaPlatform: proto.Some[int32](1),
	}, false, false)
}

func ParsePrivateFSUploadPacket(data []byte) (*operation.ApplyUploadRespV3, error) {
	rsp, e := pkt_oidb.ParseOidbPacket[operation.OfflineFileUploadResponse](data)
	if e != nil {
		return nil, e
	}
	upload := rsp.Upload
	if upload.RetCode.Unwrap() != 0 {
		return nil, fmt.Errorf("operation exception: %s (%d)", upload.RetMsg.Unwrap(), upload.RetCode.Unwrap())
	}
	return upload, nil
}

func BuildPrivateFSDownloadPacket(self_uid, uuid, hash string) (*sso_type.SsoPacket, error) {
	return pkt_oidb.BuildOidbPacket(0xE37, 1200, &operation.OidbSvcTrpcTcp0XE37_1200{
		SubCommand: 1200,
		Field2:     1,
		Field101:   3,
		Field102:   103,
		Field200:   1,
		Field99999: []byte{0xc0, 0x85, 0x2c, 0x01},
		Body: &operation.OidbSvcTrpcTcp0XE37_1200Body{
			ReceiverUid: self_uid,
			FileUuid:    uuid,
			Type:        2,
			FileHash:    hash,
			T2:          0,
		},
	}, false, false)
}

func ParsePrivateFSDownloadPacket(data []byte) (string, error) {
	rsp, e := pkt_oidb.ParseOidbPacket[operation.OidbSvcTrpcTcp0XE37_1200Response](data)
	if e != nil {
		return "", e
	}
	if rsp.Body == nil {
		return "", errors.New("文件失效")
	}
	if rsp.Body.Result == nil {
		return "", errors.New(rsp.Body.State)
	}
	return fmt.Sprintf("http://%s:%d%s&isthumb=0",
		rsp.Body.Result.Server,
		rsp.Body.Result.Port,
		rsp.Body.Result.Url,
	), nil
}
