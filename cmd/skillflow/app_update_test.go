package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckAppUpdateUsesBackgroundContextWhenAppContextIsNil(t *testing.T) {
	prevNewRequestWithContextFn := newRequestWithContextFn
	t.Cleanup(func() {
		newRequestWithContextFn = prevNewRequestWithContextFn
	})

	sentinel := errors.New("stop after context assertion")
	newRequestWithContextFn = func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
		require.NotNil(t, ctx)
		assert.Equal(t, http.MethodGet, method)
		assert.Contains(t, url, "/releases/latest")
		return nil, sentinel
	}

	info, err := NewApp().CheckAppUpdate()
	require.ErrorIs(t, err, sentinel)
	assert.Nil(t, info)
}
