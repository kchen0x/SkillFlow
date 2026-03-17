package main

import (
	"testing"

	"github.com/shinerio/skillflow/core/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptImportSessionStoreRoundTrip(t *testing.T) {
	store := newPromptImportSessionStore()
	preview := &prompt.ImportPreview{
		Creates: []prompt.ImportPrompt{{Name: "Prompt A", Category: "Default", Content: "Content A"}},
	}

	sessionID := store.Create("import.json", preview)
	require.NotEmpty(t, sessionID)

	session, ok := store.Get(sessionID)
	require.True(t, ok)
	require.NotNil(t, session)
	assert.Equal(t, "import.json", session.FilePath)
	require.NotNil(t, session.Preview)
	assert.Equal(t, "Prompt A", session.Preview.Creates[0].Name)
}

func TestPromptImportSessionStoreDeleteRemovesSession(t *testing.T) {
	store := newPromptImportSessionStore()
	sessionID := store.Create("import.json", &prompt.ImportPreview{})

	store.Delete(sessionID)

	session, ok := store.Get(sessionID)
	assert.False(t, ok)
	assert.Nil(t, session)
}

func TestPromptImportSessionStoreTakeConsumesOneSessionOnly(t *testing.T) {
	store := newPromptImportSessionStore()
	firstID := store.Create("first.json", &prompt.ImportPreview{
		Creates: []prompt.ImportPrompt{{Name: "Prompt A", Category: "Default", Content: "Content A"}},
	})
	secondID := store.Create("second.json", &prompt.ImportPreview{
		Conflicts: []prompt.ImportPrompt{{Name: "Prompt B", Category: "Writing", Content: "Content B"}},
	})

	session, ok := store.Take(firstID)
	require.True(t, ok)
	require.NotNil(t, session)
	assert.Equal(t, "first.json", session.FilePath)

	consumed, ok := store.Get(firstID)
	assert.False(t, ok)
	assert.Nil(t, consumed)

	remaining, ok := store.Get(secondID)
	require.True(t, ok)
	require.NotNil(t, remaining)
	assert.Equal(t, "second.json", remaining.FilePath)
	assert.Equal(t, "Prompt B", remaining.Preview.Conflicts[0].Name)
}
