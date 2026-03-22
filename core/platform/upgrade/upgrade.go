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
	if err := migrateJSONFile(filepath.Join(dataDir, "config_local.json"), func(payload map[string]any) (bool, error) {
		return migrateLocalConfig(payload, dataDir)
	}); err != nil {
		return err
	}
	if err := migrateStarReposFile(filepath.Join(dataDir, "star_repos.json")); err != nil {
		return err
	}
	if err := migrateJSONFile(filepath.Join(dataDir, "memory", "memory_local.json"), migrateMemoryLocalConfig); err != nil {
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
	if _, ok := payload["skillStatusVisibility"]; ok {
		delete(payload, "skillStatusVisibility")
		changed = true
	}
	return changed, nil
}

func migrateLocalConfig(payload map[string]any, dataDir string) (bool, error) {
	changed := false
	changed = renameKey(payload, "tools", "agents") || changed
	changed = renameKey(payload, "autoPushTools", "autoPushAgents") || changed
	if _, ok := payload["repoCacheDir"]; !ok || isEmptyJSONValue(payload["repoCacheDir"]) {
		payload["repoCacheDir"] = filepath.ToSlash(filepath.Join(dataDir, "cache", "repos"))
		changed = true
	}
	if _, ok := payload["skillsStorageDir"]; ok {
		delete(payload, "skillsStorageDir")
		changed = true
	}
	return changed, nil
}

func migrateMemoryLocalConfig(payload map[string]any) (bool, error) {
	if _, ok := payload["modules"]; !ok {
		return false, nil
	}
	delete(payload, "modules")
	return true, nil
}

func migrateStarReposFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}

	var payload []map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("parse %s: %w", filepath.Base(path), err)
	}

	changed := false
	for _, repo := range payload {
		if _, ok := repo["localDir"]; ok {
			delete(repo, "localDir")
			changed = true
		}
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
