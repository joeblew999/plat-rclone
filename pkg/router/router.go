// Package router provides a chi-based router with Datastar integration.
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router wraps chi.Mux with convenience methods.
type Router struct {
	*chi.Mux
}

// New creates a new Router with default middleware.
func New() *Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	return &Router{Mux: r}
}

// Handler is a function that handles requests with a Context.
type Handler func(*Context) error

// PageHandler is a function that returns HTML content.
type PageHandler func(*Context) (string, error)

// GET registers a Datastar SSE handler for GET requests.
func (r *Router) GET(pattern string, h Handler) {
	r.Mux.Get(pattern, r.wrap(h))
}

// POST registers a Datastar SSE handler for POST requests.
func (r *Router) POST(pattern string, h Handler) {
	r.Mux.Post(pattern, r.wrap(h))
}

// PUT registers a Datastar SSE handler for PUT requests.
func (r *Router) PUT(pattern string, h Handler) {
	r.Mux.Put(pattern, r.wrap(h))
}

// DELETE registers a Datastar SSE handler for DELETE requests.
func (r *Router) DELETE(pattern string, h Handler) {
	r.Mux.Delete(pattern, r.wrap(h))
}

// Page registers a handler that returns full HTML pages.
func (r *Router) Page(pattern string, h PageHandler) {
	r.Mux.Get(pattern, func(w http.ResponseWriter, req *http.Request) {
		ctx := NewContext(w, req)
		html, err := h(ctx)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.HTML(html)
	})
}

// Static serves static files from a directory.
func (r *Router) Static(pattern, dir string) {
	fs := http.FileServer(http.Dir(dir))
	r.Mux.Handle(pattern+"*", http.StripPrefix(pattern, fs))
}

func (r *Router) wrap(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := NewContext(w, req)
		if err := h(ctx); err != nil {
			ctx.Error(err)
		}
	}
}
