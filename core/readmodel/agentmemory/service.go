package agentmemory

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
)

func LoadPreview(profile agentdomain.AgentProfile) (Preview, error) {
	preview := Preview{
		AgentName:  profile.Name,
		MemoryPath: strings.TrimSpace(profile.MemoryPath),
		RulesDir:   strings.TrimSpace(profile.RulesDir),
		Rules:      []RuleFile{},
	}

	if preview.MemoryPath != "" {
		content, exists, err := readOptionalUTF8(preview.MemoryPath)
		if err != nil {
			return Preview{}, fmt.Errorf("load agent memory preview main file: %w", err)
		}
		preview.MainExists = exists
		preview.MainContent = content
	}

	if preview.RulesDir == "" {
		return preview, nil
	}

	entries, err := os.ReadDir(preview.RulesDir)
	if errors.Is(err, os.ErrNotExist) {
		return preview, nil
	}
	if err != nil {
		return Preview{}, fmt.Errorf("load agent memory preview rules dir: %w", err)
	}
	preview.RulesDirExists = true

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}
		path := filepath.Join(preview.RulesDir, entry.Name())
		content, exists, err := readOptionalUTF8(path)
		if err != nil {
			return Preview{}, fmt.Errorf("load agent memory preview rule file %q: %w", entry.Name(), err)
		}
		if !exists {
			continue
		}
		preview.Rules = append(preview.Rules, RuleFile{
			Name:    entry.Name(),
			Path:    path,
			Content: content,
			Managed: strings.HasPrefix(entry.Name(), "sf-"),
		})
	}

	sort.Slice(preview.Rules, func(i, j int) bool {
		return preview.Rules[i].Name < preview.Rules[j].Name
	})

	return preview, nil
}

func readOptionalUTF8(path string) (string, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		data = data[3:]
	}
	return string(data), true, nil
}
