// pkg/utils/http.go
package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// Client 自带连接池的 HTTP 客户端
var Client = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	},
	Timeout: 30 * time.Second,
}

// SendOption 发送选项
type SendOption func(*requestCfg)

type requestCfg struct {
	retries int
	handler []func(*http.Request)
}

// WithRetry 设置重试次数
func WithRetry(n int) SendOption {
	return func(c *requestCfg) {
		c.retries = n
	}
}

// WithHeader 添加统一头
func WithHeader(k, v string) SendOption {
	return func(c *requestCfg) {
		c.handler = append(c.handler, func(r *http.Request) {
			r.Header.Set(k, v)
		})
	}
}

// Request 简化 GET
func Request(ctx context.Context, method, url string, body io.Reader, opts ...SendOption) (*http.Response, error) {
	cfg := &requestCfg{retries: 1}
	for _, o := range opts {
		o(cfg)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, h := range cfg.handler {
		h(req)
	}

	var resp *http.Response
	var lastErr error
	for i := 0; i < cfg.retries; i++ {
		resp, lastErr = Client.Do(req)
		if lastErr == nil {
			break
		}
		if i < cfg.retries-1 {
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}
	return resp, lastErr
}

// GetJSON 获取并解包 JSON
func GetJSON(ctx context.Context, url string, dst interface{}, opts ...SendOption) error {
	resp, err := Request(ctx, http.MethodGet, url, http.NoBody, opts...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// PostJSON 发送并取回 JSON
func PostJSON(ctx context.Context, url string, in, out interface{}, opts ...SendOption) error {
	body, _ := json.Marshal(in)
	resp, err := Request(ctx, http.MethodPost, url, bytes.NewReader(body), opts...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// -------------------------------------------------------------------------
// 中间件包装
// -------------------------------------------------------------------------

// MiddlewareFunc 简单中间件签名
type MiddlewareFunc func(http.Handler) http.Handler

// Chain 将多个中间件串联
func Chain(h http.Handler, m ...MiddlewareFunc) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

// CorsMiddleware 设置跨域
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
//Personal.AI order the ending
