package main

import (
	"testing"

	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
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

func TestGetSkillMetaUsesDaemonServiceForUIProcessRole(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevDaemonInvokeServiceFn := daemonInvokeServiceFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		daemonInvokeServiceFn = prevDaemonInvokeServiceFn
	})

	activeProcessRole = processRoleUI
	expected := &skilldomain.SkillMeta{Name: "Demo", Description: "From daemon"}
	daemonInvokeServiceFn = func(method string, params any, result any) error {
		assert.Equal(t, "GetSkillMeta", method)
		assert.Equal(t, "skill-1", params)
		target, ok := result.(**skilldomain.SkillMeta)
		require.True(t, ok)
		*target = expected
		return nil
	}

	app := NewApp()
	meta, err := app.GetSkillMeta("skill-1")
	require.NoError(t, err)
	assert.Equal(t, expected, meta)
}

func TestGetSkillMetaByPathUsesDaemonServiceForUIProcessRole(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevDaemonInvokeServiceFn := daemonInvokeServiceFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		daemonInvokeServiceFn = prevDaemonInvokeServiceFn
	})

	activeProcessRole = processRoleUI
	expected := &skilldomain.SkillMeta{Name: "DemoByPath"}
	daemonInvokeServiceFn = func(method string, params any, result any) error {
		assert.Equal(t, "GetSkillMetaByPath", method)
		assert.Equal(t, "/tmp/skill", params)
		target, ok := result.(**skilldomain.SkillMeta)
		require.True(t, ok)
		*target = expected
		return nil
	}

	app := NewApp()
	meta, err := app.GetSkillMetaByPath("/tmp/skill")
	require.NoError(t, err)
	assert.Equal(t, expected, meta)
}

func TestReadSkillFileContentUsesDaemonServiceForUIProcessRole(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevDaemonInvokeServiceFn := daemonInvokeServiceFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		daemonInvokeServiceFn = prevDaemonInvokeServiceFn
	})

	activeProcessRole = processRoleUI
	daemonInvokeServiceFn = func(method string, params any, result any) error {
		assert.Equal(t, "ReadSkillFileContent", method)
		assert.Equal(t, "/tmp/skill", params)
		target, ok := result.(*string)
		require.True(t, ok)
		*target = "# demo"
		return nil
	}

	app := NewApp()
	content, err := app.ReadSkillFileContent("/tmp/skill")
	require.NoError(t, err)
	assert.Equal(t, "# demo", content)
}
