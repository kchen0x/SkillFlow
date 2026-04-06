package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGitConflictPendingUsesDaemonServiceForUIProcessRole(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevDaemonInvokeServiceFn := daemonInvokeServiceFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		daemonInvokeServiceFn = prevDaemonInvokeServiceFn
	})

	activeProcessRole = processRoleUI
	daemonInvokeServiceFn = func(method string, params any, result any) error {
		assert.Equal(t, "GetGitConflictPending", method)
		assert.Nil(t, params)
		target, ok := result.(*bool)
		require.True(t, ok)
		*target = true
		return nil
	}

	app := NewApp()
	assert.True(t, app.GetGitConflictPending())
}
