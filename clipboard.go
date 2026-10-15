package main

import (
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"regexp"
	"strings"
	"time"
)

var mediaPattern = regexp.MustCompile(`(?i)(youtube\.com|youtu\.be|twitch\.tv|kick\.com|tiktok\.com|\.mp4(?:\?|$))`)

func (a *App) watchClipboard() {
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()
	last := ""
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			s, _ := wruntime.ClipboardGetText(a.ctx)
			s = strings.TrimSpace(s)
			if s != "" && s != last && mediaPattern.MatchString(s) {
				last = s
				wruntime.EventsEmit(a.ctx, "clipboard-url-detected", s)
			}
		}
	}
}
