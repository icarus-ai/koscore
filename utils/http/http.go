package http

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kernel-ai/koscore/utils/types"
)

var __http = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext, // newHosts().DialContext,
		ForceAttemptHTTP2:   true,
		DisableCompression:  true,
		MaxIdleConns:        0,  // 所有host的连接池最大连接数量
		MaxIdleConnsPerHost: 10, // 每个host的连接池最大空闲连接数,默认2
		MaxConnsPerHost:     0,  // 对每个host的最大连接数量，0表示不限制
		IdleConnTimeout:     90 * time.Second,
		//ResponseHeaderTimeout: 15 * time.Second, // 限制读取response header的时间,默认 timeout + 5*time.Second
		ExpectContinueTimeout: 2 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true, // 跳过证书验证
			// PreferServerCipherSuites: true,
			//CipherSuites     : k_cipher_suites,
		}}}

func http_request(uri, mode string, body io.Reader, heads types.MapSS) (*http.Request, error) {
	req, e := http.NewRequest(mode, uri, body)
	if e != nil {
		return nil, e
	}
	for k, v := range heads {
		req.Header.Add(k, v)
	}
	return req, nil
}

func Get(uri string, heads types.MapSS) ([]byte, error) {
	req, e := http_request(uri, http.MethodGet, nil, heads)
	if e != nil {
		return nil, e
	}
	return doHTTP(req)
}

func Post(uri string, body []byte, heads types.MapSS) ([]byte, error) {
	req, e := http_request(uri, http.MethodPost, bytes.NewReader(body), heads)
	if e != nil {
		return nil, e
	}
	return doHTTP(req)
}

//nolint:bodyclose
func doHTTP(req *http.Request) ([]byte, error) {
	rsp, e := __http.Do(req)
	if e != nil {
		return nil, e
	}
	defer rsp.Body.Close()
	if rsp.StatusCode == http.StatusOK {
		if rsp.Body == nil {
			return nil, errors.New("http_error_body")
		}
		data, e := io.ReadAll(rsp.Body)
		if e != nil {
			return nil, e
		}
		if strings.Contains(rsp.Header.Get("Content-Encoding"), "gzip") {
			r, e := gzip.NewReader(bytes.NewReader(data))
			if e != nil {
				return nil, e
			}
			defer r.Close()
			return io.ReadAll(r)
		}
		return data, e
	}
	return nil, fmt.Errorf("fail: http.status.code: %d", rsp.StatusCode)
}
