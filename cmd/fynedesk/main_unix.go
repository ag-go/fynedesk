// +build linux openbsd freebsd netbsd

package main

import (
	"log"

	"fyne.io/fyne"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal"
	"fyne.io/fynedesk/internal/ui"
	"fyne.io/fynedesk/wm"
)

func setupDesktop(a fyne.App) fynedesk.Desktop {
	icons := internal.NewFDOIconProvider()
	mgr, err := wm.NewX11WindowManager(a)
	if err != nil {
		log.Println("Could not create window manager:", err)
		return ui.NewEmbeddedDesktop(a, icons)
	}
	desk := ui.NewDesktop(a, mgr, icons, wm.NewX11ScreensProvider(mgr))
	mgr.SetRoot(desk.Root())
	return desk
}
