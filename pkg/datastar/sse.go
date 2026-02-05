// Package datastar provides Server-Sent Events (SSE) support for Datastar integration.
package datastar

import (
	"bytes"
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
)

// SSE wraps the datastar ServerSentEventGenerator with convenience methods.
type SSE struct {
	*datastar.ServerSentEventGenerator
}

// NewSSE creates a new SSE writer for streaming responses to the client.
func NewSSE(w http.ResponseWriter, r *http.Request, opts ...datastar.SSEOption) *SSE {
	return &SSE{
		ServerSentEventGenerator: datastar.NewSSE(w, r, opts...),
	}
}

// PatchTempl renders a templ component and patches it into the DOM.
func (s *SSE) PatchTempl(c templ.Component, opts ...datastar.PatchElementOption) error {
	return s.ServerSentEventGenerator.PatchElementTempl(c, opts...)
}

// PatchTemplByID renders a templ component and patches it into a specific element by ID.
func (s *SSE) PatchTemplByID(id string, c templ.Component, opts ...datastar.PatchElementOption) error {
	opts = append(opts, datastar.WithSelectorID(id))
	return s.ServerSentEventGenerator.PatchElementTempl(c, opts...)
}

// PatchHTML patches raw HTML into the DOM.
func (s *SSE) PatchHTML(html string, opts ...datastar.PatchElementOption) error {
	return s.ServerSentEventGenerator.PatchElements(html, opts...)
}

// PatchHTMLByID patches raw HTML into a specific element by ID.
func (s *SSE) PatchHTMLByID(id string, html string, opts ...datastar.PatchElementOption) error {
	opts = append(opts, datastar.WithSelectorID(id))
	return s.ServerSentEventGenerator.PatchElements(html, opts...)
}

// AppendTemplByID appends a templ component inside a specific element by ID.
func (s *SSE) AppendTemplByID(id string, c templ.Component, opts ...datastar.PatchElementOption) error {
	opts = append(opts, datastar.WithSelectorID(id), datastar.WithModeAppend())
	return s.ServerSentEventGenerator.PatchElementTempl(c, opts...)
}

// RemoveByID removes an element by its ID.
func (s *SSE) RemoveByID(id string) error {
	return s.ServerSentEventGenerator.RemoveElementByID(id)
}

// PatchSignals updates client-side signals with the provided data.
func (s *SSE) PatchSignals(signals any, opts ...datastar.PatchSignalsOption) error {
	return s.ServerSentEventGenerator.MarshalAndPatchSignals(signals, opts...)
}

// Redirect navigates the client to a new URL.
func (s *SSE) Redirect(url string, opts ...datastar.ExecuteScriptOption) error {
	return s.ServerSentEventGenerator.Redirect(url, opts...)
}

// ExecuteScript executes JavaScript on the client.
func (s *SSE) ExecuteScript(script string, opts ...datastar.ExecuteScriptOption) error {
	return s.ServerSentEventGenerator.ExecuteScript(script, opts...)
}

// Context returns the request context.
func (s *SSE) Context() context.Context {
	return s.ServerSentEventGenerator.Context()
}

// IsClosed returns true if the client has disconnected.
func (s *SSE) IsClosed() bool {
	return s.ServerSentEventGenerator.IsClosed()
}

// ReadSignals extracts Datastar signals from an HTTP request.
func ReadSignals(r *http.Request, signals any) error {
	return datastar.ReadSignals(r, signals)
}

// RenderTempl renders a templ component to a string.
func RenderTempl(c templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Re-export commonly used options
var (
	WithModeAppend  = datastar.WithModeAppend
	WithModePrepend = datastar.WithModePrepend
	WithSelectorID  = datastar.WithSelectorID
)
