package main

import (
	_ "embed"
	"runtime"
	"strings"

	"github.com/getlantern/systray"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/windows/icon.ico
var mediaHubTrayIcon []byte

func (a *App) startTray() {
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		systray.Run(func() {
			systray.SetIcon(mediaHubTrayIcon)
			systray.SetTooltip("MediaHub")

			openLabel, playLabel, quitLabel := "Open MediaHub", "Play clipboard link", "Quit"
			if strings.EqualFold(a.settings.Language, "tr") {
				openLabel, playLabel, quitLabel = "MediaHub\u0027yi A\u00e7", "Panodaki Linki \u00c7al", "\u00c7\u0131k\u0131\u015f"
			}
			openItem := systray.AddMenuItem(openLabel, "Show the MediaHub window")
			playItem := systray.AddMenuItem(playLabel, "Play the media URL currently on the clipboard")
			systray.AddSeparator()
			quitItem := systray.AddMenuItem(quitLabel, "Exit MediaHub")

			go func() {
				for {
					select {
					case <-openItem.ClickedCh:
						wruntime.WindowShow(a.ctx)
						wruntime.WindowUnminimise(a.ctx)
					case <-playItem.ClickedCh:
						a.playClipboardFromTray()
					case <-quitItem.ClickedCh:
						wruntime.Quit(a.ctx)
						systray.Quit()
						return
					case <-a.ctx.Done():
						systray.Quit()
						return
					}
				}
			}()
		}, func() {})
	}()
}

func (a *App) playClipboardFromTray() {
	mediaURL, _ := wruntime.ClipboardGetText(a.ctx)
	mediaURL = strings.TrimSpace(mediaURL)
	if mediaURL == "" {
		a.showTrayMessage("Clipboard is empty")
		return
	}
	if err := PlayMedia(mediaURL, false); err != nil {
		a.showTrayMessage(err.Error())
	}
}

func (a *App) showTrayMessage(message string) {
	wruntime.WindowShow(a.ctx)
	wruntime.WindowUnminimise(a.ctx)
	wruntime.EventsEmit(a.ctx, "tray-message", message)
}

func systrayQuit() { systray.Quit() }
