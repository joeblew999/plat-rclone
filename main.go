// plat-rclone - Cross-platform GUI for rclone
// Build with: goup-util build macos|ios|android .
package main

import (
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/gioui-plugins/gio-plugins/plugin/gioplugins"
	"github.com/gioui-plugins/gio-plugins/webviewer/giowebview"
	"github.com/gioui-plugins/gio-plugins/webviewer/webview"

	"github.com/joeblew999/plat-rclone/pkg/datastar"
	"github.com/joeblew999/plat-rclone/pkg/rclone"
	"github.com/joeblew999/plat-rclone/pkg/router"
	"github.com/joeblew999/plat-rclone/templates"
)

//go:embed static/*
var staticFS embed.FS

var th = material.NewTheme()

func main() {
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	webview.SetDebug(os.Getenv("DEBUG") == "1")

	// Start embedded HTTP server
	serverURL := startWebServer()
	fmt.Printf("Server started at %s\n", serverURL)

	// Launch Gio app with webview
	go runApp(serverURL)
	app.Main()
}

func runApp(serverURL string) {
	window := &app.Window{}
	window.Option(app.Title("plat-rclone"))
	window.Option(app.Size(1200, 800))

	var ops op.Ops
	webviewTag := new(int)
	navigated := false
	frameCount := 0

	window.Invalidate()

	for {
		evt := gioplugins.Hijack(window)

		switch evt := evt.(type) {
		case app.DestroyEvent:
			os.Exit(0)
			return

		case app.FrameEvent:
			gtx := app.NewContext(&ops, evt)

			// Process webview events
			for {
				ev, ok := gioplugins.Event(gtx, giowebview.Filter{Target: webviewTag})
				if !ok {
					break
				}
				switch e := ev.(type) {
				case giowebview.NavigationEvent:
					fmt.Printf("Navigation: %s\n", e.URL)
				case giowebview.TitleEvent:
					fmt.Printf("Title: %s\n", e.Title)
				}
			}

			// Render WebView - fills entire window
			webviewStack := giowebview.WebViewOp{Tag: webviewTag}.Push(gtx.Ops)
			giowebview.OffsetOp{Point: f32.Point{X: 0, Y: 0}}.Add(gtx.Ops)
			giowebview.RectOp{
				Size: f32.Point{
					X: float32(gtx.Constraints.Max.X),
					Y: float32(gtx.Constraints.Max.Y),
				},
			}.Add(gtx.Ops)
			webviewStack.Pop(gtx.Ops)

			// Navigate after frames to ensure webview is ready
			frameCount++
			if !navigated && frameCount > 10 {
				fmt.Printf("Navigating to: %s\n", serverURL)
				gioplugins.Execute(gtx, giowebview.NavigateCmd{
					View: webviewTag,
					URL:  serverURL,
				})
				navigated = true
			}

			// Request frames until navigated
			if !navigated {
				gtx.Execute(op.InvalidateCmd{})
			}

			evt.Frame(gtx.Ops)
		}
	}
}

func startWebServer() string {
	rcloneURL := os.Getenv("RCLONE_URL")
	if rcloneURL == "" {
		rcloneURL = "http://localhost:5572"
	}

	rc := rclone.NewClient(rcloneURL)

	// Find available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	r := setupRoutes(rc)
	serverAddr := fmt.Sprintf("127.0.0.1:%d", port)

	go func() {
		if err := http.ListenAndServe(serverAddr, r.Mux); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	return fmt.Sprintf("http://%s", serverAddr)
}

func setupRoutes(rc *rclone.Client) *router.Router {
	r := router.New()

	// Embedded static files
	r.Mux.Get("/static/css/style.css", func(w http.ResponseWriter, req *http.Request) {
		data, _ := staticFS.ReadFile("static/css/style.css")
		w.Header().Set("Content-Type", "text/css")
		w.Write(data)
	})
	r.Mux.Get("/static/js/datastar.js", func(w http.ResponseWriter, req *http.Request) {
		data, _ := staticFS.ReadFile("static/js/datastar.js")
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(data)
	})

	// Pages
	r.Page("/", func(ctx *router.Context) (string, error) {
		remotes, err := getRemotesInfo(rc)
		if err != nil {
			remotes = []templates.RemoteInfo{}
		}
		return datastar.RenderTempl(templates.RemotesPage(remotes))
	})

	r.Page("/jobs", func(ctx *router.Context) (string, error) {
		jobs, _ := getJobsInfo(rc)
		return datastar.RenderTempl(templates.JobsPage(jobs))
	})

	r.Page("/stats", func(ctx *router.Context) (string, error) {
		stats, version := getStatsInfo(rc)
		return datastar.RenderTempl(templates.StatsPage(stats, version))
	})

	// API routes
	r.GET("/api/remotes/refresh", func(ctx *router.Context) error {
		sse := ctx.SSE()
		remotes, err := getRemotesInfo(rc)
		if err != nil {
			return sse.PatchHTMLByID("remotes-list", `<div class="error">`+err.Error()+`</div>`)
		}
		return sse.PatchTemplByID("remotes-list", templates.RemotesList(remotes))
	})

	r.GET("/api/remotes/{name}/browse", func(ctx *router.Context) error {
		sse := ctx.SSE()
		name := ctx.Param("name")
		path := ctx.Query("path")
		if path == ".." {
			path = ""
		}

		items, err := rc.List(name, path)
		if err != nil {
			return sse.PatchHTMLByID("file-browser", `<div class="error">`+err.Error()+`</div>`)
		}

		fileItems := make([]templates.FileItem, len(items))
		for i, item := range items {
			fileItems[i] = templates.FileItem{
				Name:    item.Name,
				Size:    formatSize(item.Size),
				ModTime: item.ModTime,
				IsDir:   item.IsDir,
			}
		}
		return sse.PatchTempl(templates.FileBrowser(name, path, fileItems))
	})

	r.DELETE("/api/remotes/{name}", func(ctx *router.Context) error {
		sse := ctx.SSE()
		name := ctx.Param("name")
		if err := rc.DeleteRemote(name); err != nil {
			return sse.PatchHTMLByID("remotes-list", `<div class="error">`+err.Error()+`</div>`)
		}
		return sse.RemoveByID("remote-" + name)
	})

	// Jobs API
	r.GET("/api/jobs/refresh", func(ctx *router.Context) error {
		sse := ctx.SSE()
		jobs, err := getJobsInfo(rc)
		if err != nil {
			return sse.PatchHTMLByID("jobs-list", `<div class="error">`+err.Error()+`</div>`)
		}
		return sse.PatchTemplByID("jobs-list", templates.JobsList(jobs))
	})

	r.POST("/api/jobs/{id}/stop", func(ctx *router.Context) error {
		sse := ctx.SSE()
		id := ctx.Param("id")
		var jobID int64
		fmt.Sscanf(id, "%d", &jobID)
		if err := rc.StopJob(jobID); err != nil {
			return sse.PatchHTMLByID(fmt.Sprintf("job-%d", jobID), `<div class="error">`+err.Error()+`</div>`)
		}
		// Refresh jobs list after stopping
		jobs, _ := getJobsInfo(rc)
		return sse.PatchTemplByID("jobs-list", templates.JobsList(jobs))
	})

	// Stats API
	r.GET("/api/stats/refresh", func(ctx *router.Context) error {
		sse := ctx.SSE()
		stats, version := getStatsInfo(rc)
		return sse.PatchTemplByID("stats-content", templates.StatsContent(stats, version))
	})

	return r
}

func getRemotesInfo(rc *rclone.Client) ([]templates.RemoteInfo, error) {
	names, err := rc.ListRemotes()
	if err != nil {
		return nil, err
	}

	remotes := make([]templates.RemoteInfo, len(names))
	for i, name := range names {
		config, _ := rc.GetRemote(name)
		remoteType := "unknown"
		if t, ok := config["type"]; ok {
			remoteType = t
		}
		remotes[i] = templates.RemoteInfo{Name: name, Type: remoteType}
	}
	return remotes, nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func getJobsInfo(rc *rclone.Client) ([]templates.JobInfo, error) {
	jobs, err := rc.ListJobs()
	if err != nil {
		return nil, err
	}

	result := make([]templates.JobInfo, len(jobs))
	for i, job := range jobs {
		status := "running"
		if job.Finished {
			if job.Success {
				status = "finished"
			} else {
				status = "error"
			}
		}
		result[i] = templates.JobInfo{
			ID:        job.ID,
			Group:     job.Group,
			StartTime: job.StartTime,
			Status:    status,
			Error:     job.Error,
		}
	}
	return result, nil
}

func getStatsInfo(rc *rclone.Client) (templates.StatsInfo, templates.VersionInfo) {
	stats := templates.StatsInfo{
		Bytes:       "0 B",
		Speed:       "0 B/s",
		Eta:         "-",
		ElapsedTime: "0s",
	}
	version := templates.VersionInfo{
		Version:   "unknown",
		GoVersion: "unknown",
		Os:        "unknown",
		Arch:      "unknown",
	}

	// Get stats
	if s, err := rc.Stats(); err == nil {
		if b, ok := s["bytes"].(float64); ok {
			stats.Bytes = formatSize(int64(b))
		}
		if sp, ok := s["speed"].(float64); ok {
			stats.Speed = formatSize(int64(sp)) + "/s"
		}
		if eta, ok := s["eta"].(float64); ok {
			stats.Eta = fmt.Sprintf("%.0fs", eta)
		}
		if elapsed, ok := s["elapsedTime"].(float64); ok {
			stats.ElapsedTime = fmt.Sprintf("%.1fs", elapsed)
		}
		if t, ok := s["transfers"].(float64); ok {
			stats.Transfers = int64(t)
		}
		if tt, ok := s["totalTransfers"].(float64); ok {
			stats.TotalTransfers = int64(tt)
		}
		if c, ok := s["checks"].(float64); ok {
			stats.Checks = int64(c)
		}
		if tc, ok := s["totalChecks"].(float64); ok {
			stats.TotalChecks = int64(tc)
		}
		if e, ok := s["errors"].(float64); ok {
			stats.Errors = int64(e)
		}
		if d, ok := s["deletes"].(float64); ok {
			stats.Deletes = int64(d)
		}
	}

	// Get version
	if v, err := rc.Version(); err == nil {
		if ver, ok := v["version"].(string); ok {
			version.Version = ver
		}
		if gv, ok := v["goVersion"].(string); ok {
			version.GoVersion = gv
		}
		if os, ok := v["os"].(string); ok {
			version.Os = os
		}
		if arch, ok := v["arch"].(string); ok {
			version.Arch = arch
		}
	}

	return stats, version
}
