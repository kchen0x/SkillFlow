package shellsettings

import "strings"

// ProxyMode controls how outbound HTTP requests are routed.
// "none" = direct, "system" = read HTTP_PROXY/HTTPS_PROXY env vars, "manual" = use URL field.
type ProxyMode string

const (
	ProxyModeNone   ProxyMode = "none"
	ProxyModeSystem ProxyMode = "system"
	ProxyModeManual ProxyMode = "manual"
)

type ProxyConfig struct {
	Mode ProxyMode `json:"mode"` // "none" | "system" | "manual"
	URL  string    `json:"url"`  // used when Mode == "manual", e.g. "http://127.0.0.1:7890"
}

func NormalizeProxyConfig(proxy ProxyConfig) ProxyConfig {
	mode := ProxyMode(strings.ToLower(strings.TrimSpace(string(proxy.Mode))))
	switch mode {
	case ProxyModeSystem, ProxyModeManual:
		proxy.Mode = mode
	default:
		proxy.Mode = ProxyModeNone
	}
	proxy.URL = strings.TrimSpace(proxy.URL)
	return proxy
}

func IsZeroProxyConfig(proxy ProxyConfig) bool {
	mode := ProxyMode(strings.ToLower(strings.TrimSpace(string(proxy.Mode))))
	url := strings.TrimSpace(proxy.URL)
	return url == "" && (mode == "" || mode == ProxyModeNone)
}
