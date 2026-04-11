package oidb

import (
	"encoding/hex"
	"errors"

	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/oidb"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/message"
)

func BuildAiCharacterListService(groupUin uint64, chatType entity.ChatType) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x929D, 0, &oidb.OidbSvcTrpcTcp0X929D_0_Req{GroupUin: groupUin, ChatType: uint32(chatType)}, false, false)
}

func ParseAiCharacterListService(data []byte) (*entity.AiCharacterList, error) {
	rsp, err := ParseOidbPacket[oidb.OidbSvcTrpcTcp0X929D_0_Rsp](data)
	if err != nil {
		return nil, err
	}
	var ret entity.AiCharacterList
	for _, property := range rsp.Property {
		info := entity.AiCharacterInfo{Type: property.Type}
		for _, v := range property.Value {
			info.Characters = append(info.Characters, entity.AiCharacter{
				Name:     v.CharacterName,
				VoiceId:  v.CharacterId,
				VoiceURL: v.CharacterVoiceUrl,
			})
		}
		ret.List = append(ret.List, info)
	}
	if len(ret.List) == 0 {
		return nil, errors.New("ai character list nil")
	}
	return &ret, nil
}

func BuildGroupAiRecordService(groupUin uint64, voiceId, text string, chatType entity.ChatType, chatId uint32) (*sso_type.SsoPacket, error) {
	return BuildOidbPacket(0x929B, 0, &oidb.OidbSvcTrpcTcp0X929B_0_Req{
		GroupUin:      groupUin,
		VoiceId:       voiceId,
		Text:          text,
		ChatType:      uint32(chatType),
		ClientMsgInfo: &oidb.OidbSvcTrpcTcp0X929B_0_Req_ClientMsgInfo{MsgRandom: chatId},
	}, false, false)
}

func ParseGroupAiRecordService(data []byte) (*message.VoiceElement, error) {
	rsp, err := ParseOidbPacket[oidb.OidbSvcTrpcTcp0X929B_0_Rsp](data)
	if err != nil {
		return nil, err
	}
	if rsp.MsgInfo == nil || len(rsp.MsgInfo.MsgInfoBody) == 0 {
		return nil, errors.New("rsp msg info nil")
	}
	index := rsp.MsgInfo.MsgInfoBody[0].Index
	elem := &message.VoiceElement{
		UUid:     index.FileUuid.Unwrap(),
		Name:     index.Info.FileName.Unwrap(),
		Size:     index.Info.FileSize.Unwrap(),
		Duration: index.Info.Time.Unwrap(),
		MsgInfo:  rsp.MsgInfo,
		// Compat ??
	}
	elem.Md5, _ = hex.DecodeString(index.Info.FileHash.Unwrap())
	elem.Sha1, _ = hex.DecodeString(index.Info.FileSha1.Unwrap())
	return elem, nil
}
