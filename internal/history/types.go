package history

import "time"

type HistoryEntry struct {
	Profile   string    `json:"profile"`
	Operation string    `json:"operation"`
	Key       string    `json:"key"`
	ValueHash string    `json:"value_hash,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type ProfileSnapshot struct {
	Profile   string
	Keys      map[string]KeySnapshot
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type KeySnapshot struct {
	Op        string    `json:"op"`
	ValueHash string    `json:"value_hash,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}
