// Package rclone provides a client for the rclone RC API.
// It supports both embedded librclone and remote HTTP backends.
package rclone

import (
	"encoding/json"
	"fmt"
)

// Client wraps a Backend with all the rclone business logic.
// The same methods work whether using embedded librclone or HTTP remote.
type Client struct {
	backend Backend
}

// NewClient creates a new rclone client using HTTP to connect to a remote rclone instance.
// This is the traditional approach - connect to rclone running with `rclone rcd`.
func NewClient(baseURL string) *Client {
	return &Client{backend: NewHTTPBackend(baseURL)}
}

// NewClientWithAuth creates a new HTTP client with basic auth.
func NewClientWithAuth(baseURL, username, password string) *Client {
	return &Client{backend: NewHTTPBackend(baseURL).WithAuth(username, password)}
}

// NewEmbedded creates a new rclone client using embedded librclone.
// No external rclone daemon needed - rclone runs in-process.
func NewEmbedded() *Client {
	return &Client{backend: NewEmbeddedBackend()}
}

// WithAuth sets basic auth credentials (only works with HTTP backend).
func (c *Client) WithAuth(username, password string) *Client {
	if h, ok := c.backend.(*HTTPBackend); ok {
		h.WithAuth(username, password)
	}
	return c
}

// Close releases resources. Should be called when using embedded backend.
func (c *Client) Close() {
	if e, ok := c.backend.(*EmbeddedBackend); ok {
		e.Close()
	}
}

// call makes an RC API call and returns the JSON response.
func (c *Client) call(method string, params any) (json.RawMessage, error) {
	// Encode params to JSON
	var paramsJSON string
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params: %w", err)
		}
		paramsJSON = string(data)
	}

	// Make the call via backend
	resp, status := c.backend.Call(method, paramsJSON)

	// Check for errors
	if status != 200 {
		return nil, fmt.Errorf("rc error %d: %s", status, resp)
	}

	return json.RawMessage(resp), nil
}

// --- Config Operations ---

// Remote represents an rclone remote configuration.
type Remote struct {
	Name string            `json:"name"`
	Type string            `json:"type"`
	Opt  map[string]string `json:"opt,omitempty"`
}

// ListRemotes returns all configured remotes.
func (c *Client) ListRemotes() ([]string, error) {
	resp, err := c.call("config/listremotes", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Remotes []string `json:"remotes"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("unmarshal remotes: %w", err)
	}
	return result.Remotes, nil
}

// GetRemote returns configuration for a specific remote.
func (c *Client) GetRemote(name string) (map[string]string, error) {
	resp, err := c.call("config/get", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}

	var result map[string]string
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return result, nil
}

// DeleteRemote deletes a remote configuration.
func (c *Client) DeleteRemote(name string) error {
	_, err := c.call("config/delete", map[string]string{"name": name})
	return err
}

// --- Operations ---

// ListItem represents a file or directory in a listing.
type ListItem struct {
	Path    string `json:"Path"`
	Name    string `json:"Name"`
	Size    int64  `json:"Size"`
	ModTime string `json:"ModTime"`
	IsDir   bool   `json:"IsDir"`
}

// List lists files in a remote path.
func (c *Client) List(remote, path string) ([]ListItem, error) {
	fs := remote + ":"
	if path != "" {
		fs += path
	}

	resp, err := c.call("operations/list", map[string]any{
		"fs":     fs,
		"remote": "",
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		List []ListItem `json:"list"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("unmarshal list: %w", err)
	}
	return result.List, nil
}

// Mkdir creates a directory.
func (c *Client) Mkdir(remote, path string) error {
	_, err := c.call("operations/mkdir", map[string]string{
		"fs":     remote + ":",
		"remote": path,
	})
	return err
}

// Delete deletes a file.
func (c *Client) Delete(remote, path string) error {
	_, err := c.call("operations/deletefile", map[string]string{
		"fs":     remote + ":",
		"remote": path,
	})
	return err
}

// Purge deletes a directory and all contents.
func (c *Client) Purge(remote, path string) error {
	_, err := c.call("operations/purge", map[string]string{
		"fs":     remote + ":",
		"remote": path,
	})
	return err
}

// --- Core Operations ---

// Version returns rclone version info.
func (c *Client) Version() (map[string]any, error) {
	resp, err := c.call("core/version", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("unmarshal version: %w", err)
	}
	return result, nil
}

// Stats returns current transfer statistics.
func (c *Client) Stats() (map[string]any, error) {
	resp, err := c.call("core/stats", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("unmarshal stats: %w", err)
	}
	return result, nil
}

// --- Sync/Copy Operations ---

// Copy copies files from source to destination.
func (c *Client) Copy(srcRemote, srcPath, dstRemote, dstPath string) error {
	_, err := c.call("sync/copy", map[string]string{
		"srcFs": srcRemote + ":" + srcPath,
		"dstFs": dstRemote + ":" + dstPath,
	})
	return err
}

// Move moves files from source to destination.
func (c *Client) Move(srcRemote, srcPath, dstRemote, dstPath string) error {
	_, err := c.call("sync/move", map[string]string{
		"srcFs": srcRemote + ":" + srcPath,
		"dstFs": dstRemote + ":" + dstPath,
	})
	return err
}

// --- Job Operations ---

// Job represents an rclone job.
type Job struct {
	ID        int64  `json:"id"`
	Group     string `json:"group"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime,omitempty"`
	Error     string `json:"error,omitempty"`
	Finished  bool   `json:"finished"`
	Success   bool   `json:"success"`
}

// ListJobs returns all current jobs.
func (c *Client) ListJobs() ([]Job, error) {
	resp, err := c.call("job/list", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		JobIDs []int64 `json:"jobids"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("unmarshal jobs: %w", err)
	}

	// Get details for each job
	jobs := make([]Job, 0, len(result.JobIDs))
	for _, id := range result.JobIDs {
		job, err := c.GetJob(id)
		if err == nil {
			jobs = append(jobs, *job)
		}
	}
	return jobs, nil
}

// GetJob returns details of a specific job.
func (c *Client) GetJob(id int64) (*Job, error) {
	resp, err := c.call("job/status", map[string]int64{"jobid": id})
	if err != nil {
		return nil, err
	}

	var job Job
	if err := json.Unmarshal(resp, &job); err != nil {
		return nil, fmt.Errorf("unmarshal job: %w", err)
	}
	job.ID = id
	return &job, nil
}

// StopJob stops a running job.
func (c *Client) StopJob(id int64) error {
	_, err := c.call("job/stop", map[string]int64{"jobid": id})
	return err
}
