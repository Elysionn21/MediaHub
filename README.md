# MediaHub

MediaHub is a Windows desktop media launcher built with Wails v2, Go, and a vanilla JavaScript frontend. It can open media links in VLC, download through yt-dlp, and keep download history.

<img width="1002" height="720" alt="mediahub" src="https://github.com/user-attachments/assets/d5b73606-a921-4744-b02c-a746b1562f3d" />

## Requirements

- Windows 10 or 11, 64-bit
- Go 1.21 or later
- Wails v2 CLI. The project uses Wails v2.9.2 in go.mod; a matching CLI is recommended.
- VLC Media Player for playback
- WebView2 Runtime, normally present on current Windows installations

The app downloads yt-dlp.exe to %LOCALAPPDATA%\MediaHub\bin on first launch if yt-dlp is not already on PATH. VLC is detected from PATH and standard VideoLAN installation locations. VLC is not bundled.

## Build

From the project root in PowerShell, run:

    wails build -clean

The portable Windows executable is created at:

    build\bin\MediaHub.exe

The executable icon is build\windows\icon.ico. The app header logo is embedded from frontend\MediaHubIcon.png. Keep both assets in source control.

## Features

- Open media URLs in VLC. For supported sites such as YouTube, yt-dlp resolves stream URLs first.
- Download video or MP3 audio with yt-dlp.
- Use the Downloads folder by default; change language or download location in Settings.
- Detect supported media links copied to the clipboard and insert the detected link by clicking its notification.
- Check GitHub Releases for updates and show release notes and download progress.
- Use the system tray icon to reopen MediaHub, play the clipboard link, or quit. Closing the window hides it to the tray; minimizing keeps it in the taskbar.
- A per-user Start Menu shortcut is created automatically and uses the executable icon.
- Store preferences, yt-dlp, and history under %LOCALAPPDATA%\MediaHub.

## Current limitations

- The system tray menu is implemented. The global Alt+V hotkey is not implemented.
- VLC and yt-dlp are external programs. VLC is not bundled with the portable executable.
- Automatic updates are configured for the public repository Elysionn21/MediaHub. Each release must include a Windows .exe asset named MediaHub.exe.
- Before replacing the running executable, the updater verifies the downloaded file against the SHA-256 digest returned by the GitHub Releases API and confirms it is a Windows executable. GitHub asset digests are checked over HTTPS. The app does not use Windows code signing; publish release assets from a trusted build environment.
