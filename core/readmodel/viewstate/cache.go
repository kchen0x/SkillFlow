package viewstate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Manager struct {
	root          string
	schemaVersion int
}

func NewManager(root string) *Manager {
	return &Manager{
		root:          root,
		schemaVersion: CurrentSchemaVersion,
	}
}

func (m *Manager) Load(name, fingerprint string, out any) (State, error) {
	data, err := os.ReadFile(m.path(name))
	if err != nil {
		if os.IsNotExist(err) {
			return StateMiss, nil
		}
		return StateMiss, err
	}

	var record snapshotRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return StateMiss, nil
	}
	if record.SchemaVersion != m.schemaVersion || record.Fingerprint != fingerprint {
		return StateStale, nil
	}
	if out == nil {
		return StateHit, nil
	}
	if err := json.Unmarshal(record.Payload, out); err != nil {
		return StateMiss, nil
	}
	return StateHit, nil
}

func (m *Manager) Save(name, fingerprint string, payload any) error {
	if err := os.MkdirAll(m.root, 0755); err != nil {
		return err
	}

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	recordData, err := json.MarshalIndent(snapshotRecord{
		SchemaVersion: m.schemaVersion,
		Fingerprint:   fingerprint,
		BuiltAt:       time.Now().UTC(),
		Payload:       encodedPayload,
	}, "", "  ")
	if err != nil {
		return err
	}

	path := m.path(name)
	tmp, err := os.CreateTemp(m.root, strings.TrimSuffix(filepath.Base(path), ".json")+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(recordData); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	_ = os.Remove(path)
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func (m *Manager) path(name string) string {
	if strings.HasSuffix(name, ".json") {
		return filepath.Join(m.root, name)
	}
	return filepath.Join(m.root, name+".json")
}
