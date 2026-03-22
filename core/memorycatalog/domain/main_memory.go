package domain

import "time"

// MainMemory is the single global memory file (main.md).
type MainMemory struct {
	Content   string
	UpdatedAt time.Time
}
