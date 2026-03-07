package pathutil

import (
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

// NormalizeStoredRelativePath normalizes a serialized relative path to a
// forward-slash form so it can round-trip across platforms.
func NormalizeStoredRelativePath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	normalized := strings.ReplaceAll(trimmed, `\`, "/")
	return path.Clean(normalized)
}

// IsSafeStoredRelativePath reports whether rel stays within the storage root.
func IsSafeStoredRelativePath(rel string) bool {
	if rel == "" || rel == "." || rel == ".." || path.IsAbs(rel) {
		return false
	}
	return !strings.HasPrefix(rel, "../")
}

// IsPortableAbsolutePath detects both native absolute paths and serialized
// absolute paths from other platforms (for example Windows drive-letter paths
// read on macOS/Linux).
func IsPortableAbsolutePath(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false
	}
	if filepath.IsAbs(trimmed) || strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, `\\`) {
		return true
	}
	return len(trimmed) >= 3 && unicode.IsLetter(rune(trimmed[0])) && trimmed[1] == ':' && (trimmed[2] == '\\' || trimmed[2] == '/')
}

// RelativeToBase converts a native absolute path under base to a portable
// forward-slash relative path.
func RelativeToBase(base, target string) (string, bool) {
	cleanTarget := strings.TrimSpace(target)
	if !filepath.IsAbs(cleanTarget) {
		return "", false
	}
	rel, err := filepath.Rel(base, cleanTarget)
	if err != nil {
		return "", false
	}
	rel = NormalizeStoredRelativePath(filepath.ToSlash(rel))
	if !IsSafeStoredRelativePath(rel) {
		return "", false
	}
	return rel, true
}

// ResolveStoredPath converts a serialized path to an absolute runtime path. It
// also reports whether the serialized form should be rewritten to the normalized
// relative format.
func ResolveStoredPath(base, storedPath, fallbackAbs string) (string, bool) {
	storedPath = strings.TrimSpace(storedPath)
	if storedPath == "" {
		if fallbackAbs != "" {
			return filepath.Clean(fallbackAbs), true
		}
		return "", false
	}
	if rel, ok := RelativeToBase(base, storedPath); ok {
		return filepath.Join(base, filepath.FromSlash(rel)), true
	}
	if IsPortableAbsolutePath(storedPath) {
		if fallbackAbs != "" {
			return filepath.Clean(fallbackAbs), true
		}
		return filepath.Clean(storedPath), false
	}
	rel := NormalizeStoredRelativePath(storedPath)
	if !IsSafeStoredRelativePath(rel) {
		if fallbackAbs != "" {
			return filepath.Clean(fallbackAbs), true
		}
		return "", true
	}
	return filepath.Clean(filepath.Join(base, filepath.FromSlash(rel))), rel != storedPath
}

// StorePath returns the portable relative path that should be persisted for a
// runtime absolute path.
func StorePath(base, currentAbs, fallbackAbs string) string {
	if rel, ok := RelativeToBase(base, currentAbs); ok {
		return rel
	}
	if rel, ok := RelativeToBase(base, fallbackAbs); ok {
		return rel
	}
	rel := NormalizeStoredRelativePath(currentAbs)
	if IsSafeStoredRelativePath(rel) {
		return rel
	}
	rel = NormalizeStoredRelativePath(fallbackAbs)
	if IsSafeStoredRelativePath(rel) {
		return rel
	}
	return ""
}
