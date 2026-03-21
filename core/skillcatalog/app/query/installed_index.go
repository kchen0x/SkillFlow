package query

import (
	"strings"

	"github.com/shinerio/skillflow/core/shared/logicalkey"
	"github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type CorrelationStatus struct {
	LogicalKey    string                   `json:"logicalKey,omitempty"`
	MatchStrength logicalkey.MatchStrength `json:"matchStrength,omitempty"`
	Source        string                   `json:"source,omitempty"`
	Installed     bool                     `json:"installed"`
	Imported      bool                     `json:"imported"`
	Updatable     bool                     `json:"updatable"`
	ContentKeys   []string                 `json:"-"`
}

type InstalledIndex struct {
	byLogical map[string]*installedGroup
	byContent map[string][]*installedGroup
	byName    map[string][]*installedGroup
}

type installedGroup struct {
	LogicalKey string
	Name       string
	Skills     []*domain.InstalledSkill
}

func BuildInstalledIndex(skills []*domain.InstalledSkill) *InstalledIndex {
	idx := &InstalledIndex{
		byLogical: map[string]*installedGroup{},
		byContent: map[string][]*installedGroup{},
		byName:    map[string][]*installedGroup{},
	}

	groupsByID := map[string]*installedGroup{}
	contentGroups := map[string]map[string]*installedGroup{}
	for _, sk := range skills {
		if sk == nil {
			continue
		}
		logicalKey, _ := domain.LogicalKey(sk)
		groupID := logicalKey
		if groupID == "" {
			groupID = "instance:" + sk.ID
		}

		group := groupsByID[groupID]
		if group == nil {
			group = &installedGroup{
				LogicalKey: logicalKey,
				Name:       sk.Name,
			}
			groupsByID[groupID] = group
			if logicalKey != "" {
				idx.byLogical[logicalKey] = group
			}
		}
		group.Skills = append(group.Skills, sk)
		if strings.TrimSpace(group.Name) == "" {
			group.Name = sk.Name
		}

		if strings.TrimSpace(sk.Path) != "" {
			if contentKey, err := logicalkey.ContentFromDir(sk.Path); err == nil && strings.TrimSpace(contentKey) != "" {
				if contentGroups[contentKey] == nil {
					contentGroups[contentKey] = map[string]*installedGroup{}
				}
				contentGroups[contentKey][groupID] = group
			}
		}
	}

	for _, group := range groupsByID {
		nameKey := normalizedName(group.Name)
		if nameKey == "" {
			continue
		}
		idx.byName[nameKey] = append(idx.byName[nameKey], group)
	}

	for contentKey, groups := range contentGroups {
		for _, group := range groups {
			idx.byContent[contentKey] = append(idx.byContent[contentKey], group)
		}
	}

	return idx
}

func (idx *InstalledIndex) Resolve(name, logicalKey string) CorrelationStatus {
	if idx == nil {
		return CorrelationStatus{LogicalKey: logicalKey}
	}

	if logicalKey != "" {
		if group, ok := idx.byLogical[logicalKey]; ok {
			return group.status(logicalKey, logicalkey.MatchStrengthLogical)
		}
		if groups := idx.byContent[logicalKey]; len(groups) == 1 {
			return groups[0].status(coalesceLogicalKey(groups[0].LogicalKey, logicalKey), logicalkey.MatchStrengthContent)
		}
	}

	nameKey := normalizedName(name)
	if nameKey == "" {
		return CorrelationStatus{LogicalKey: logicalKey}
	}
	if groups := idx.byName[nameKey]; len(groups) == 1 {
		return groups[0].status(coalesceLogicalKey(logicalKey, groups[0].LogicalKey), logicalkey.MatchStrengthFallback)
	}

	return CorrelationStatus{LogicalKey: logicalKey}
}

func (idx *InstalledIndex) IsInstalled(name, logicalKey string) bool {
	return idx.Resolve(name, logicalKey).Installed
}

func (g *installedGroup) status(logicalKey string, matchStrength logicalkey.MatchStrength) CorrelationStatus {
	status := CorrelationStatus{
		LogicalKey:    logicalKey,
		MatchStrength: matchStrength,
		Source:        g.source(),
		Installed:     true,
		Imported:      true,
		ContentKeys:   g.contentKeys(),
	}
	for _, sk := range g.Skills {
		if sk != nil && sk.HasUpdate() {
			status.Updatable = true
			break
		}
	}
	return status
}

func (g *installedGroup) source() string {
	for _, sk := range g.Skills {
		if sk == nil {
			continue
		}
		if source := strings.TrimSpace(string(sk.Source)); source != "" {
			return source
		}
	}
	return ""
}

func (g *installedGroup) contentKeys() []string {
	var keys []string
	for _, sk := range g.Skills {
		if sk == nil || strings.TrimSpace(sk.Path) == "" {
			continue
		}
		contentKey, err := logicalkey.ContentFromDir(sk.Path)
		if err != nil || strings.TrimSpace(contentKey) == "" {
			continue
		}
		keys = appendUniqueContentKey(keys, contentKey)
	}
	return keys
}

func appendUniqueContentKey(keys []string, key string) []string {
	for _, existing := range keys {
		if existing == key {
			return keys
		}
	}
	return append(keys, key)
}

func normalizedName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func coalesceLogicalKey(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return secondary
}
