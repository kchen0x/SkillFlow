package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeGitHubJSONResponseReturnsStatusError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"message":"API rate limit exceeded"}`)),
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	err := decodeGitHubJSONResponse(resp, &release)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "github status 403")
	assert.Contains(t, err.Error(), "API rate limit exceeded")
}

func TestDecodeGitHubJSONResponseDecodesSuccessPayload(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"tag_name":"v1.2.3"}`)),
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	require.NoError(t, decodeGitHubJSONResponse(resp, &release))
	assert.Equal(t, "v1.2.3", release.TagName)
}

func TestProxyHTTPClientSetsResponseHeaderTimeoutForManualProxy(t *testing.T) {
	dataDir := t.TempDir()
	svc := config.NewService(dataDir)

	cfg := config.DefaultConfig(dataDir)
	cfg.Proxy = config.ProxyConfig{
		Mode: config.ProxyModeManual,
		URL:  "http://127.0.0.1:7890",
	}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc

	client := app.proxyHTTPClient()
	require.NotNil(t, client)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Greater(t, transport.ResponseHeaderTimeout, time.Duration(0))
}
