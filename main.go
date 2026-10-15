package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title:            "MediaHub",
		Width:            1000,
		Height:           720,
		MinWidth:         760,
		MinHeight:        560,
		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 18, A: 255},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.onBeforeClose,
		Bind:             []interface{}{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}
