package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func ensureStartMenuShortcut() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData, err = os.UserConfigDir()
		if err != nil {
			return err
		}
	}
	programs := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs")
	if err := os.MkdirAll(programs, 0755); err != nil {
		return err
	}
	shortcutPath := filepath.Join(programs, "MediaHub.lnk")

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer unknown.Release()
	shell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer shell.Release()

	link, err := oleutil.CallMethod(shell, "CreateShortcut", shortcutPath)
	if err != nil {
		return err
	}
	defer link.Clear()
	shortcut := link.ToIDispatch()

	setProperty := func(name string, value interface{}) error {
		result, err := oleutil.PutProperty(shortcut, name, value)
		if result != nil {
			_ = result.Clear()
		}
		return err
	}
	for name, value := range map[string]interface{}{
		"TargetPath":       exe,
		"WorkingDirectory": filepath.Dir(exe),
		"IconLocation":     fmt.Sprintf("%s,0", exe),
		"Description":      "MediaHub",
	} {
		if err := setProperty(name, value); err != nil {
			return err
		}
	}
	result, err := oleutil.CallMethod(shortcut, "Save")
	if result != nil {
		_ = result.Clear()
	}
	return err
}
