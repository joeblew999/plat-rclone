package rclone

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// LatestReleaseURL is the GitHub API endpoint for latest release
	LatestReleaseURL = "https://github.com/rclone/rclone/releases/latest"
	// DownloadBaseURL is the base URL for rclone downloads
	DownloadBaseURL = "https://downloads.rclone.org"
)

// DownloadOptions configures the rclone download
type DownloadOptions struct {
	Version string // e.g., "v1.73.0" or "current" for latest
	OS      string // e.g., "osx", "linux", "windows"
	Arch    string // e.g., "arm64", "amd64"
	Dir     string // Directory to install to
}

// DefaultDownloadOptions returns options for the current platform
func DefaultDownloadOptions() DownloadOptions {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go OS names to rclone names
	switch osName {
	case "darwin":
		osName = "osx"
	}

	// Map Go arch names to rclone names
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	}

	return DownloadOptions{
		Version: "current",
		OS:      osName,
		Arch:    arch,
		Dir:     ".",
	}
}

// DownloadURL returns the download URL for rclone
func (o DownloadOptions) DownloadURL() string {
	// Format: https://downloads.rclone.org/v1.73.0/rclone-v1.73.0-osx-arm64.zip
	// Or: https://downloads.rclone.org/rclone-current-osx-arm64.zip
	name := fmt.Sprintf("rclone-%s-%s-%s.zip", o.Version, o.OS, o.Arch)
	if o.Version == "current" {
		return fmt.Sprintf("%s/%s", DownloadBaseURL, name)
	}
	return fmt.Sprintf("%s/%s/%s", DownloadBaseURL, o.Version, name)
}

// Download downloads and extracts rclone to the specified directory
func Download(opts DownloadOptions) (string, error) {
	url := opts.DownloadURL()
	fmt.Printf("Downloading rclone from %s\n", url)

	// Download the zip file
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Create temp file for the zip
	tmpFile, err := os.CreateTemp("", "rclone-*.zip")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy download to temp file
	size, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("save download: %w", err)
	}
	fmt.Printf("Downloaded %d bytes\n", size)

	// Extract the zip
	binaryPath, err := extractRclone(tmpFile.Name(), opts.Dir)
	if err != nil {
		return "", fmt.Errorf("extract: %w", err)
	}

	// Make executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return "", fmt.Errorf("chmod: %w", err)
	}

	fmt.Printf("Installed rclone to %s\n", binaryPath)
	return binaryPath, nil
}

func extractRclone(zipPath, destDir string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var binaryPath string

	for _, f := range r.File {
		// Look for the rclone binary (in rclone-*/rclone or rclone-*/rclone.exe)
		name := filepath.Base(f.Name)
		if name == "rclone" || name == "rclone.exe" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}

			binaryPath = filepath.Join(destDir, name)
			outFile, err := os.Create(binaryPath)
			if err != nil {
				rc.Close()
				return "", err
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			if err != nil {
				return "", err
			}
			break
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("rclone binary not found in archive")
	}

	return binaryPath, nil
}

// FindRclone looks for rclone in common locations
func FindRclone() string {
	// Check current directory first
	if _, err := os.Stat("./rclone"); err == nil {
		return "./rclone"
	}

	// Check PATH
	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	for _, p := range paths {
		rclonePath := filepath.Join(p, "rclone")
		if _, err := os.Stat(rclonePath); err == nil {
			return rclonePath
		}
	}

	return ""
}

// EnsureRclone ensures rclone is available, downloading if necessary
func EnsureRclone(dir string) (string, error) {
	// Check if already exists
	if path := FindRclone(); path != "" {
		return path, nil
	}

	// Download to specified directory
	opts := DefaultDownloadOptions()
	opts.Dir = dir
	return Download(opts)
}
