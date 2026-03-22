package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatchMemoryFileChangesInvokesCallbackWhenFileChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.md")
	require.NoError(t, os.WriteFile(path, []byte("before"), 0o644))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	changed := make(chan struct{}, 1)
	watchMemoryFileChanges(ctx, path, 10*time.Millisecond, func() {
		select {
		case changed <- struct{}{}:
		default:
		}
	})

	require.Eventually(t, func() bool {
		return os.WriteFile(path, []byte("after"), 0o644) == nil
	}, time.Second, 20*time.Millisecond)

	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("expected file watcher callback after file content changed")
	}
}
