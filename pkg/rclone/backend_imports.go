//go:build !rclone_full

package rclone

// Default build: local backend only (smaller binary, faster compile)
import (
	_ "github.com/rclone/rclone/backend/local"
	_ "github.com/rclone/rclone/fs/operations" // Registers operations/list, etc.
)
