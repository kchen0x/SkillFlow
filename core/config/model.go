package config

import (
	"strings"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/platform/logging"
	"github.com/shinerio/skillflow/core/platform/settingsstore"
)

type AgentConfig = agentdomain.AgentProfile

type CloudConfig struct {
	Provider            string            `json:"provider"`
	Enabled             bool              `json:"enabled"`
	BucketName          string            `json:"bucketName"`
	RemotePath          string            `json:"remotePath"`
	Credentials         map[string]string `json:"credentials"`
	SyncIntervalMinutes int               `json:"syncIntervalMinutes"` // 0 = on mutation only
}

type CloudProviderConfig struct {
	BucketName  string            `json:"bucketName"`
	RemotePath  string            `json:"remotePath"`
	Credentials map[string]string `json:"credentials"`
}

const DefaultCloudRemotePath = "skillflow/"

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

type WindowState = settingsstore.WindowState

type AppConfig struct {
	SkillsStorageDir      string                         `json:"skillsStorageDir"`
	AutoUpdateSkills      bool                           `json:"autoUpdateSkills"`
	AutoPushAgents        []string                       `json:"autoPushAgents"`
	LaunchAtLogin         bool                           `json:"launchAtLogin"`
	DefaultCategory       string                         `json:"defaultCategory"`
	LogLevel              string                         `json:"logLevel"` // "debug" | "info" | "error"
	RepoScanMaxDepth      int                            `json:"repoScanMaxDepth"`
	SkillStatusVisibility SkillStatusVisibilityConfig    `json:"skillStatusVisibility"`
	Agents                []AgentConfig                  `json:"agents"`
	Cloud                 CloudConfig                    `json:"cloud"`
	CloudProfiles         map[string]CloudProviderConfig `json:"cloudProfiles,omitempty"`
	Proxy                 ProxyConfig                    `json:"proxy"`
	SkippedUpdateVersion  string                         `json:"skippedUpdateVersion,omitempty"` // version tag to suppress startup update prompt
}

const (
	LogLevelDebug           = "debug"
	LogLevelInfo            = "info"
	LogLevelError           = "error"
	DefaultLogLevel         = logging.DefaultLevelString
	DefaultRepoScanMaxDepth = 5
	MinRepoScanMaxDepth     = 1
	MaxRepoScanMaxDepth     = 20
)

func NormalizeLogLevel(level string) string {
	return logging.NormalizeLevelString(level)
}

func NormalizeAgentNameList(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(names))
	normalized := make([]string, 0, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func NormalizeRepoScanMaxDepth(depth int) int {
	if depth < MinRepoScanMaxDepth {
		return DefaultRepoScanMaxDepth
	}
	if depth > MaxRepoScanMaxDepth {
		return MaxRepoScanMaxDepth
	}
	return depth
}

func NormalizeCloudRemotePath(path string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
	if trimmed == "" {
		return DefaultCloudRemotePath
	}

	parts := make([]string, 0)
	for _, part := range strings.Split(trimmed, "/") {
		part = strings.TrimSpace(part)
		if part == "" || part == "." {
			continue
		}
		parts = append(parts, part)
	}

	if len(parts) == 0 {
		return DefaultCloudRemotePath
	}
	if parts[len(parts)-1] != "skillflow" {
		parts = append(parts, "skillflow")
	}
	return strings.Join(parts, "/") + "/"
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

func NormalizeWindowState(state WindowState) WindowState {
	return settingsstore.NormalizeWindowState(state)
}
