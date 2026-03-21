package domain

import "time"

type StarRepo struct {
	URL       string    `json:"url"`
	Name      string    `json:"name"`
	Source    string    `json:"source"`
	LocalDir  string    `json:"localDir"`
	LastSync  time.Time `json:"lastSync"`
	SyncError string    `json:"syncError,omitempty"`
}

type SourceSkillCandidate struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	SubPath    string `json:"subPath"`
	RepoURL    string `json:"repoUrl"`
	RepoName   string `json:"repoName"`
	Source     string `json:"source"`
	LogicalKey string `json:"logicalKey,omitempty"`
}
