package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joeblew999/plat-rclone/pkg/datastar"
)

// Context provides request data and response helpers.
type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	written  bool
}

// NewContext creates a new Context.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{Request: r, Response: w}
}

// Param returns a URL path parameter.
func (c *Context) Param(key string) string {
	return chi.URLParam(c.Request, key)
}

// Query returns a query string parameter.
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// FormValue returns a form field value.
func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

// Header returns a request header value.
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header.
func (c *Context) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

// --- Datastar Integration ---

// IsDatastar returns true if this is a Datastar SSE request.
func (c *Context) IsDatastar() bool {
	return c.Request.Header.Get("Accept") == "text/event-stream"
}

// SSE creates a new SSE writer for Datastar responses.
func (c *Context) SSE() *datastar.SSE {
	return datastar.NewSSE(c.Response, c.Request)
}

// ReadSignals extracts Datastar signals from the request.
func (c *Context) ReadSignals(v any) error {
	return datastar.ReadSignals(c.Request, v)
}

// --- Standard HTTP Responses ---

// HTML writes an HTML response.
func (c *Context) HTML(html string) {
	c.HTMLStatus(http.StatusOK, html)
}

// HTMLStatus writes an HTML response with custom status.
func (c *Context) HTMLStatus(status int, html string) {
	c.written = true
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(status)
	c.Response.Write([]byte(html))
}

// JSON writes a JSON response.
func (c *Context) JSON(data any) {
	c.written = true
	c.Response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Response).Encode(data)
}

// Error writes an error response.
func (c *Context) Error(err error) {
	c.ErrorStatus(http.StatusInternalServerError, err.Error())
}

// ErrorStatus writes an error with custom status.
func (c *Context) ErrorStatus(status int, message string) {
	c.written = true
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(status)
	c.Response.Write([]byte(`<div class="error">` + message + `</div>`))
}

// Redirect sends a redirect response.
func (c *Context) Redirect(url string) {
	c.written = true
	http.Redirect(c.Response, c.Request, url, http.StatusSeeOther)
}

// Bind decodes JSON body into the given struct.
func (c *Context) Bind(v any) error {
	return json.NewDecoder(c.Request.Body).Decode(v)
}
