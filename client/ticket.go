package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/RomiChan/syncx"
)

type (
	keyInfo struct {
		key        string
		expireTime time.Time
	}

	TicketService struct {
		client *http.Client
		psKeys syncx.Map[string, *keyInfo]
		sKey   *keyInfo
	}

	Cookies struct {
		uin   uint64
		SKey  string
		PsKey string
	}
)

func (m *QQClient) SendRequestWithCookie(request *http.Request) (*http.Response, error) {
	u, e := url.Parse(request.URL.String())
	if e != nil {
		return nil, e
	}
	cookies, e := m.GetCookies(u.Hostname())
	if e != nil {
		return nil, e
	}
	request.AddCookie(&http.Cookie{Name: "skey", Value: cookies.SKey})
	request.AddCookie(&http.Cookie{Name: "p_uin", Value: strconv.Itoa(int(cookies.uin))})
	request.AddCookie(&http.Cookie{Name: "p_skey", Value: cookies.PsKey})
	return m.ticket.client.Do(request)
}

func (m *QQClient) GetSkey() (string, error) {
	if time.Now().Before(m.ticket.sKey.expireTime) {
		return m.ticket.sKey.key, nil
	}
	key, e := m.FetchClientKey() // clientKey
	if e != nil {
		return "", e
	}
	jump := "https%3A%2F%2Fh5.qzone.qq.com%2Fqqnt%2Fqzoneinpcqq%2Ffriend%3Frefresh%3D0%26clientuin%3D0%26darkMode%3D0&keyindex=19&random=2599"
	u, _ := url.Parse(fmt.Sprintf("https://ssl.ptlogin2.qq.com/jump?ptlang=1033&clientuin=%d&clientkey=%s&u1=%s", m.Uin(), key.ClientKey, jump))
	rsp, e := m.ticket.client.Get(u.String())
	if e != nil {
		return "", e
	}
	rsp.Body.Close()
	for _, cookie := range m.ticket.client.Jar.Cookies(u) {
		if cookie.Name == "skey" {
			m.ticket.sKey.key = cookie.Value
			m.ticket.sKey.expireTime = time.Now().Add(24 * time.Hour)
			break
		}
	}
	return m.ticket.sKey.key, nil
}

func (m *QQClient) GetCsrfToken() (int, error) {
	skey, e := m.GetSkey()
	if e != nil {
		return -1, e
	}
	hash := 5381
	for _, ch := range skey {
		hash += (hash << 5) + int(ch)
	}
	return hash & 2147483647, nil
}

func (m *QQClient) GetCookies(domain string) (*Cookies, error) {
	skey, e := m.GetSkey()
	if e != nil {
		return nil, e
	}
	var token string
	if tokenTime, ok := m.ticket.psKeys.Load(domain); ok {
		if time.Now().Before(tokenTime.expireTime) {
			token = tokenTime.key
		} else {
			cookies, e := m.FetchCookies([]string{domain})
			if e != nil {
				return nil, e
			}
			token = cookies[domain]
			m.ticket.psKeys.Store(domain, &keyInfo{
				key:        token,
				expireTime: time.Now().Add(24 * time.Hour),
			})
		}
	} else {
		cookies, e := m.FetchCookies([]string{domain})
		if e != nil {
			return nil, e
		}
		token = cookies[domain]
		m.ticket.psKeys.Store(domain, &keyInfo{
			key:        token,
			expireTime: time.Now().Add(24 * time.Hour),
		})
	}
	return &Cookies{
		uin:   m.Uin(),
		SKey:  skey,
		PsKey: token,
	}, nil
}

func GTK(s string) int {
	hash := 5381
	for _, ch := range s {
		hash += (hash<<5)&2147483647 + int(ch)&2147483647
		hash &= 2147483647
	}
	return hash
}
