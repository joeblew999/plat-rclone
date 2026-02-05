package rclone

import (
	"sync"

	"github.com/rclone/rclone/librclone/librclone"
)

// EmbeddedBackend implements Backend using librclone for embedded rclone.
// This provides direct in-process rclone functionality without HTTP overhead.
type EmbeddedBackend struct {
	initOnce sync.Once
}

// NewEmbeddedBackend creates a new embedded backend using librclone.
// The embedded rclone runs in the same process - no external rclone daemon needed.
func NewEmbeddedBackend() *EmbeddedBackend {
	return &EmbeddedBackend{}
}

// Call implements Backend by calling librclone.RPC directly.
func (e *EmbeddedBackend) Call(method string, params string) (string, int) {
	// Initialize librclone on first call (thread-safe)
	e.initOnce.Do(func() {
		librclone.Initialize()
	})

	// rclone RC API always expects JSON, even if empty
	if params == "" {
		params = "{}"
	}

	return librclone.RPC(method, params)
}

// Close finalizes librclone. Call this when done using the embedded backend.
func (e *EmbeddedBackend) Close() {
	librclone.Finalize()
}
