# MediaHub

Portable Wails v2 desktop media launcher. Build with Go 1.21+ and the Wails CLI (`wails build`). The current scaffold bundles the UI and downloads yt-dlp to `%LOCALAPPDATA%\MediaHub\bin` when needed. VLC is detected from PATH or standard install locations.

Before release, set the GitHub owner/repository in `updater.go`, add a signed update package strategy, and complete native Windows tray/global hotkey integrations. `ApplyAppUpdate` deliberately refuses an unsafe executable swap; the GitHub asset and integrity/signature verification policy must be configured before enabling it.
