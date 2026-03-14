package main

type PushConflict struct {
	SkillID    string `json:"skillId,omitempty"`
	SkillName  string `json:"skillName"`
	SkillPath  string `json:"skillPath,omitempty"`
	AgentName  string `json:"agentName"`
	TargetPath string `json:"targetPath"`
}
