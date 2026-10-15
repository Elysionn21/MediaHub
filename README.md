# MediaHub

MediaHub is a Windows desktop media launcher built with Wails v2, Go, and a vanilla JavaScript frontend. It can open media links in VLC, download through yt-dlp, and keep download history.

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
- Store preferences, yt-dlp, and history under %LOCALAPPDATA%\MediaHub.

## Current limitations

- System tray and global Alt+V hotkey integrations are placeholders and are not implemented.
- VLC and yt-dlp are external programs. VLC is not bundled with the portable executable.
- Automatic updates are configured for the public repository Elysionn21/MediaHub. Each release must include a Windows .exe asset named MediaHub.exe.
- Before replacing the running executable, the updater verifies the downloaded file against the SHA-256 digest returned by the GitHub Releases API and confirms it is a Windows executable. GitHub asset digests are checked over HTTPS. The app does not use Windows code signing; publish release assets from a trusted build environment.

## Publish a release

1. Push the source branch to GitHub.
2. Build from the repository root with wails build -clean.
3. On GitHub, choose Releases, then Create a new release.
4. Create tag v0.1.0 for the first release, matching CurrentVersion in updater.go. Target the main branch.
5. Upload build\bin\MediaHub.exe. Keep its filename exactly MediaHub.exe, then publish the release.
6. Make sure the repository and release are public. The updater uses the unauthenticated GitHub Releases API.
7. Confirm GitHub shows an SHA-256 digest for the uploaded asset. The updater will reject a release asset if that digest is absent or invalid.

For each later release, update CurrentVersion in updater.go and the displayed version in frontend\index.html, rebuild, and publish the executable under the matching tag. For example, after v0.1.0, build and publish v0.1.1.

## Test the updater

A release with the same version as the installed app does not trigger an update. To test the complete update flow:

1. Publish the v0.1.0 release and keep a copy of its executable.
2. Change CurrentVersion and the displayed frontend version to v0.1.1, then build a new executable.
3. Publish the new executable as the MediaHub.exe asset of release v0.1.1.
4. Launch the saved v0.1.0 executable, open Settings, and choose Check for updates.
5. Select Update now. The app should download and verify the asset, replace itself, and restart as v0.1.1.

Publishing the main branch publishes source code only. A GitHub Release with a version tag and executable asset must be created separately.