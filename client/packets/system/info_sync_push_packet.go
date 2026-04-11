package system

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/system"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/client/packets/system/system_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func ParseInfoSyncPushPacket(pkt *sso_type.SsoPacket) *system_type.InfoSyncRsp {
	rsp, e := proto.Unmarshal[system.InfoSyncPush](pkt.Data)
	ret := &system_type.InfoSyncRsp{}
	if e != nil {
		ret.Message = e.Error()
	} else if rsp.ErrMsg.IsSome() {
		ret.Message = rsp.ErrMsg.Unwrap()
	}
	return ret
}
