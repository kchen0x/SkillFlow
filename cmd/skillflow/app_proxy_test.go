package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestProxyConnectionDefaultsTargetURL(t *testing.T) {
	var (
		mu      sync.Mutex
		gotHost string
		gotVerb string
	)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotHost = r.Host
		gotVerb = r.Method
		mu.Unlock()
		http.Error(w, "proxy blocked", http.StatusBadGateway)
	}))
	defer proxy.Close()

	app := newProxyTestApp(t, config.ProxyConfig{Mode: config.ProxyModeNone})
	result, err := callTestProxyConnection(t, app, "", config.ProxyConfig{
		Mode: config.ProxyModeManual,
		URL:  proxy.URL,
	})
	require.NoError(t, err)

	assert.Equal(t, "https://github.com", proxyTestStringField(t, result, "TargetURL"))
	assert.False(t, proxyTestBoolField(t, result, "Success"))

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, http.MethodConnect, gotVerb)
	assert.Equal(t, "github.com:443", gotHost)
}

func TestTestProxyConnectionUsesProvidedProxyConfig(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()

	app := newProxyTestApp(t, config.ProxyConfig{
		Mode: config.ProxyModeManual,
		URL:  "http://127.0.0.1:1",
	})
	result, err := callTestProxyConnection(t, app, target.URL, config.ProxyConfig{Mode: config.ProxyModeNone})
	require.NoError(t, err)

	assert.True(t, proxyTestBoolField(t, result, "Success"))
	assert.Equal(t, int64(http.StatusNoContent), proxyTestIntField(t, result, "StatusCode"))
}

func TestTestProxyConnectionTreatsHTTPResponseAsSuccess(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer target.Close()

	app := newProxyTestApp(t, config.ProxyConfig{Mode: config.ProxyModeNone})
	result, err := callTestProxyConnection(t, app, target.URL, config.ProxyConfig{Mode: config.ProxyModeNone})
	require.NoError(t, err)

	assert.True(t, proxyTestBoolField(t, result, "Success"))
	assert.Equal(t, int64(http.StatusForbidden), proxyTestIntField(t, result, "StatusCode"))
}

func TestTestProxyConnectionTimesOut(t *testing.T) {
	prevTimeout := proxyConnectionTestTimeout
	proxyConnectionTestTimeout = 100 * time.Millisecond
	t.Cleanup(func() {
		proxyConnectionTestTimeout = prevTimeout
	})

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(150 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()

	app := newProxyTestApp(t, config.ProxyConfig{Mode: config.ProxyModeNone})
	startedAt := time.Now()
	result, err := callTestProxyConnection(t, app, target.URL, config.ProxyConfig{Mode: config.ProxyModeNone})
	require.NoError(t, err)

	assert.False(t, proxyTestBoolField(t, result, "Success"))
	assert.Contains(t, strings.ToLower(proxyTestStringField(t, result, "Message")), "deadline")
	assert.GreaterOrEqual(t, proxyTestIntField(t, result, "ElapsedMs"), int64(80))
	assert.Less(t, time.Since(startedAt), time.Second)
}

func newProxyTestApp(t *testing.T, savedProxy config.ProxyConfig) *App {
	t.Helper()

	dataDir := t.TempDir()
	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.Proxy = savedProxy
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc
	return app
}

func callTestProxyConnection(t *testing.T, app *App, targetURL string, proxy config.ProxyConfig) (reflect.Value, error) {
	t.Helper()

	method := reflect.ValueOf(app).MethodByName("TestProxyConnection")
	require.True(t, method.IsValid(), "TestProxyConnection method is missing")

	results := method.Call([]reflect.Value{reflect.ValueOf(targetURL), reflect.ValueOf(proxy)})
	require.Len(t, results, 2)

	var err error
	if !results[1].IsNil() {
		err, _ = results[1].Interface().(error)
	}
	require.False(t, results[0].IsNil())
	return results[0], err
}

func proxyTestStruct(t *testing.T, value reflect.Value) reflect.Value {
	t.Helper()
	require.True(t, value.IsValid())
	require.Equal(t, reflect.Pointer, value.Kind())
	require.False(t, value.IsNil())
	return value.Elem()
}

func proxyTestStringField(t *testing.T, value reflect.Value, field string) string {
	t.Helper()
	v := proxyTestStruct(t, value).FieldByName(field)
	require.True(t, v.IsValid(), "missing field %s", field)
	return v.String()
}

func proxyTestBoolField(t *testing.T, value reflect.Value, field string) bool {
	t.Helper()
	v := proxyTestStruct(t, value).FieldByName(field)
	require.True(t, v.IsValid(), "missing field %s", field)
	return v.Bool()
}

func proxyTestIntField(t *testing.T, value reflect.Value, field string) int64 {
	t.Helper()
	v := proxyTestStruct(t, value).FieldByName(field)
	require.True(t, v.IsValid(), "missing field %s", field)
	return v.Int()
}
