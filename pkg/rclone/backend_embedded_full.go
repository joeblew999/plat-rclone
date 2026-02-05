//go:build rclone_full

package rclone

import (
	"sync"

	// Import all rclone backends and operations for full functionality
	_ "github.com/rclone/rclone/backend/all"
	_ "github.com/rclone/rclone/fs/operations" // Registers operations/list, etc.
	"github.com/rclone/rclone/librclone/librclone"
)

// EmbeddedBackend implements Backend using librclone for embedded rclone.
// This full version includes all backends (GDrive, S3, Azure, etc).
// Results in larger binary - use default build for mobile.
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
