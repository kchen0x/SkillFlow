package agentmemory

type Preview struct {
	AgentName      string
	MemoryPath     string
	RulesDir       string
	MainExists     bool
	MainContent    string
	RulesDirExists bool
	Rules          []RuleFile
}

type RuleFile struct {
	Name    string
	Path    string
	Content string
	Managed bool
}
