package sign

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

var __http = &http.Client{
	Timeout: 18 * time.Second,
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   8 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		//ForceAttemptHTTP2    : true,
		//DisableCompression   : true,
		MaxIdleConns:        0,  // 所有host的连接池最大连接数量
		MaxIdleConnsPerHost: 10, // 每个host的连接池最大空闲连接数,默认2
		MaxConnsPerHost:     0,  // 对每个host的最大连接数量，0表示不限制
		IdleConnTimeout:     60 * time.Second,
		//ResponseHeaderTimeout: 15 * time.Second, // 限制读取response header的时间,默认 timeout + 5*time.Second
		ExpectContinueTimeout: 2 * time.Second,
		TLSHandshakeTimeout:   8 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true, // 跳过证书验证
			//PreferServerCipherSuites: true,
		}}}

func http_request(uri, mode string, body io.Reader, heads header) (*http.Request, error) {
	req, e := http.NewRequest(mode, uri, body)
	if e != nil {
		return nil, e
	}
	for k, v := range heads {
		req.Header.Add(k, v)
	}
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func httpGet[T any](uri string, heads header) (target T, err error) {
	req, e := http_request(uri, http.MethodGet, nil, heads)
	if e != nil {
		err = e
		return
	}
	return doHTTP[T](req)
}

func httpPost[T any](uri string, body []byte, heads header) (target T, err error) {
	req, e := http_request(uri, http.MethodPost, bytes.NewReader(body), heads)
	if e != nil {
		err = e
		return
	}
	return doHTTP[T](req)
}

//nolint:bodyclose
func doHTTP[T any](req *http.Request) (target T, e error) {
	var rsp *http.Response
	if rsp, e = __http.Do(req); e == nil {
		defer rsp.Body.Close()
		if rsp.StatusCode == http.StatusOK {
			e = json.NewDecoder(rsp.Body).Decode(&target)
		} else {
			e = fmt.Errorf("fail: http.status.code: %d", rsp.StatusCode)
		}
	}
	return
}
