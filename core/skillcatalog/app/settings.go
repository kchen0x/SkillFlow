package app

import (
	"strings"

	"github.com/shinerio/skillflow/core/platform/appdata"
)

const DefaultCategoryName = "Default"

type SharedSettings struct {
	DefaultCategory string `json:"defaultCategory"`
}

type LocalSettings struct {
	SkillsStorageDir string `json:"skillsStorageDir"`
}

type Settings struct {
	Shared SharedSettings
	Local  LocalSettings
}

func DefaultSharedSettings() SharedSettings {
	return SharedSettings{
		DefaultCategory: DefaultCategoryName,
	}
}

func DefaultLocalSettings(dataDir string) LocalSettings {
	return LocalSettings{
		SkillsStorageDir: appdata.SkillsDir(dataDir),
	}
}

func DefaultSettings(dataDir string) Settings {
	return Settings{
		Shared: DefaultSharedSettings(),
		Local:  DefaultLocalSettings(dataDir),
	}
}

func NormalizeLocalSettings(settings LocalSettings, dataDir string) LocalSettings {
	if strings.TrimSpace(settings.SkillsStorageDir) == "" {
		settings.SkillsStorageDir = appdata.SkillsDir(dataDir)
	}
	return settings
}
