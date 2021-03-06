package main

import (
	"log"
	"runtime"

	"fyne.io/fyne"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal"
	"fyne.io/fynedesk/internal/ui"
)

func setupDesktop(a fyne.App) fynedesk.Desktop {
	log.Println("Full desktop not possible on", runtime.GOOS)
	return ui.NewEmbeddedDesktop(a, internal.NewMacOSAppProvider())
}
