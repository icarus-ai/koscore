package oidb

import (
	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
)

func BuildFetchFaceDetailsPacket() (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x9154, 1, &oidb.FetchFaceDetailsReq{Field1: 0, Field2: 7, Field3: "0"}, false, false)
}

func ParseFetchFaceDetailsPacket(data []byte) ([]entity.BotFaceDetail, error) {
	rsp, e := ParseOidbPacket[oidb.FetchFaceDetailsResp](data)
	if e != nil {
		return nil, e
	}

	var details []entity.BotFaceDetail
	// 解析普通表情
	for _, pack := range rsp.CommonFace.EmojiList {
		for _, emoji := range pack.Detail {
			details = append(details, entity.ToBotFaceEntry(emoji))
		}
	}
	// 解析超大表情
	for _, pack := range rsp.SpecialBigFace.EmojiList {
		for _, big := range pack.Detail {
			details = append(details, entity.ToBotFaceEntry(big))
		}
	}
	// 解析魔法表情
	for _, magic := range rsp.SpecialMagicFace.Magic.EmojiList {
		details = append(details, entity.ToBotFaceEntry(magic))
	}
	return details, nil
}
