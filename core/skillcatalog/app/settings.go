package app

import "github.com/shinerio/skillflow/core/platform/appdata"

const DefaultCategoryName = "Default"

type SharedSettings struct {
	DefaultCategory string `json:"defaultCategory"`
}

type LocalSettings struct{}

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
	_ = appdata.SkillsDir(dataDir)
	return LocalSettings{}
}

func DefaultSettings(dataDir string) Settings {
	return Settings{
		Shared: DefaultSharedSettings(),
		Local:  DefaultLocalSettings(dataDir),
	}
}

func NormalizeLocalSettings(settings LocalSettings, _ string) LocalSettings {
	return settings
}
