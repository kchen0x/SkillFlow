package snapshot

import (
	"path/filepath"
	"strings"
)

var excludedDirs = []string{
	"cache",
	"runtime",
	"logs",
	"meta_local",
	".git",
}

var excludedFiles = []string{
	".DS_Store",
}

// excludedPatterns are gitignore glob patterns for files that should never be backed up.
var excludedPatterns = []string{
	"*local.json",
}

func ExcludedDirectories() []string {
	return append([]string(nil), excludedDirs...)
}

func ExcludedFiles() []string {
	return append([]string(nil), excludedFiles...)
}

func ExcludedPatterns() []string {
	return append([]string(nil), excludedPatterns...)
}

func ShouldSkipBackupPath(rel string) bool {
	normalized := filepath.ToSlash(filepath.Clean(strings.TrimSpace(rel)))
	if normalized == "." || normalized == "" {
		return false
	}
	for _, dir := range excludedDirs {
		if normalized == dir || strings.HasPrefix(normalized, dir+"/") {
			return true
		}
	}
	base := normalized[strings.LastIndex(normalized, "/")+1:]
	for _, file := range excludedFiles {
		if base == file {
			return true
		}
	}
	if strings.HasSuffix(base, "local.json") {
		return true
	}
	return false
}
