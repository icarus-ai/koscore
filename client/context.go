package client

import (
	"sync/atomic"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/internal/cache"
	"github.com/kernel-ai/koscore/client/internal/highway"
	"github.com/kernel-ai/koscore/client/sign"
)

type QQClient struct {
	version     *auth.AppInfo
	device      *auth.DeviceInfo
	session     *auth.Session
	sig_context sign.Provider
	sso_context *PacketContext

	is_heart_beat bool
	is_online     atomic.Bool
	stat          Statistics
	cache         cache.Cache
	hw_session    highway.Session

	*Events
	logger
}

func (m *QQClient) SetLogger(log logger) {
	if log != nil {
		m.logger = log
	}
}
func (m *QQClient) GetSignProvider() sign.Provider     { return m.sig_context }
func (m *QQClient) SetSignProvider(prov sign.Provider) { m.sig_context = prov }
func (m *QQClient) GetDevice() *auth.DeviceInfo        { return m.device }
func (m *QQClient) SetDevice(info *auth.DeviceInfo)    { m.device = info }
func (m *QQClient) GetVersion() *auth.AppInfo          { return m.version }
func (m *QQClient) SetVersion(info *auth.AppInfo) {
	m.version = info
	m.hw_session.AppId = info.AppId
	m.hw_session.SubAppId = info.SubAppId
}

func (m *QQClient) SaveToken(path string) error { return m.session.Save(path) }
func (m *QQClient) LoadToken(path string) error {
	session, e := auth.LoadSession(path)
	if e != nil {
		return e
	}
	m.session = session
	return nil
}

func (m *QQClient) Uin() uint64  { return m.session.Info.Uin }
func (m *QQClient) Uid() string  { return m.session.Info.Uid }
func (m *QQClient) Nick() string { return m.session.Info.Name }

// 设置qq已经上线
func (m *QQClient) setOnline() { m.is_online.Store(true) }

func NewClient(uin uint64, password string) *QQClient {
	ctx := &QQClient{
		session:       auth.NewSession(),
		is_heart_beat: false,
		logger:        log_t{},
		Events:        newEventCall(),
	}
	ctx.session.Info.Uin = uin
	ctx.sso_context = NewPacketContext(ctx)
	ctx.hw_session.Uin = &ctx.session.Info.Uin
	return ctx
}

func (m *QQClient) Release() {
	m.sig_context.Release()
	if m.is_online.Load() {
		m.sso_context.Disconnect()
	}
}

// *****

func (m *QQClient) Session() *auth.Session     { return m.session }
func (m *QQClient) SsoPaacket() *PacketContext { return m.sso_context }

func (m *QQClient) GetCache() *cache.Cache                       { return &m.cache }
func (m *QQClient) GetCacheUid(uin uint64, gin ...uint64) string { return m.cache.GetUid(uin, gin...) }
func (m *QQClient) GetCacheUin(uid string, gin ...uint64) uint64 { return m.cache.GetUin(uid, gin...) }
