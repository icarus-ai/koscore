package websso

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/utils"
)

type web_rsp struct {
	ErrorCode int32 `json:"ErrorCode"`
	BlockList []struct {
		Uin         uint64 `json:"uint64_uin"`
		NickBytes   string `json:"bytes_nick"`
		Age         uint32 `json:"uint32_age"`
		Sex         uint32 `json:"uint32_sex"`
		SourceBytes string `json:"bytes_source"`
		Uid         string `json:"str_uid"`
	} `json:"rpt_block_list"`
}

func ParseUnidirectionalFriendsPacket(data []byte) ([]*entity.User, error) {
	var rsp web_rsp
	if err := json.Unmarshal(data, &rsp); err != nil {
		return nil, errors.Wrap(err, "unmarshal json error")
	}
	if rsp.ErrorCode != 0 {
		return nil, fmt.Errorf("web sso request error: %v", rsp.ErrorCode)
	}

	decodeBase64String := func(str string) string {
		b, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return ""
		}
		return utils.B2S(b)
	}
	ret := make([]*entity.User, 0, len(rsp.BlockList))
	for _, block := range rsp.BlockList {
		ret = append(ret, &entity.User{
			Uin:      block.Uin,
			Uid:      block.Uid,
			Nickname: decodeBase64String(block.NickBytes),
			Age:      block.Age,
			Source:   decodeBase64String(block.SourceBytes),
		})
	}
	return ret, nil
}
