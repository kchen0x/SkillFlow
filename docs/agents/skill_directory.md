# Built-in Agent Skill Directories

This file records the built-in agent scan-directory defaults used by SkillFlow. These paths match `core/agentintegration/domain/defaults.go`; the first entry is also used as the default `PushDir`.

| Agent | Reference | Default scan directories |
|------|-----------|--------------------------|
| Claude Code | [docs](https://code.claude.com/docs/en/skills) | `~/.claude/skills/`, `~/.claude/plugins/marketplaces/` |
| OpenCode | [docs](https://opencode.ai/docs/zh-cn/skills/) | `~/.config/opencode/skills/`, `~/.agents/skills/` |
| Codex | [docs](https://developers.openai.com/codex/skills) | `~/.agents/skills/` |
| Gemini CLI | [docs](https://geminicli.com/docs/cli/skills) | `~/.gemini/skills/`, `~/.agents/skills/` |
| OpenClaw | [docs](https://docs.openclaw.ai/tools/skills) | `~/.openclaw/skills/`, `~/.openclaw/workspace/skills/` |
