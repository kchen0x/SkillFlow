package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeGitHubJSONResponseReturnsStatusError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"message":"API rate limit exceeded"}`)),
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	err := decodeGitHubJSONResponse(resp, &release)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "github status 403")
	assert.Contains(t, err.Error(), "API rate limit exceeded")
}

func TestDecodeGitHubJSONResponseDecodesSuccessPayload(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"tag_name":"v1.2.3"}`)),
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	require.NoError(t, decodeGitHubJSONResponse(resp, &release))
	assert.Equal(t, "v1.2.3", release.TagName)
}
