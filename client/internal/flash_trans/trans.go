package flash_trans

import (
	"crypto/sha1"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/exception"
	"github.com/kernel-ai/koscore/utils/http"
	"github.com/kernel-ai/koscore/utils/proto"
)

const (
	BASE_URI   = "https://multimedia.qfile.qq.com/sliceupload"
	CHUNK_SIZE = 1048576 // 1024 * 1024
)

func UploadFile(appid uint32, ukey string, body *Stream) error {
	fn_make_min_chunk_size := func(a, b int64) []byte {
		if a < b {
			return make([]byte, a)
		}
		return make([]byte, b)
	}

	body.ToStart()
	hash_ctx := sha1.New()
	count := (body.Size + CHUNK_SIZE - 1) / CHUNK_SIZE // chunk_count
	req_body := &operation.FlashTransferUploadBody{
		UKey: proto.Some(ukey),
		Sha1StateV: &operation.FlashTransferSha1StateV{
			State: crypto.ComputeBlockSha1(body.R, CHUNK_SIZE),
		},
	}
	req := &operation.FlashTransferUploadReq{
		AppId:  proto.Some(appid),
		Field1: proto.Some[uint32](0),
		Field3: proto.Some[uint32](2),
		Body:   req_body,
	}
	req_head := map[string]string{
		"Accept":     "*/*",
		"Connection": "Keep-Alive",
		"Expect":     "100-continue",
	}

	body.ToStart()

	for i := int64(0); i < count; i++ {
		idx := i * CHUNK_SIZE // chunk_start
		data := fn_make_min_chunk_size(CHUNK_SIZE, body.Size-idx)
		n, _ := body.Read(idx, data)
		hash_ctx.Write(data)
		// upload_chunk(ukey string, start uint32, chunkSha1S *pb_msg.FlashTransferStateV, body []byte) error
		req_body.Start = proto.Some(uint32(idx))
		req_body.End = proto.Some(uint32(idx + int64(n) - 1))
		req_body.Sha1 = hash_ctx.Sum(nil)
		req_body.Body = data
		data, _ = proto.Marshal(req)
		if data, _ = http.Post(BASE_URI, data, req_head); data == nil {
			return exception.ErrEmptyRsp
		}
		rsp, e := proto.Unmarshal[operation.FlashTransferUploadResp](data)
		if e != nil {
			return e
		}
		if rsp.Status.Unwrap() != "success" {
			return exception.NewFormat("status: %s", rsp.Status.Unwrap())
		}
		hash_ctx.Reset()
	}
	return nil
}
