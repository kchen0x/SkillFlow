package upgrade

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Run(dataDir string) error {
	if err := migrateJSONFile(filepath.Join(dataDir, "config.json"), migrateSharedConfig); err != nil {
		return err
	}
	if err := migrateJSONFile(filepath.Join(dataDir, "config_local.json"), migrateLocalConfig); err != nil {
		return err
	}
	return nil
}

func migrateJSONFile(path string, migrate func(map[string]any) (bool, error)) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("parse %s: %w", filepath.Base(path), err)
	}

	changed, err := migrate(payload)
	if err != nil {
		return fmt.Errorf("migrate %s: %w", filepath.Base(path), err)
	}
	if !changed {
		return nil
	}

	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", filepath.Base(path), err)
	}

	return writeFileAtomically(path, encoded, 0o644)
}

func migrateSharedConfig(payload map[string]any) (bool, error) {
	changed := renameKey(payload, "tools", "agents")

	visibility, ok := payload["skillStatusVisibility"]
	if !ok {
		return changed, nil
	}
	visibilityMap, ok := visibility.(map[string]any)
	if !ok {
		return false, fmt.Errorf("skillStatusVisibility must be an object")
	}
	visibilityChanged, err := migrateVisibilityConfig(visibilityMap)
	if err != nil {
		return false, err
	}
	return changed || visibilityChanged, nil
}

func migrateLocalConfig(payload map[string]any) (bool, error) {
	changed := false
	changed = renameKey(payload, "tools", "agents") || changed
	changed = renameKey(payload, "autoPushTools", "autoPushAgents") || changed
	return changed, nil
}

func migrateVisibilityConfig(payload map[string]any) (bool, error) {
	changed := false
	changed = renameKey(payload, "myTools", "myAgents") || changed
	changed = renameKey(payload, "pushToTool", "pushToAgent") || changed
	changed = renameKey(payload, "pullFromTool", "pullFromAgent") || changed

	for key, value := range payload {
		items, ok := value.([]any)
		if !ok {
			continue
		}
		listChanged := false
		for i, item := range items {
			str, ok := item.(string)
			if !ok {
				continue
			}
			if str == "pushedTools" {
				items[i] = "pushedAgents"
				listChanged = true
			}
		}
		if listChanged {
			payload[key] = items
			changed = true
		}
	}

	return changed, nil
}

func renameKey(payload map[string]any, oldKey string, newKey string) bool {
	value, ok := payload[oldKey]
	if !ok {
		return false
	}
	if existing, exists := payload[newKey]; !exists || shouldPromoteLegacyValue(existing, value) {
		payload[newKey] = value
	}
	delete(payload, oldKey)
	return true
}

func shouldPromoteLegacyValue(current any, legacy any) bool {
	return isEmptyJSONValue(current) && !isEmptyJSONValue(legacy)
}

func isEmptyJSONValue(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case []any:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0
	default:
		return false
	}
}

func writeFileAtomically(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()

	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tempPath, perm); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
