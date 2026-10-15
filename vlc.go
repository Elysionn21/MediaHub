package main

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func findVLC() string {
	names := []string{"vlc"}
	if runtime.GOOS == "windows" {
		names = []string{"vlc.exe", `C:\Program Files\VideoLAN\VLC\vlc.exe`, `C:\Program Files (x86)\VideoLAN\VLC\vlc.exe`}
	}
	for _, n := range names {
		if p, e := exec.LookPath(n); e == nil {
			return p
		}
		if strings.Contains(n, `:\`) {
			if _, e := exec.LookPath(n); e == nil {
				return n
			}
		}
	}
	return ""
}
func PlayMedia(mediaURL string, enqueue bool) error {
	mediaURL = strings.TrimSpace(mediaURL)
	if mediaURL == "" {
		return fmt.Errorf("URL boş")
	}
	bin := findVLC()
	if bin == "" {
		return fmt.Errorf("VLC bulunamadı")
	}
	playURL := mediaURL
	audioURL := ""
	playbackWarning := ""
	if u, err := url.Parse(mediaURL); err == nil && u.Scheme != "file" && !strings.Contains(strings.ToLower(u.Path), ".mp4") && !strings.Contains(strings.ToLower(u.Path), ".m3u8") && !strings.Contains(strings.ToLower(u.Path), ".mpd") {
		ytdlp := execYtdlp()
		if ytdlp == "" {
			playbackWarning = "yt-dlp is not available; VLC will try the original link"
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			out, err := exec.CommandContext(ctx, ytdlp, "--no-config", "--quiet", "--no-warnings", "--no-playlist", "-f", "bestvideo+bestaudio/best", "-g", mediaURL).CombinedOutput()
			cancel()
			streams := []string{}
			if err == nil {
				for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
					candidate := strings.TrimSpace(line)
					if parsed, parseErr := url.Parse(candidate); parseErr == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") {
						streams = append(streams, candidate)
					}
				}
			}
			if len(streams) > 0 {
				playURL = streams[0]
				if len(streams) > 1 {
					audioURL = streams[1]
				}
			} else {
				playbackWarning = "yt-dlp could not resolve the YouTube streams"
				if err != nil && strings.TrimSpace(string(out)) != "" {
					detail := strings.TrimSpace(string(out))
					if len(detail) > 240 {
						detail = detail[:240] + "…"
					}
					playbackWarning += ": " + detail
				}
				playbackWarning += "; VLC will try the original link"
			}
		}
	}
	args := []string{"--no-one-instance", "--no-video-title-show", "--http-user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"}
	if u, err := url.Parse(mediaURL); err == nil && u.Scheme != "" && u.Host != "" {
		args = append(args, "--http-referrer="+mediaURL)
	}
	if audioURL != "" {
		args = append(args, "--input-slave="+audioURL)
	}
	if enqueue {
		args = append(args, "--playlist-enqueue")
	}
	args = append(args, playURL)
	cmd := exec.Command(bin, args...)
	if e := cmd.Start(); e != nil {
		return fmt.Errorf("could not start VLC: %w", e)
	}
	go cmd.Wait()
	if playbackWarning != "" {
		return fmt.Errorf("%s", playbackWarning)
	}
	return nil
}
