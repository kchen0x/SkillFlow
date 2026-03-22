package main

import (
	"context"
	"os"
	"time"
)

const (
	memoryExternalWatchDuration = 10 * time.Minute
	memoryExternalWatchInterval = 500 * time.Millisecond
)

type memoryFileStamp struct {
	modTime time.Time
	size    int64
}

func watchMemoryFileChanges(ctx context.Context, path string, interval time.Duration, onChange func()) {
	if path == "" || onChange == nil {
		return
	}
	if interval <= 0 {
		interval = memoryExternalWatchInterval
	}

	current, err := readMemoryFileStamp(path)
	if err != nil {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		deadline := time.NewTimer(memoryExternalWatchDuration)
		defer ticker.Stop()
		defer deadline.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-deadline.C:
				return
			case <-ticker.C:
				next, err := readMemoryFileStamp(path)
				if err != nil {
					continue
				}
				if next != current {
					current = next
					onChange()
				}
			}
		}
	}()
}

func readMemoryFileStamp(path string) (memoryFileStamp, error) {
	info, err := os.Stat(path)
	if err != nil {
		return memoryFileStamp{}, err
	}
	return memoryFileStamp{modTime: info.ModTime().UTC(), size: info.Size()}, nil
}

func (a *App) watchExternalMemoryChanges(memoryType, moduleName, path string) {
	if a == nil || path == "" {
		return
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	watchMemoryFileChanges(ctx, path, memoryExternalWatchInterval, func() {
		payload := map[string]interface{}{"type": memoryType}
		if moduleName != "" {
			payload["name"] = moduleName
		}
		a.emitMemoryEvent(EventMemoryContentChanged, payload)
	})
}
