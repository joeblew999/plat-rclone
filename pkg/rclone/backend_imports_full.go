//go:build rclone_full

package rclone

// Full build: all backends (S3, GDrive, Azure, Dropbox, etc.)
// Results in larger binary (~100MB) and slower compile.
// Use: go build -tags=rclone_full
import (
	_ "github.com/rclone/rclone/backend/all"
	_ "github.com/rclone/rclone/fs/operations" // Registers operations/list, etc.
)
