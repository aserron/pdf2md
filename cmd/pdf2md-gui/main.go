package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/aserron/pdf2md/pkg/pdf2md"
)

//go:embed ui/*
var uiFS embed.FS

type ConversionResponse struct {
	Markdown          string   `json:"markdown"`
	Pages             int      `json:"pages"`
	CleanedWatermarks int      `json:"cleaned_watermarks"`
	CleanedMargins    int      `json:"cleaned_margins"`
	CleanedArtifacts  int      `json:"cleaned_artifacts"`
	DetectedEncoding  string   `json:"detected_encoding"`
	ImagesCount       int      `json:"images_count"`
	Images            []string `json:"images,omitempty"`
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start listener: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	appURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	engine := pdf2md.NewEngine()

	mux := http.NewServeMux()

	// Serve embedded UI assets
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := uiFS.ReadFile("ui/index.html")
		if err != nil {
			http.Error(w, "UI template missing", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	// API Conversion endpoint
	mux.HandleFunc("/api/convert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Upload parsing error: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Missing 'file' field", http.StatusBadRequest)
			return
		}
		defer file.Close()

		tmpFile, err := os.CreateTemp("", "pdf2md-upload-*.pdf")
		if err != nil {
			http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
			return
		}
		tmpPath := tmpFile.Name()
		defer func() {
			tmpFile.Close()
			os.Remove(tmpPath)
		}()

		if _, err := io.Copy(tmpFile, file); err != nil {
			http.Error(w, "Failed to write temp upload", http.StatusInternalServerError)
			return
		}
		tmpFile.Close()

		opts := pdf2md.DefaultOptions()
		opts.CleanWatermarks = parseBool(r.FormValue("watermarks"), true)
		opts.CleanMargins = parseBool(r.FormValue("margins"), true)
		opts.MergeFormatting = parseBool(r.FormValue("merge_styles"), true)
		opts.ExtractImages = parseBool(r.FormValue("images"), true)
		opts.FormatTables = parseBool(r.FormValue("tables"), true)
		opts.PreserveStyles = parseBool(r.FormValue("styles"), true)
		opts.ExtractDiagrams = parseBool(r.FormValue("diagrams"), false)
		opts.StripRegex = r.FormValue("strip")

		if opts.ExtractImages {
			opts.ImageDir = filepath.Join(os.TempDir(), fmt.Sprintf("pdf2md-assets-%d", time.Now().UnixNano()))
			defer os.RemoveAll(opts.ImageDir)
		}

		res, err := engine.ConvertFile(r.Context(), tmpPath, opts)
		if err != nil {
			http.Error(w, "Conversion error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resp := ConversionResponse{
			Markdown:          res.Markdown,
			Pages:             res.PageCount,
			CleanedWatermarks: res.CleanedWatermarks,
			CleanedMargins:    res.CleanedMargins,
			CleanedArtifacts:  res.CleanedRows,
			DetectedEncoding:  res.DetectedEncoding,
			ImagesCount:       len(res.ImagePaths),
			Images:            res.ImagePaths,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	fmt.Printf("pdf2md Windows GUI running at: %s\n", appURL)

	go launchDesktopWindow(appURL)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\nShutting down pdf2md Windows GUI...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func parseBool(val string, def bool) bool {
	if val == "" {
		return def
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return def
	}
	return b
}

func launchDesktopWindow(url string) {
	time.Sleep(100 * time.Millisecond)

	if runtime.GOOS == "windows" {
		edgePaths := []string{
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			`msedge.exe`,
		}

		for _, p := range edgePaths {
			cmd := exec.Command(p, fmt.Sprintf("--app=%s", url), "--window-size=1200,800")
			if err := cmd.Start(); err == nil {
				return
			}
		}

		exec.Command("cmd", "/c", "start", url).Start()
	} else {
		exec.Command("xdg-open", url).Start()
	}
}
