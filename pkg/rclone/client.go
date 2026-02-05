// Package rclone provides a client for the rclone RC API.
package rclone

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an rclone RC API client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Username   string
	Password   string
}

// NewClient creates a new rclone RC client.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WithAuth sets basic auth credentials.
func (c *Client) WithAuth(username, password string) *Client {
	c.Username = username
	c.Password = password
	return c
}

// Call makes an RC API call.
func (c *Client) Call(method string, params any) (json.RawMessage, error) {
	url := c.BaseURL + "/" + method

	var body io.Reader
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rc error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
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
	resp, err := c.Call("config/listremotes", nil)
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
	resp, err := c.Call("config/get", map[string]string{"name": name})
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
	_, err := c.Call("config/delete", map[string]string{"name": name})
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

	resp, err := c.Call("operations/list", map[string]any{
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
	_, err := c.Call("operations/mkdir", map[string]string{
		"fs":     remote + ":",
		"remote": path,
	})
	return err
}

// Delete deletes a file.
func (c *Client) Delete(remote, path string) error {
	_, err := c.Call("operations/deletefile", map[string]string{
		"fs":     remote + ":",
		"remote": path,
	})
	return err
}

// Purge deletes a directory and all contents.
func (c *Client) Purge(remote, path string) error {
	_, err := c.Call("operations/purge", map[string]string{
		"fs":     remote + ":",
		"remote": path,
	})
	return err
}

// --- Core Operations ---

// Version returns rclone version info.
func (c *Client) Version() (map[string]any, error) {
	resp, err := c.Call("core/version", nil)
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
	resp, err := c.Call("core/stats", nil)
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
	_, err := c.Call("sync/copy", map[string]string{
		"srcFs": srcRemote + ":" + srcPath,
		"dstFs": dstRemote + ":" + dstPath,
	})
	return err
}

// Move moves files from source to destination.
func (c *Client) Move(srcRemote, srcPath, dstRemote, dstPath string) error {
	_, err := c.Call("sync/move", map[string]string{
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
	resp, err := c.Call("job/list", nil)
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
	resp, err := c.Call("job/status", map[string]int64{"jobid": id})
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
	_, err := c.Call("job/stop", map[string]int64{"jobid": id})
	return err
}
