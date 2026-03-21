package app

import "strings"

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

type Settings struct {
	Cloud         CloudConfig                    `json:"cloud"`
	CloudProfiles map[string]CloudProviderConfig `json:"cloudProfiles,omitempty"`
}

const DefaultCloudRemotePath = "skillflow/"

func DefaultSettings() Settings {
	return Settings{
		Cloud: CloudConfig{
			RemotePath: DefaultCloudRemotePath,
		},
	}
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
