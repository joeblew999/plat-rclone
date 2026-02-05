package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joeblew999/plat-rclone/pkg/datastar"
	"github.com/joeblew999/plat-rclone/pkg/rclone"
	"github.com/joeblew999/plat-rclone/pkg/router"
	"github.com/joeblew999/plat-rclone/templates"
)

var (
	addr       = ":8080"
	rcloneURL  = "http://localhost:5572"
	rcloneUser = ""
	rclonePass = ""
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "download":
			cmdDownload()
			return
		case "serve":
			parseFlags()
			cmdServe()
			return
		case "help", "-h", "--help":
			printUsage()
			return
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}

	// Default: serve
	parseFlags()
	cmdServe()
}

func printUsage() {
	fmt.Println(`plat-rclone - GUI for rclone

Commands:
  serve      Start the web server (default)
  download   Download rclone binary from GitHub releases

Options for serve:
  -addr      HTTP server address (default ":8080")
  -rclone    rclone RC API URL (default "http://localhost:5572")
  -user      rclone RC username
  -pass      rclone RC password

Examples:
  plat-rclone                    # Start web server on :8080
  plat-rclone serve -addr :3000  # Start on custom port
  plat-rclone download           # Download rclone to current directory
  plat-rclone download ./bin     # Download to ./bin directory`)
}

func parseFlags() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "serve" {
		args = args[1:]
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-addr":
			if i+1 < len(args) {
				addr = args[i+1]
				i++
			}
		case "-rclone":
			if i+1 < len(args) {
				rcloneURL = args[i+1]
				i++
			}
		case "-user":
			if i+1 < len(args) {
				rcloneUser = args[i+1]
				i++
			}
		case "-pass":
			if i+1 < len(args) {
				rclonePass = args[i+1]
				i++
			}
		}
	}
}

func cmdDownload() {
	dir := "."
	if len(os.Args) > 2 {
		dir = os.Args[2]
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	path, err := rclone.Download(rclone.DownloadOptions{
		Version: "current",
		OS:      getOS(),
		Arch:    getArch(),
		Dir:     dir,
	})
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	fmt.Printf("rclone installed to: %s\n", path)
	fmt.Println("\nTo start rclone RC server:")
	fmt.Printf("  %s rcd --rc-no-auth\n", path)
}

func getOS() string {
	switch os := os.Getenv("GOOS"); os {
	case "":
		// Get from runtime
		switch {
		case fileExists("/System/Library/CoreServices/SystemVersion.plist"):
			return "osx"
		case fileExists("/etc/os-release"):
			return "linux"
		default:
			return "osx" // Default to macOS
		}
	case "darwin":
		return "osx"
	default:
		return os
	}
}

func getArch() string {
	switch arch := os.Getenv("GOARCH"); arch {
	case "":
		// Detect from uname
		return "arm64" // Default for Apple Silicon
	default:
		return arch
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func cmdServe() {
	// Create rclone client
	rc := rclone.NewClient(rcloneURL)
	if rcloneUser != "" {
		rc.WithAuth(rcloneUser, rclonePass)
	}

	// Create router
	r := router.New()

	// Static files
	r.Static("/static/", "static")

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

	// API: Refresh remotes list
	r.GET("/api/remotes/refresh", func(ctx *router.Context) error {
		sse := ctx.SSE()
		remotes, err := getRemotesInfo(rc)
		if err != nil {
			return sse.PatchHTMLByID("remotes-list", `<div class="error">`+err.Error()+`</div>`)
		}
		return sse.PatchTemplByID("remotes-list", templates.RemotesList(remotes))
	})

	// API: Browse remote
	r.GET("/api/remotes/{name}/browse", func(ctx *router.Context) error {
		sse := ctx.SSE()
		name := ctx.Param("name")
		path := ctx.Query("path")

		// Handle parent directory
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

	// API: Delete remote
	r.DELETE("/api/remotes/{name}", func(ctx *router.Context) error {
		sse := ctx.SSE()
		name := ctx.Param("name")

		if err := rc.DeleteRemote(name); err != nil {
			return sse.PatchHTMLByID("remotes-list", `<div class="error">`+err.Error()+`</div>`)
		}

		// Remove card from DOM
		return sse.RemoveByID("remote-" + name)
	})

	// API: Delete file
	r.DELETE("/api/files/{remote}", func(ctx *router.Context) error {
		sse := ctx.SSE()
		remote := ctx.Param("remote")
		path := ctx.Query("path")

		if err := rc.Delete(remote, path); err != nil {
			return sse.ExecuteScript(`alert("Error: ` + err.Error() + `")`)
		}

		// Refresh the file browser
		return sse.Redirect("/api/remotes/" + remote + "/browse")
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
		jobs, _ := getJobsInfo(rc)
		return sse.PatchTemplByID("jobs-list", templates.JobsList(jobs))
	})

	// Stats API
	r.GET("/api/stats/refresh", func(ctx *router.Context) error {
		sse := ctx.SSE()
		stats, version := getStatsInfo(rc)
		return sse.PatchTemplByID("stats-content", templates.StatsContent(stats, version))
	})

	fmt.Printf("plat-rclone starting on %s\n", addr)
	fmt.Printf("rclone RC API: %s\n", rcloneURL)
	fmt.Println("Open http://localhost" + addr)
	log.Fatal(http.ListenAndServe(addr, r.Mux))
}

func getRemotesInfo(rc *rclone.Client) ([]templates.RemoteInfo, error) {
	names, err := rc.ListRemotes()
	if err != nil {
		return nil, err
	}

	remotes := make([]templates.RemoteInfo, len(names))
	for i, name := range names {
		config, err := rc.GetRemote(name)
		remoteType := "unknown"
		if err == nil {
			if t, ok := config["type"]; ok {
				remoteType = t
			}
		}
		remotes[i] = templates.RemoteInfo{
			Name: name,
			Type: remoteType,
		}
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
