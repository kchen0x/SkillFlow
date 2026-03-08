package main

type PushConflict struct {
	SkillID    string `json:"skillId,omitempty"`
	SkillName  string `json:"skillName"`
	SkillPath  string `json:"skillPath,omitempty"`
	ToolName   string `json:"toolName"`
	TargetPath string `json:"targetPath"`
}
