// main.go
package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.design/x/clipboard"
)

func main() {
	// Initialize the clipboard watcher
	if err := clipboard.Init(); err != nil {
		walk.MsgBox(nil, "Error", "clipboard.Init failed: "+err.Error(), walk.MsgBoxIconError)
		return
	}

	// Build the hidden main window + tray icon
	var ni *walk.NotifyIcon
	MainWindow{
		AssignTo: &ni,
		Visible:  false,
		Stager:   NewTrayIcon(),
		Title:    "Magnet Watcher",
		OnInitialize: func() {
			ni.SetToolTip("Magnet Clipboard Watcher")
		},
		MenuItems: []MenuItem{
			{Text: "E&xit", OnTriggered: func() { ni.Dispose(); os.Exit(0) }},
		},
	}.Run()

	// Launch the watcher in the background
	go watchLoop()

	// Run the tray icon message loop
	walk.RunMainWindow()
}

// watchLoop watches for clipboard changes
func watchLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := clipboard.Watch(ctx, clipboard.FmtText)
	for data := range ch {
		txt := string(data)
		if strings.HasPrefix(txt, "magnet:") {
			// on a goroutine so we don’t block the watcher
			go handleMagnet(txt)
		}
	}
}

// handleMagnet shows a dialog to choose folder, then writes the file
func handleMagnet(uri string) {
	// run on main GUI thread
	walk.Synchronize(func() {
		dlg := newDialog()
		choice := dlg.Run()

		if choice == "" {
			return // user closed dialog or cancelled
		}

		dir := filepath.Join(`C:\Users\ponzi`, choice)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			walk.MsgBox(nil, "Error", "Could not create folder: "+err.Error(), walk.MsgBoxIconError)
			return
		}

		// find a unique filename
		base := filepath.Join(dir, "download.magnet")
		path := uniquePath(base)

		if err := os.WriteFile(path, []byte(uri), fs.ModePerm); err != nil {
			walk.MsgBox(nil, "Error", "Could not write file: "+err.Error(), walk.MsgBoxIconError)
			return
		}
	})
}

// uniquePath returns a non-existent filename by appending 1,2,3…
func uniquePath(base string) string {
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return base
	}
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s%d%s", name, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// newDialog builds a simple modal with two buttons
func newDialog() *choiceDialog {
	dlg := &choiceDialog{}
	Dialog{
		AssignTo: &dlg.Dialog,
		Title:    "Download Type",
		MinSize:  Size{200, 0},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: "Choose where to save the magnet:"},
			HSplitter{
				Children: []Widget{
					PushButton{Text: "special", OnClicked: func() { dlg.choice = "special"; dlg.Accept() }},
					PushButton{Text: "vr", OnClicked: func() { dlg.choice = "vr"; dlg.Accept() }},
				},
			},
		},
	}.Run(dlg)
	return dlg
}

type choiceDialog struct {
	*walk.Dialog
	choice string
}

func (d *choiceDialog) Run() string {
	if d.choice == "" {
		return ""
	}
	return d.choice
}
