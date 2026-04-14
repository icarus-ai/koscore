package client

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/global.go

import (
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/kernel-ai/koscore/utils"
)

func qualityTest(addr string) (int64, error) {
	// see QualityTestManager
	start := time.Now()
	conn, e := net.DialTimeout("tcp", addr, time.Second*5)
	if e != nil {
		return 0, errors.Wrap(e, "failed to connect to server during quality test")
	}
	_ = conn.Close()
	return time.Since(start).Milliseconds(), nil
}

// ConnectionQualityInfo 客户端连接质量测试结果
// 延迟单位为 ms 如为 9999 则测试失败 测试方法为 TCP 连接测试
// 丢包测试方法为 ICMP. 总共发送 10 个包, 记录丢包数
type ConnectionQualityInfo struct {
	ChatServerLatency    int64 // 聊天服务器延迟
	ChatServerPacketLoss int   // 聊天服务器ICMP丢包数
	SrvServerLatency     int64 // Highway服务器延迟. 涉及媒体以及群文件上传
	SrvServerPacketLoss  int   // Highway服务器ICMP丢包数.
}

func (m *QQClient) ConnectionQualityTest() *ConnectionQualityInfo {
	if !m.Online.Load() {
		return nil
	}

	r := &ConnectionQualityInfo{}
	wg := sync.WaitGroup{}
	wg.Add(4)

	t_addr := m.sso_context.sock.GetCurrServer().String()

	go func() {
		defer wg.Done()
		if latency, e := qualityTest(t_addr); e != nil {
			//c.error("test chat server latency error: %v", e)
			r.ChatServerLatency = 9999
		} else {
			r.ChatServerLatency = latency
		}
	}()

	go func() {
		defer wg.Done()
		_ = m.ensureHighwayServers()
		if m.hw_session.AddrLength() > 0 {
			if latency, e := qualityTest(m.hw_session.SsoAddr[0].String()); e != nil {
				//c.error("test srv server latency error: %v", e)
				r.SrvServerLatency = 9999
			} else {
				r.SrvServerLatency = latency
			}
		}
	}()

	go func() {
		defer wg.Done()
		res := utils.RunTCPPingLoop(t_addr, 10)
		r.ChatServerPacketLoss = res.PacketsLoss
	}()

	go func() {
		defer wg.Done()
		if m.hw_session.AddrLength() > 0 {
			res := utils.RunTCPPingLoop(m.hw_session.SsoAddr[0].String(), 10)
			r.SrvServerPacketLoss = res.PacketsLoss
		}
	}()
	wg.Wait()
	return r
}
