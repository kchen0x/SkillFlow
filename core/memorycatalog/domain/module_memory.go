package domain

import (
	"time"
)

// ModuleMemory is a single topic memory file under rules/.
// Name is the filename without .md suffix and also serves as ID.
type ModuleMemory struct {
	Name      string // e.g. "coding-style"
	Content   string
	UpdatedAt time.Time
}

// ValidateModuleName checks that name matches [a-z][a-z0-9-]{0,63} with no trailing hyphen.
func ValidateModuleName(name string) error {
	if len(name) == 0 || len(name) > 64 {
		return ErrInvalidModuleName
	}
	runes := []rune(name)
	// Must start with lowercase ASCII letter
	if runes[0] < 'a' || runes[0] > 'z' {
		return ErrInvalidModuleName
	}
	// Must not end with hyphen
	if runes[len(runes)-1] == '-' {
		return ErrInvalidModuleName
	}
	// All chars must be a-z, 0-9, or -
	for _, r := range runes {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return ErrInvalidModuleName
		}
	}
	return nil
}
