package viewstate

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerLoadReturnsHitForValidSnapshot(t *testing.T) {
	manager := NewManager(t.TempDir())
	require.NoError(t, manager.Save("installed_skills", "fp-1", []string{"alpha", "beta"}))

	var payload []string
	state, err := manager.Load("installed_skills", "fp-1", &payload)
	require.NoError(t, err)
	assert.Equal(t, StateHit, state)
	assert.Equal(t, []string{"alpha", "beta"}, payload)
}

func TestManagerLoadReturnsStaleOnSchemaMismatch(t *testing.T) {
	root := t.TempDir()
	manager := NewManager(root)
	require.NoError(t, manager.Save("installed_skills", "fp-1", []string{"alpha"}))

	newerManager := &Manager{root: root, schemaVersion: CurrentSchemaVersion + 1}
	var payload []string
	state, err := newerManager.Load("installed_skills", "fp-1", &payload)
	require.NoError(t, err)
	assert.Equal(t, StateStale, state)
	assert.Empty(t, payload)
}

func TestManagerLoadReturnsStaleOnFingerprintMismatch(t *testing.T) {
	manager := NewManager(t.TempDir())
	require.NoError(t, manager.Save("installed_skills", "fp-1", []string{"alpha"}))

	var payload []string
	state, err := manager.Load("installed_skills", "fp-2", &payload)
	require.NoError(t, err)
	assert.Equal(t, StateStale, state)
	assert.Empty(t, payload)
}

func TestManagerLoadTreatsCorruptSnapshotAsMiss(t *testing.T) {
	manager := NewManager(t.TempDir())
	path := manager.path("installed_skills")
	require.NoError(t, os.WriteFile(path, []byte("{not-json"), 0644))

	var payload []string
	state, err := manager.Load("installed_skills", "fp-1", &payload)
	require.NoError(t, err)
	assert.Equal(t, StateMiss, state)
	assert.Empty(t, payload)
}

func TestManagerSaveWritesSchemaVersionFingerprintAndPayload(t *testing.T) {
	manager := NewManager(t.TempDir())
	require.NoError(t, manager.Save("installed_skills", "fp-1", []string{"alpha"}))

	data, err := os.ReadFile(manager.path("installed_skills"))
	require.NoError(t, err)

	var record snapshotRecord
	require.NoError(t, json.Unmarshal(data, &record))
	assert.Equal(t, CurrentSchemaVersion, record.SchemaVersion)
	assert.Equal(t, "fp-1", record.Fingerprint)

	var payload []string
	require.NoError(t, json.Unmarshal(record.Payload, &payload))
	assert.Equal(t, []string{"alpha"}, payload)
}
