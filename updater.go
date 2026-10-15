package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const CurrentVersion = "v0.1.0"
const GitHubRepo = "https://api.github.com/repos/Elysionn21/MediaHub/releases/latest"

type releaseInfo struct {
	Tag    string `json:"tag_name"`
	Notes  string `json:"body"`
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func CheckForUpdates(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GitHubRepo, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "MediaHub-Updater")
	c := http.Client{Timeout: 20 * time.Second}
	r, err := c.Do(req)
	if err != nil {
		return false, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return false, fmt.Errorf("GitHub release check failed: %s", r.Status)
	}
	var rel releaseInfo
	if err = json.NewDecoder(r.Body).Decode(&rel); err != nil {
		return false, err
	}
	if rel.Tag == "" || !newerVersion(rel.Tag, CurrentVersion) {
		return false, nil
	}
	assetURL := ""
	for _, a := range rel.Assets {
		if strings.EqualFold(filepath.Ext(a.Name), ".exe") {
			assetURL = a.URL
			if strings.EqualFold(a.Name, "MediaHub.exe") {
				break
			}
		}
	}
	if assetURL == "" {
		return false, fmt.Errorf("release %s has no .exe asset", rel.Tag)
	}
	wruntime.EventsEmit(ctx, "app-update-available", map[string]string{"version": rel.Tag, "notes": rel.Notes, "downloadUrl": assetURL})
	return true, nil
}

func newerVersion(latest, current string) bool {
	parse := func(s string) ([3]int, string, bool) {
		var nums [3]int
		s = strings.TrimPrefix(strings.TrimSpace(s), "v")
		s = strings.SplitN(s, "+", 2)[0]
		parts := strings.SplitN(s, "-", 2)
		for i, p := range strings.Split(parts[0], ".") {
			if i >= 3 {
				return nums, "", false
			}
			n, e := strconv.Atoi(p)
			if e != nil || n < 0 {
				return nums, "", false
			}
			nums[i] = n
		}
		if strings.Count(parts[0], ".") != 2 {
			return nums, "", false
		}
		pre := ""
		if len(parts) == 2 {
			pre = parts[1]
		}
		return nums, pre, true
	}
	a, ap, okA := parse(latest)
	b, bp, okB := parse(current)
	if !okA || !okB {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return comparePrerelease(ap, bp) > 0
}
func comparePrerelease(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return 1
	}
	if b == "" {
		return -1
	}
	aa, bb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(aa) && i < len(bb); i++ {
		x, y := aa[i], bb[i]
		xn, xe := strconv.ParseUint(x, 10, 64)
		yn, ye := strconv.ParseUint(y, 10, 64)
		switch {
		case xe == nil && ye == nil:
			if xn < yn {
				return -1
			}
			if xn > yn {
				return 1
			}
		case xe == nil && ye != nil:
			return -1
		case xe != nil && ye == nil:
			return 1
		default:
			if x < y {
				return -1
			}
			if x > y {
				return 1
			}
		}
	}
	if len(aa) < len(bb) {
		return -1
	}
	if len(aa) > len(bb) {
		return 1
	}
	return 0
}

func ApplyUpdate(ctx context.Context, downloadURL string) error {
	u, err := url.Parse(downloadURL)
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Host, "github.com") || !strings.HasPrefix(u.Path, "/Elysionn21/MediaHub/releases/download/") || !strings.EqualFold(filepath.Ext(u.Path), ".exe") {
		return fmt.Errorf("update URL is not a MediaHub GitHub release executable")
	}
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		root, _ = os.UserConfigDir()
	}
	dir := filepath.Join(root, "MediaHub", "updates")
	if err = os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp := filepath.Join(dir, "MediaHub_new.exe")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "MediaHub-Updater")
	r, err := (&http.Client{Timeout: 30 * time.Minute}).Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("update download failed: %s", r.Status)
	}
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	pw := &updateProgressWriter{ctx: ctx, total: r.ContentLength}
	_, copyErr := io.Copy(f, io.TeeReader(r.Body, pw))
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	check, err := os.Open(tmp)
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	magic := make([]byte, 2)
	_, err = io.ReadFull(check, magic)
	_ = check.Close()
	if err != nil || string(magic) != "MZ" {
		_ = os.Remove(tmp)
		return fmt.Errorf("downloaded update is not a valid Windows executable")
	}
	current, err := os.Executable()
	if err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		return fmt.Errorf("automatic executable replacement is supported on Windows only")
	}
	batch := filepath.Join(dir, "apply_update.cmd")
	script := fmt.Sprintf("@echo off\r\ntimeout /t 2 /nobreak >NUL\r\nmove /Y %s %s >NUL\r\nif errorlevel 1 (start \"\" %s & exit /b 1)\r\nstart \"\" %s\r\ndel %%~f0\r\n", cmdQuote(tmp), cmdQuote(current), cmdQuote(current), cmdQuote(current))
	if err = os.WriteFile(batch, []byte(script), 0600); err != nil {
		return err
	}
	cmd := exec.Command("cmd.exe", "/C", batch)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x00000008 | 0x08000000}
	if err = cmd.Start(); err != nil {
		return err
	}
	go func() { time.Sleep(250 * time.Millisecond); wruntime.Quit(ctx) }()
	return nil
}

func cmdQuote(s string) string { return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"` }

type updateProgressWriter struct {
	ctx   context.Context
	total int64
	read  int64
	last  int
}

func (w *updateProgressWriter) Write(p []byte) (int, error) {
	w.read += int64(len(p))
	pct := 0
	if w.total > 0 {
		pct = int(w.read * 100 / w.total)
	}
	if pct != w.last || pct == 100 {
		w.last = pct
		wruntime.EventsEmit(w.ctx, "update-download-progress", map[string]interface{}{"percent": pct, "downloaded": w.read, "total": w.total})
	}
	return len(p), nil
}
