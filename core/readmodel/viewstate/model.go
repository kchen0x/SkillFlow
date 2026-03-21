package viewstate

import (
	"encoding/json"
	"time"
)

const CurrentSchemaVersion = 1

type State string

const (
	StateMiss  State = "miss"
	StateHit   State = "hit"
	StateStale State = "stale"
)

type snapshotRecord struct {
	SchemaVersion int             `json:"schemaVersion"`
	Fingerprint   string          `json:"fingerprint"`
	BuiltAt       time.Time       `json:"builtAt"`
	Payload       json.RawMessage `json:"payload"`
}
