package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shinerio/skillflow/core/config"
)

const defaultProxyTestTargetURL = "https://github.com"

var proxyConnectionTestTimeout = 5 * time.Second

type ProxyConnectionTestResult struct {
	TargetURL  string `json:"targetURL"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"statusCode"`
	ElapsedMs  int64  `json:"elapsedMs"`
	Message    string `json:"message"`
}

func proxyHTTPClientWithConfig(proxyCfg config.ProxyConfig) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = proxyResponseHeaderTimeout

	switch proxyCfg = config.NormalizeProxyConfig(proxyCfg); proxyCfg.Mode {
	case config.ProxyModeSystem:
		transport.Proxy = http.ProxyFromEnvironment
	case config.ProxyModeManual:
		if proxyCfg.URL == "" {
			transport.Proxy = nil
			break
		}
		proxyURL, err := url.Parse(proxyCfg.URL)
		if err != nil {
			transport.Proxy = nil
			break
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	default:
		transport.Proxy = nil
	}
	return &http.Client{Transport: transport}
}

func normalizeProxyTestTargetURL(raw string) (string, error) {
	targetURL := strings.TrimSpace(raw)
	if targetURL == "" {
		targetURL = defaultProxyTestTargetURL
	}

	parsed, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid target url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("target url must use http or https")
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return "", fmt.Errorf("target url host is required")
	}
	return parsed.String(), nil
}

func (a *App) TestProxyConnection(targetURL string, proxy config.ProxyConfig) (*ProxyConnectionTestResult, error) {
	normalizedTargetURL, err := normalizeProxyTestTargetURL(targetURL)
	if err != nil {
		a.logErrorf("proxy connectivity test failed: target=%s err=%v", strings.TrimSpace(targetURL), err)
		return nil, err
	}

	baseCtx := a.ctx
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(baseCtx, proxyConnectionTestTimeout)
	defer cancel()

	proxy = config.NormalizeProxyConfig(proxy)
	a.logInfof("proxy connectivity test started: target=%s mode=%s", normalizedTargetURL, proxy.Mode)
	startedAt := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalizedTargetURL, nil)
	if err != nil {
		a.logErrorf("proxy connectivity test failed: target=%s mode=%s err=%v", normalizedTargetURL, proxy.Mode, err)
		return nil, err
	}

	resp, err := proxyHTTPClientWithConfig(proxy).Do(req)
	elapsedMs := time.Since(startedAt).Milliseconds()
	if err != nil {
		a.logErrorf("proxy connectivity test failed: target=%s mode=%s elapsedMs=%d err=%v", normalizedTargetURL, proxy.Mode, elapsedMs, err)
		return &ProxyConnectionTestResult{
			TargetURL: normalizedTargetURL,
			Success:   false,
			ElapsedMs: elapsedMs,
			Message:   err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	message := strings.TrimSpace(resp.Status)
	if message == "" {
		message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	a.logInfof("proxy connectivity test completed: target=%s mode=%s status=%d elapsedMs=%d", normalizedTargetURL, proxy.Mode, resp.StatusCode, elapsedMs)
	return &ProxyConnectionTestResult{
		TargetURL:  normalizedTargetURL,
		Success:    true,
		StatusCode: resp.StatusCode,
		ElapsedMs:  elapsedMs,
		Message:    message,
	}, nil
}
