package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	_ "fyne.io/fynedesk/modules/composit"
	_ "fyne.io/fynedesk/modules/desktops"
	_ "fyne.io/fynedesk/modules/launcher"
	_ "fyne.io/fynedesk/modules/status"
	_ "fyne.io/fynedesk/modules/systray"
)

func main() {
	a := app.NewWithID("io.fyne.fynedesk")
	a.SetIcon(theme.FyneLogo())
	desk := setupDesktop(a)

	desk.Run()
}
