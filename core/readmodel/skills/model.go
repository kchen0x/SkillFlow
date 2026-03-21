package skills

import agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"

const (
	InstalledSkillsSnapshotName = "installed_skills"
	AllStarSkillsSnapshotName   = "all_star_skills"
	AgentPresenceSnapshotName   = "agent_presence"
)

type InstalledSkillEntry struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Category     string   `json:"category"`
	Source       string   `json:"source"`
	SourceSHA    string   `json:"sourceSha"`
	LatestSHA    string   `json:"latestSha"`
	Updatable    bool     `json:"updatable"`
	Pushed       bool     `json:"pushed"`
	PushedAgents []string `json:"pushedAgents"`
}

type StarSkillEntry struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	SubPath      string   `json:"subPath"`
	RepoURL      string   `json:"repoUrl"`
	RepoName     string   `json:"repoName"`
	Source       string   `json:"source"`
	LogicalKey   string   `json:"logicalKey"`
	Installed    bool     `json:"installed"`
	Imported     bool     `json:"imported"`
	Updatable    bool     `json:"updatable"`
	Pushed       bool     `json:"pushed"`
	PushedAgents []string `json:"pushedAgents"`
}

type InstalledSkillsInput struct {
	DefaultCategory     string
	RepoScanMaxDepth    int
	AgentProfiles       []agentdomain.AgentProfile
	SnapshotFingerprint string
}

type StarSkillsInput struct {
	RepoScanMaxDepth    int
	AgentProfiles       []agentdomain.AgentProfile
	SnapshotFingerprint string
}
