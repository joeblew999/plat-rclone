package rclone

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// HTTPBackend implements Backend using HTTP calls to a remote rclone RC API.
type HTTPBackend struct {
	BaseURL    string
	HTTPClient *http.Client
	Username   string
	Password   string
}

// NewHTTPBackend creates a new HTTP backend for connecting to a remote rclone instance.
func NewHTTPBackend(baseURL string) *HTTPBackend {
	return &HTTPBackend{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WithAuth sets basic auth credentials.
func (h *HTTPBackend) WithAuth(username, password string) *HTTPBackend {
	h.Username = username
	h.Password = password
	return h
}

// Call implements Backend by making HTTP POST requests to the rclone RC API.
func (h *HTTPBackend) Call(method string, params string) (string, int) {
	url := h.BaseURL + "/" + method

	// rclone RC API always expects JSON, even if empty
	if params == "" {
		params = "{}"
	}
	body := bytes.NewReader([]byte(params))

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return `{"error":"` + err.Error() + `"}`, 500
	}

	req.Header.Set("Content-Type", "application/json")
	if h.Username != "" {
		req.SetBasicAuth(h.Username, h.Password)
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return `{"error":"` + err.Error() + `"}`, 500
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return `{"error":"` + err.Error() + `"}`, 500
	}

	return string(respBody), resp.StatusCode
}
