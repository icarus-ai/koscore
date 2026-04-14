package sign

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/common"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/types"
	//"github.com/kernel-ai/koscore/utils/comm"
)

type (
	ClientV2 struct {
		lock         sync.RWMutex
		signCount    atomic.Uint32
		instances    []*remoteV2
		app          *auth.AppInfo
		lastTestTime time.Time

		device *auth.DeviceInfo
		uin    uint32
	}

	remoteV2 struct {
		server  string
		headers types.MapSS
		latency atomic.Uint32
	}
)

func NewSignerV2(uin uint32, app *auth.AppInfo, device *auth.DeviceInfo, sign_server_token []string) *ClientV2 {
	var servs []*remoteV2
	for i := 0; i < len(sign_server_token); i += 2 {
		servs = append(servs, &remoteV2{server: sign_server_token[i], headers: types.MapSS{
			"Authorization": "Bearer " + sign_server_token[i+1],
			"User-Agent":    "kosbot qq/" + app.CurrentVersion,
		}})
	}
	client := &ClientV2{instances: servs, uin: uin, app: app, device: device}
	go client.test()
	return client
}

func (c *ClientV2) Reset() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, i := range c.instances {
		i.latency.Store(0)
	}
}

func (c *ClientV2) Release()                           {}
func (c *ClientV2) AddRequestHeader(heads types.MapSS) {}
func (c *ClientV2) AddSignServer(servers ...string)    {}

func (c *ClientV2) GetSignServer() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return utils.Map(c.instances, func(sign *remoteV2) string { return sign.server })
}

func (c *ClientV2) SetAppInfo(app *auth.AppInfo) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.app = app
}

func (c *ClientV2) getAvailableSign() *remoteV2 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, i := range c.instances {
		if i.latency.Load() < serverLatencyDown {
			return i
		}
	}
	return nil
}

func (c *ClientV2) sortByLatency() {
	c.lock.Lock()
	defer c.lock.Unlock()
	sort.SliceStable(c.instances, func(i, j int) bool {
		return c.instances[i].latency.Load() < c.instances[j].latency.Load()
	})
}

func (c *ClientV2) Sign(cmd string, seq uint32, data []byte) (*common.SsoSecureInfo, error) {
	if !ContainSignPKG(cmd) {
		return nil, nil
	}
	if time.Now().After(c.lastTestTime.Add(30 * time.Minute)) {
		go c.test()
	}
	//ts := time.Now().UnixMilli()
	if sign := c.getAvailableSign(); sign == nil {
		c.Reset()
	}
	if sign := c.getAvailableSign(); sign != nil {
		rsp, e := sign.sign(cmd, seq, data, c.uin, c.device.GUID.ToLowHexStr(), c.app.QUA)
		if e == nil {
			if !bytes.Contains(rsp.Value.Extra, []byte(c.app.QUA)) {
				return nil, ErrVersionMismatch
			}
			//comm.LOGD("signed for [%s:%d] %X", cmd, seq, rsp.Value.Sign)
			c.signCount.Add(1)
			if rsp.Value.Token == nil {
				rsp.Value.Token = binary.Empty
			}
			return &common.SsoSecureInfo{
				SecSign:  rsp.Value.Sign,
				SecToken: rsp.Value.Token,
				SecExtra: rsp.Value.Extra,
			}, nil
		}
		sign.latency.Store(serverLatencyDown)
	}
	go c.test() // 全寄了, 重新再测下
	return nil, ErrAllSignServiceDown
}

func (c *ClientV2) test() {
	c.lock.Lock()
	if time.Now().Before(c.lastTestTime.Add(10 * time.Minute)) {
		c.lock.Unlock()
		return
	}
	c.lastTestTime = time.Now()
	c.lock.Unlock()
	for _, i := range c.instances {
		i.test(c.uin, c.device.GUID.ToLowHexStr(), c.app.QUA)
	}
	c.sortByLatency()
}

func (i *remoteV2) sign(cmd string, seq uint32, buf []byte, uin uint32, guid, qua string) (*ResponseV2, error) {
	s := fmt.Sprintf(`{"command":"%s","seq":"%d","body":"%x","uin":"%d","guid":"%s","qua":"%s"}`, cmd, seq, buf, uin, guid, qua)
	rsp, e := httpPost[ResponseV2](i.server, utils.S2B(s), i.headers)
	if e != nil {
		return nil, e
	}
	if len(rsp.Value.Sign) == 0 {
		return nil, k_err_sign_rsp
	}
	return &rsp, nil
}

func (i *remoteV2) test(uin uint32, guid string, qua string) {
	ts := time.Now().UnixMilli()
	rsp, e := i.sign("wtlogin.login", 1, []byte{11, 45, 14}, uin, guid, qua)
	if e != nil || len(rsp.Value.Sign) == 0 {
		//comm.LOGW("测试签名服务器: %s时出现错误: %v", i.server, e)
		i.latency.Store(serverLatencyDown)
		//???return
	}
	// 有长连接的情况，取两次平均值
	rsp, e = i.sign("wtlogin.login", 1, []byte{11, 45, 14}, uin, guid, qua)
	if e != nil || len(rsp.Value.Sign) == 0 {
		//comm.LOGW("测试签名服务器: %s时出现错误: %v", i.server, e)
		i.latency.Store(serverLatencyDown)
		return
	}
	// 粗略计算，应该足够了
	i.latency.Store(uint32(time.Now().UnixMilli()-ts) / 2)
	//comm.LOGW("签名服务器: %s 延迟: %dms", i.server, latency)
}
