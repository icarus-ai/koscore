package entity

import "github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"

// Bot 表情条目
// qSid 表情 ID
// qDes 表情描述
type BotFaceDetail struct {
	QSid             string
	QDes             string
	EmCode           string
	QCid             uint32
	AniStickerType   uint32
	AniStickerPackId uint32
	AniStickerId     uint32
	BaseUrl          string
	AdvUrl           string
	EmojiNameAlias   []string
	AniStickerWidth  uint32
	AniStickerHeight uint32
}

func ToBotFaceEntry(e *oidb.FetchFaceDetailsEmoji) BotFaceDetail {
	return BotFaceDetail{
		QSid:             e.QSid,
		QDes:             e.QDes,
		EmCode:           e.EmCode,
		QCid:             e.QCid,
		AniStickerType:   e.AniStickerType,
		AniStickerPackId: e.AniStickerPackId,
		AniStickerId:     e.AniStickerId,
		BaseUrl:          e.Url.BaseUrl,
		AdvUrl:           e.Url.AdvUrl,
		EmojiNameAlias:   e.EmojiNameAlias,
		AniStickerWidth:  e.AniStickerWidth,
		AniStickerHeight: e.AniStickerHeight,
	}
}
