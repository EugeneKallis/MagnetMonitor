package main

import (
	"context"
	_ "embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
	"golang.design/x/clipboard"
)

//go:embed icon.ico
var iconData []byte

func main() {
	// 1) Initialize clipboard watcher
	if err := clipboard.Init(); err != nil {
		println("clipboard.Init error:", err.Error())
		return
	}

	// 2) Start watching in the background
	go watchClipboard(context.Background())

	// 3) Launch the tray icon
	systray.Run(onReady, onExit)
}

func onReady() {
	// Set embedded icon, title, and tooltip
	systray.SetIcon(iconData)
	systray.SetTitle("MagnetMon")
	systray.SetTooltip("Magnet Clipboard Monitor")

	// Quit menu
	mQuit := systray.AddMenuItem("Quit", "Exit the app")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	// nothing to clean up
}

func watchClipboard(ctx context.Context) {
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	for data := range ch {
		txt := string(data)
		if strings.HasPrefix(txt, "magnet:") {
			saveMagnet(txt)
		}
	}
}

func saveMagnet(uri string) {
	// Show Save As dialog with default filename
	path, err := dialog.
		File().
		Filter("Magnet Files", "magnet").
		Title("Save Magnet…").
		SetStartFile("download.magnet").
		Save()
	if err != nil {
		// user cancelled or error
		return
	}

	// Ensure .magnet extension
	if filepath.Ext(path) != ".magnet" {
		path += ".magnet"
	}

	// Write file
	if err := os.WriteFile(path, []byte(uri), fs.ModePerm); err != nil {
		dialog.Message("Failed to write:\n%s", err.Error()).
			Title("Error").
			Error()
	}
}
