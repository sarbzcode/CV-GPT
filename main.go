package main

import (
    "embed"
    "log"

    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"

    "resume-gpt/internal/matcher"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    matcher.LoadDotEnv()
    app := NewApp()

    err := wails.Run(&options.App{
        Title:  "CV-GPT",
        Width:  1100,
        Height: 720,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        OnStartup: app.startup,
        Bind: []interface{}{
            app,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
