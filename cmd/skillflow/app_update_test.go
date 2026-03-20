package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func TestBuildWindowsUpdateScriptTerminatesHelperAndRetriesReplace(t *testing.T) {
	script := buildWindowsUpdateScript(`C:\Temp\skillflow_update.exe`, `C:\Program Files\SkillFlow\SkillFlow.exe`, 4321)

	assert.Contains(t, script, `taskkill /PID 4321 /T /F > nul 2>&1`)
	assert.Contains(t, script, `set /a attempts=0`)
	assert.Contains(t, script, `:retry`)
	assert.Contains(t, script, `move /y "C:\Temp\skillflow_update.exe" "C:\Program Files\SkillFlow\SkillFlow.exe" > nul 2>&1`)
	assert.Contains(t, script, `if %attempts% GEQ 15 exit /b 1`)
	assert.Contains(t, script, `start "" "C:\Program Files\SkillFlow\SkillFlow.exe"`)
}

func TestBuildWindowsUpdateScriptSkipsHelperTerminationWhenUnavailable(t *testing.T) {
	script := buildWindowsUpdateScript(`C:\Temp\skillflow_update.exe`, `C:\Program Files\SkillFlow\SkillFlow.exe`, 0)

	assert.NotContains(t, script, `taskkill /PID`)
}

func TestSelectReleaseAssetDownloadURLPrefersExactWindowsExecutable(t *testing.T) {
	url := selectReleaseAssetDownloadURL("windows", "amd64", []releaseAsset{
		{Name: "SkillFlow-windows-installer.exe", BrowserDownloadURL: "https://example.com/installer.exe"},
		{Name: "SkillFlow-windows.exe.sha256", BrowserDownloadURL: "https://example.com/SkillFlow-windows.exe.sha256"},
		{Name: "SkillFlow-windows.exe", BrowserDownloadURL: "https://example.com/SkillFlow-windows.exe"},
	})

	assert.Equal(t, "https://example.com/SkillFlow-windows.exe", url)
}

func TestReleaseChecksumURLAppendsSHA256Suffix(t *testing.T) {
	assert.Equal(
		t,
		"https://github.com/shinerio/SkillFlow/releases/download/v1.0.8/SkillFlow-windows.exe.sha256",
		releaseChecksumURL("https://github.com/shinerio/SkillFlow/releases/download/v1.0.8/SkillFlow-windows.exe"),
	)
}

func TestValidateDownloadedWindowsUpdateRejectsHashMismatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SkillFlow.exe")
	require.NoError(t, os.WriteFile(path, append(validPEHeader(), []byte("payload")...), 0644))

	err := validateDownloadedWindowsUpdate(path, strings.Repeat("0", 64))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sha256 mismatch")
}

func TestValidateDownloadedWindowsUpdateAcceptsMatchingHash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SkillFlow.exe")
	content := append(validPEHeader(), []byte("payload")...)
	require.NoError(t, os.WriteFile(path, content, 0644))
	sum := sha256Hex(content)

	require.NoError(t, validateDownloadedWindowsUpdate(path, sum))
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return strings.ToLower(hex.EncodeToString(sum[:]))
}

func validPEHeader() []byte {
	header := make([]byte, 0x80)
	header[0] = 'M'
	header[1] = 'Z'
	header[0x3c] = 0x40
	header[0x40] = 'P'
	header[0x41] = 'E'
	return header
}
