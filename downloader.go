package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type HistoryItem struct {
	Title    string `json:"title"`
	Path     string `json:"path"`
	URL      string `json:"url"`
	Finished string `json:"finished"`
}
type Progress struct {
	ID      string  `json:"id"`
	Line    string  `json:"line"`
	Percent float64 `json:"percent"`
	Speed   string  `json:"speed"`
	ETA     string  `json:"eta"`
	Done    bool    `json:"done"`
	Error   string  `json:"error,omitempty"`
}
type Downloader struct {
	ctx    context.Context
	root   string
	outDir string
	mu     sync.Mutex
	active map[string]bool
}

func NewDownloader(ctx context.Context, root string) *Downloader {
	return &Downloader{ctx: ctx, root: root, outDir: filepath.Join(mustUserHome(), "Downloads"), active: map[string]bool{}}
}
func (d *Downloader) SetDirectory(dir string) { d.mu.Lock(); d.outDir = dir; d.mu.Unlock() }
func (d *Downloader) DownloadMedia(url, format string) error {
	url = strings.TrimSpace(url)
	if !validMediaURL(url) {
		return fmt.Errorf("geçerli bir medya URL'si girin")
	}
	bin := execYtdlp()
	if bin == "" {
		return fmt.Errorf("yt-dlp bulunamadı; bağımlılık hazırlanıyor")
	}
	id := url
	d.mu.Lock()
	if d.active[id] {
		d.mu.Unlock()
		return fmt.Errorf("bu indirme zaten etkin")
	}
	d.active[id] = true
	d.mu.Unlock()
	go d.run(bin, url, format, id)
	return nil
}
func (d *Downloader) run(bin, url, format, id string) {
	defer func() { d.mu.Lock(); delete(d.active, id); d.mu.Unlock() }()
	d.mu.Lock()
	outDir := d.outDir
	d.mu.Unlock()
	args := []string{"--newline", "--no-playlist", "-o", filepath.Join(outDir, "%(title)s.%(ext)s")}
	if format == "audio" {
		args = append(args, "-x", "--audio-format", "mp3")
	}
	args = append(args, url)
	_ = os.MkdirAll(outDir, 0755)
	cmd := exec.CommandContext(d.ctx, bin, args...)
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		emit(d.ctx, "download-progress", Progress{ID: id, Error: e.Error()})
		return
	}
	cmd.Stderr = cmd.Stdout
	if e = cmd.Start(); e != nil {
		emit(d.ctx, "download-progress", Progress{ID: id, Error: e.Error()})
		return
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	re := regexp.MustCompile(`\[download\]\s+([\d.]+)%.*?at\s+(.+?)\s+ETA\s+(.+)`)
	for scanner.Scan() {
		line := scanner.Text()
		p := Progress{ID: id, Line: line}
		m := re.FindStringSubmatch(line)
		if len(m) > 0 {
			p.Percent, _ = strconv.ParseFloat(m[1], 64)
			p.Speed = m[2]
			p.ETA = m[3]
		}
		emit(d.ctx, "download-progress", p)
	}
	e = cmd.Wait()
	if e != nil {
		emit(d.ctx, "download-progress", Progress{ID: id, Error: e.Error()})
		return
	}
	item := HistoryItem{Title: url, Path: outDir, URL: url}
	d.appendHistory(item)
	emit(d.ctx, "download-progress", Progress{ID: id, Percent: 100, Done: true})
}
func (d *Downloader) History() []HistoryItem {
	f, e := os.Open(filepath.Join(d.root, "history.json"))
	if e != nil {
		return []HistoryItem{}
	}
	defer f.Close()
	var out []HistoryItem
	_ = json.NewDecoder(f).Decode(&out)
	return out
}
func (d *Downloader) appendHistory(item HistoryItem) {
	items := d.History()
	items = append([]HistoryItem{item}, items...)
	b, _ := json.MarshalIndent(items, "", "  ")
	_ = os.WriteFile(filepath.Join(d.root, "history.json"), b, 0600)
}
func execYtdlp() string {
	if p, e := exec.LookPath("yt-dlp"); e == nil {
		return p
	}
	root := os.Getenv("LOCALAPPDATA")
	if root != "" {
		p := filepath.Join(root, "MediaHub", "bin", "yt-dlp.exe")
		if _, e := os.Stat(p); e == nil {
			return p
		}
	}
	return ""
}
func mustUserHome() string {
	h, e := os.UserHomeDir()
	if e != nil {
		return "."
	}
	return h
}
func validMediaURL(s string) bool {
	return strings.HasPrefix(strings.ToLower(s), "http://") || strings.HasPrefix(strings.ToLower(s), "https://")
}
func OpenFolder(path string) error {
	if runtime.GOOS == "windows" {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return exec.Command("explorer.exe", path).Start()
		}
		return exec.Command("explorer.exe", "/select,", path).Start()
	}
	return exec.Command("xdg-open", filepath.Dir(path)).Start()
}
func emit(ctx context.Context, name string, v interface{}) { wruntime.EventsEmit(ctx, name, v) }
