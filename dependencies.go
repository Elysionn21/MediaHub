package main

import (
	"context"
	"fmt"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const ytdlpURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"

type Dependencies struct {
	dir     string
	mu      sync.RWMutex
	message string
	errText string
}

func NewDependencies(root string) *Dependencies {
	return &Dependencies{dir: filepath.Join(root, "bin"), message: "Bağımlılıklar kontrol ediliyor..."}
}
func (d *Dependencies) setStatus(message, errText string) {
	d.mu.Lock()
	d.message, d.errText = message, errText
	d.mu.Unlock()
	payload := map[string]string{"message": message}
	if errText != "" {
		payload["error"] = errText
	}
	if currentApp != nil && currentApp.ctx != nil {
		wruntime.EventsEmit(currentApp.ctx, "dependency-status", payload)
	}
}
func (d *Dependencies) Status() map[string]string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return map[string]string{"message": d.message, "error": d.errText}
}
func (d *Dependencies) path() string {
	n := "yt-dlp"
	if runtime.GOOS == "windows" {
		n += ".exe"
	}
	return filepath.Join(d.dir, n)
}
func (d *Dependencies) binary() string {
	if p, e := exec.LookPath("yt-dlp"); e == nil {
		return p
	}
	if _, e := os.Stat(d.path()); e == nil {
		return d.path()
	}
	return ""
}
func (d *Dependencies) Ensure() {
	if d.binary() == "" {
		d.setStatus("yt-dlp indiriliyor...", "")
		if e := d.download(); e != nil {
			d.setStatus("Bağımlılık indirilemedi", e.Error())
			if findVLC() == "" {
				wruntime.EventsEmit(appContext(), "vlc-missing")
			}
			return
		}
	}
	if findVLC() == "" {
		wruntime.EventsEmit(appContext(), "vlc-missing")
	}
	d.setStatus("Hazır", "")
}

// Context is supplied through the package-level app reference below.
var currentApp *App

func appContext() context.Context { return currentApp.ctx }
func (d *Dependencies) download() error {
	if e := os.MkdirAll(d.dir, 0755); e != nil {
		return e
	}
	client := http.Client{Timeout: 5 * time.Minute}
	r, e := client.Get(ytdlpURL)
	if e != nil {
		return e
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return fmt.Errorf("yt-dlp indirme hatası: %s", r.Status)
	}
	tmp := d.path() + ".tmp"
	f, e := os.Create(tmp)
	if e != nil {
		return e
	}
	_, e = io.Copy(f, r.Body)
	ce := f.Close()
	if e != nil {
		os.Remove(tmp)
		return e
	}
	if ce != nil {
		return ce
	}
	return os.Rename(tmp, d.path())
}
func (d *Dependencies) UpdateYtdlp() error {
	if p := d.binary(); p != "" {
		cmd := exec.Command(p, "-U")
		return cmd.Run()
	}
	return d.download()
}
