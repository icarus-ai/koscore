package internal

import (
	"net"
	"net/netip"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/kernel-ai/koscore/client/internal/network"
	"github.com/kernel-ai/koscore/utils"
)

type SocketContext struct {
	//HeaderSize uint32
	//Connected  bool
	//_client    ClientListener
	//context BotContext
	servers         []netip.AddrPort
	customs         []netip.AddrPort
	sock            network.TCPClient
	alive           bool
	currServerIndex int
	retryTimes      int

	getOptimumServer bool
	useIPv6Network   bool

	onRecvPacket     func(data []byte)
	onQuickReconnect func()
}

func (m *SocketContext) SetCustomServer(v []netip.AddrPort) { m.customs = v }

func (m *SocketContext) GetCurrServer() netip.AddrPort { return m.servers[m.currServerIndex] }

func (m *SocketContext) resolve_dns() {
	m.servers = nil
	m.servers = append(m.servers, m.customs...)
	hsot := utils.Ternary(m.useIPv6Network, "msfwifiv6.3g.qq.com", "msfwifi.3g.qq.com")
	adds, e := net.LookupIP(hsot) // host servers
	if e == nil && len(adds) > 0 {
		for _, addr := range adds {
			if ip, ok := netip.AddrFromSlice(addr.To4()); ok {
				m.servers = append(m.servers, netip.AddrPortFrom(ip, 8080))
			}
			if ip, ok := netip.AddrFromSlice(addr.To16()); ok {
				m.servers = append(m.servers, netip.AddrPortFrom(ip, 14000))
			}
		}
	}
	if len(m.servers) == 0 {
		m.servers = []netip.AddrPort{ // default servers
			netip.AddrPortFrom(netip.AddrFrom4([4]byte{43, 135, 106, 161}), 8080),
			netip.AddrPortFrom(netip.AddrFrom4([4]byte{43, 154, 240, 13}), 8080),
		}
	}
}

func quality_test(addr string) (int64, error) {
	// see QualityTestManager
	start := time.Now()
	conn, e := net.DialTimeout("tcp", addr, time.Second*5)
	if e != nil {
		return 0, errors.Wrap(e, "failed to connect to server during quality test")
	}
	_ = conn.Close()
	return time.Since(start).Milliseconds(), nil
}

func (m *SocketContext) sort_servers() {
	pings := make([]int64, len(m.servers))
	wg := sync.WaitGroup{}
	wg.Add(len(m.servers))
	for i := range m.servers {
		go func(idx int) {
			defer wg.Done()
			p, e := quality_test(m.servers[idx].String())
			if e == nil {
				pings[idx] = p
			} else {
				pings[idx] = 9999
			}
		}(i)
	}
	wg.Wait()
	sort.Slice(m.servers, func(i, k int) bool { return pings[i] < pings[k] })
	if len(m.servers) > 3 {
		m.servers = m.servers[0 : len(m.servers)/2]
	} // 保留ping值中位数以上的server
}

func NewSocketContext(
	onRecvPacket func(data []byte),
	onQuickReconnect func(),
	plannedDisconnect func(*network.TCPClient),
	unexpectedDisconnect func(*network.TCPClient, error),
) *SocketContext {
	ctx := &SocketContext{
		alive:            false,
		servers:          nil,
		currServerIndex:  0,
		getOptimumServer: false,
		useIPv6Network:   false,
		onRecvPacket:     onRecvPacket,
		onQuickReconnect: onQuickReconnect,
	}
	ctx.sock.PlannedDisconnect(plannedDisconnect)
	ctx.sock.UnexpectedDisconnect(unexpectedDisconnect)
	return ctx
}

func (m *SocketContext) Send(packet []byte) error { return m.sock.Write(packet) }

func (m *SocketContext) IsConnect() bool { return m.sock.Connected }
func (m *SocketContext) Connect() error {
	// init qq servers
	m.resolve_dns()
	if m.getOptimumServer {
		m.sort_servers()
	}

	addr := m.servers[m.currServerIndex].String()
	//c.info("connect to server: %v", addr)
	err := m.sock.Connect(addr)
	m.currServerIndex++
	if m.currServerIndex == len(m.servers) {
		m.currServerIndex = 0
	}
	if err != nil {
		m.retryTimes++
		if m.retryTimes > len(m.servers) {
			return errors.New("all servers are unreachable")
		}
		return err
	}
	m.retryTimes = 0

	if !m.alive {
		m.alive = true
		go m.netLoop()
	}

	return nil
}

// 通过循环来不停接收数据包
func (m *SocketContext) netLoop() {
	errCount := 0
	//comm.LOGD("netLoop: m.alive %v", m.alive)
	for m.alive {
		l, err := m.sock.ReadInt32()
		//comm.LOGD("netLoop: m.TCP.ReadInt32() %v %v", err, l)
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			continue
		}
		if l < 4 || l > 1024*1024*10 { // max 10MB
			//comm.LOGD("parse incoming packet error: invalid packet length %v", l)
			errCount++
			if errCount > 2 {
				go m.onQuickReconnect()
				errCount = 0
			}
			continue
		}
		packet, _ := m.sock.ReadBytes_AddHeadSizeU32(int(l) - 4)
		go m.onRecvPacket(packet)
	}
}

func (m *SocketContext) Disconnect() { m.sock.Close() }
