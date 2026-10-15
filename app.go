package main

import (
	"context"
	"encoding/json"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

type App struct {
	ctx       context.Context
	dataDir   string
	deps      *Dependencies
	downloads *Downloader
	settings  Preferences
	quitting  atomic.Bool
}

type Preferences struct {
	Language    string `json:"language"`
	DownloadDir string `json:"downloadDir"`
}

func NewApp() *App { return &App{} }
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	currentApp = a
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		root, _ = os.UserConfigDir()
	}
	a.dataDir = filepath.Join(root, "MediaHub")
	_ = os.MkdirAll(filepath.Join(a.dataDir, "bin"), 0755)
	a.settings = a.loadPreferences()
	a.deps = NewDependencies(a.dataDir)
	a.downloads = NewDownloader(ctx, a.dataDir)
	a.downloads.SetDirectory(a.settings.DownloadDir)
	go a.deps.Ensure()
	a.startTray()
	go func() {
		if err := ensureStartMenuShortcut(); err != nil {
			log.Printf("create Start Menu shortcut: %v", err)
		}
	}()
	go a.watchClipboard()
	go func() { _, _ = CheckForUpdates(ctx) }()
}
func (a *App) shutdown(context.Context)                 { systrayQuit() }
func (a *App) PlayMedia(url string, enqueue bool) error { return PlayMedia(url, enqueue) }
func (a *App) DownloadMedia(url, format string) error   { return a.downloads.DownloadMedia(url, format) }
func (a *App) History() []HistoryItem                   { return a.downloads.History() }
func (a *App) DependencyStatus() map[string]string      { return a.deps.Status() }
func (a *App) GetSettings() Preferences                 { return a.settings }
func (a *App) SaveSettings(p Preferences) error {
	if p.Language != "en" && p.Language != "tr" {
		return os.ErrInvalid
	}
	p.DownloadDir = strings.TrimSpace(p.DownloadDir)
	if p.DownloadDir == "" || !filepath.IsAbs(p.DownloadDir) {
		return os.ErrInvalid
	}
	if err := os.MkdirAll(p.DownloadDir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(a.dataDir, "settings.json"), b, 0600); err != nil {
		return err
	}
	a.settings = p
	a.downloads.SetDirectory(p.DownloadDir)
	return nil
}
func (a *App) SelectDownloadFolder() (string, error) {
	return wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{Title: "Choose download folder", DefaultDirectory: a.settings.DownloadDir})
}
func (a *App) loadPreferences() Preferences {
	home, _ := os.UserHomeDir()
	p := Preferences{Language: "en", DownloadDir: filepath.Join(home, "Downloads")}
	b, err := os.ReadFile(filepath.Join(a.dataDir, "settings.json"))
	if err == nil {
		var saved Preferences
		if json.Unmarshal(b, &saved) == nil {
			if saved.Language == "en" || saved.Language == "tr" {
				p.Language = saved.Language
			}
			if filepath.IsAbs(saved.DownloadDir) {
				p.DownloadDir = saved.DownloadDir
			}
		}
	}
	_ = os.MkdirAll(p.DownloadDir, 0755)
	return p
}
func (a *App) OpenFolder(path string) error         { return OpenFolder(path) }
func (a *App) UpdateYtdlp() error                   { return a.deps.UpdateYtdlp() }
func (a *App) CheckForUpdates() (bool, error)       { return CheckForUpdates(a.ctx) }
func (a *App) ApplyUpdate(url, sha256 string) error { return ApplyUpdate(a.ctx, url, sha256) }
func (a *App) Minimise()                            { wruntime.WindowHide(a.ctx) }
func (a *App) Close()                               { wruntime.WindowHide(a.ctx) }
func (a *App) quit() {
	a.quitting.Store(true)
	wruntime.Quit(a.ctx)
}
func (a *App) onBeforeClose(ctx context.Context) bool {
	if a.quitting.Load() {
		return false
	}
	wruntime.WindowHide(ctx)
	return true
}
func (a *App) ClipboardText() string { s, _ := wruntime.ClipboardGetText(a.ctx); return s }
