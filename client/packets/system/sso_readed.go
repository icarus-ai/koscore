package system

import (
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/client/packets/system/system_type"
	"github.com/kernel-ai/koscore/utils/proto"
)

func BuildPrivateSsoReadedReportPacket(uid string, timestamp int64, seq uint64) *sso_type.SsoPacket {
	return buildSsoReadedReportPacket(&operation.SsoReadedReport{
		C2C: &operation.SsoReadedReportC2C{TargetUid: proto.Some(uid), Time: timestamp, StartSeq: seq},
	})
}

func BuildGroupSsoReadedReportPacket(gin, seq uint64) *sso_type.SsoPacket {
	return buildSsoReadedReportPacket(&operation.SsoReadedReport{
		Grp: &operation.SsoReadedReportGroup{GroupUin: gin, StartSeq: seq},
	})
}

func buildSsoReadedReportPacket(m *operation.SsoReadedReport) *sso_type.SsoPacket {
	data, _ := proto.Marshal(m)
	return system_type.AttributeSsoReadedReport.NewSsoPacket(0, data)
}
