package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func nearestExistingDirectory(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	current := filepath.Clean(trimmed)
	for {
		info, err := os.Stat(current)
		if err == nil {
			if info.IsDir() {
				return current
			}
			parent := filepath.Dir(current)
			if parent == current {
				return ""
			}
			return parent
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func resolveOpenPathTarget(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("path is empty")
	}
	cleaned := filepath.Clean(trimmed)
	info, err := os.Stat(cleaned)
	if err == nil {
		if info.IsDir() {
			return cleaned, nil
		}
		return filepath.Dir(cleaned), nil
	}
	fallback := nearestExistingDirectory(cleaned)
	if fallback == "" {
		return "", fmt.Errorf("path not found: %s", trimmed)
	}
	return fallback, nil
}
