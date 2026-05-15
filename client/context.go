package client

import (
	"net/http"
	"net/http/cookiejar"
	"net/netip"

	"golang.org/x/net/publicsuffix"

	"github.com/kernel-ai/koscore/client/auth"
	"github.com/kernel-ai/koscore/client/event"
	"github.com/kernel-ai/koscore/client/internal/cache"
	"github.com/kernel-ai/koscore/client/internal/highway"
	"github.com/kernel-ai/koscore/client/internal/utils"
	"github.com/kernel-ai/koscore/client/packets/structs/sso_type"
	"github.com/kernel-ai/koscore/client/sign"
	"github.com/kernel-ai/koscore/utils/types"
)

type QQClient struct {
	version     *auth.AppInfo
	device      *auth.DeviceInfo
	session     *auth.Session
	sig_context sign.Provider
	sso_context *PacketContext

	stat       utils.Statistics
	cache      cache.Cache
	hw_session *highway.Session
	ticket     *TicketService

	alive  *types.Signal
	Online *types.Signal

	decoders event.DecodersEvent

	Uin uint64

	face_details cache.FaceDetailCache // 表情信息缓存

	*Events
	logger
}

func (m *QQClient) RegistryDecodersEvent(fn_name string, fn event.DecodersCall) {
	m.decoders[fn_name] = fn
}

func (m *QQClient) GetSignProvider() sign.Provider     { return m.sig_context }
func (m *QQClient) SetSignProvider(prov sign.Provider) { m.sig_context = prov }
func (m *QQClient) GetDevice() *auth.DeviceInfo        { return m.device }
func (m *QQClient) SetDevice(info *auth.DeviceInfo)    { m.device = info }
func (m *QQClient) GetVersion() *auth.AppInfo          { return m.version }
func (m *QQClient) SetVersion(info *auth.AppInfo) {
	m.version = info
	m.initHighwayServers()
}

func (m *QQClient) UseSig(info *auth.Session) {
	m.session = info
	m.setSessionId()
}

func (m *QQClient) SetLogger(v logger) {
	if v != nil {
		m.logger = v
	}
}
func (m *QQClient) SetCustomServer(v []netip.AddrPort) { m.sso_context.sock.SetCustomServer(v) }

func (m *QQClient) setSessionId() { m.Uin = m.session.Info.Uin }

func NewClient(uin uint64, password string) *QQClient {
	cookie, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	ctx := &QQClient{
		session: auth.NewSession(uin),
		logger:  log_t{},
		Events:  newEventCall(),
		ticket: &TicketService{
			client: &http.Client{Jar: cookie},
			sKey:   &keyInfo{},
		},
		decoders: make(event.DecodersEvent),

		alive:  types.NewSignal(),
		Online: types.NewSignal(),
	}
	ctx.sso_context = NewPacketContext(ctx)
	return ctx
}

func (m *QQClient) Disconnect() { m.sso_context.Disconnect() }

func (m *QQClient) Release() {
	m.session.Clear()
	m.sig_context.Release()
	if m.Online.Load() {
		m.sso_context.Disconnect()
	}
	m.Online.Reset()
	m.alive.Reset()
}

func (m *QQClient) GetStatistics() *utils.Statistics { return &m.stat }

// *****

func (m *QQClient) SendOidbPacketAndWait(pkt *sso_type.SsoPacket) (*sso_type.SsoPacket, error) {
	return m.sendOidbPacketAndWait(pkt)
}

func (m *QQClient) Session() *auth.Session                       { return m.session }
func (m *QQClient) GetCache() *cache.Cache                       { return &m.cache }
func (m *QQClient) GetCacheUid(uin uint64, gin ...uint64) string { return m.cache.GetUid(uin, gin...) }
func (m *QQClient) GetCacheUin(uid string, gin ...uint64) uint64 { return m.cache.GetUin(uid, gin...) }
