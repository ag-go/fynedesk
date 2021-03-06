package ui

import (
	"os"
	"strings"
	"sync"

	"fyne.io/fyne"

	"fyne.io/fynedesk"
)

type deskSettings struct {
	background             string
	iconTheme              string
	launcherIcons          []string
	launcherIconSize       int
	launcherDisableTaskbar bool
	launcherDisableZoom    bool
	launcherZoomScale      float64

	listenerLock    sync.Mutex
	changeListeners []chan fynedesk.DeskSettings
}

func (d *deskSettings) Background() string {
	return d.background
}

func (d *deskSettings) IconTheme() string {
	return d.iconTheme
}

func (d *deskSettings) LauncherIcons() []string {
	return d.launcherIcons
}

func (d *deskSettings) LauncherIconSize() int {
	return d.launcherIconSize
}

func (d *deskSettings) LauncherDisableTaskbar() bool {
	return d.launcherDisableTaskbar
}

func (d *deskSettings) LauncherDisableZoom() bool {
	return d.launcherDisableZoom
}

func (d *deskSettings) LauncherZoomScale() float64 {
	return d.launcherZoomScale
}

func (d *deskSettings) AddChangeListener(listener chan fynedesk.DeskSettings) {
	d.listenerLock.Lock()
	defer d.listenerLock.Unlock()
	d.changeListeners = append(d.changeListeners, listener)
}

func (d *deskSettings) apply() {
	d.listenerLock.Lock()
	defer d.listenerLock.Unlock()

	for _, listener := range d.changeListeners {
		select {
		case listener <- d:
		default:
			l := listener
			go func() { l <- d }()
		}
	}
}

func (d *deskSettings) setBackground(name string) {
	d.background = name
	fyne.CurrentApp().Preferences().SetString("background", d.background)
	d.apply()
}

func (d *deskSettings) setIconTheme(name string) {
	d.iconTheme = name
	fyne.CurrentApp().Preferences().SetString("icontheme", d.iconTheme)
	d.apply()
}

func (d *deskSettings) setLauncherIcons(defaultApps []string) {
	newLauncherIcons := strings.Join(defaultApps, "|")
	d.launcherIcons = defaultApps
	fyne.CurrentApp().Preferences().SetString("launchericons", newLauncherIcons)
	d.apply()
}

func (d *deskSettings) setLauncherIconSize(size int) {
	d.launcherIconSize = size
	fyne.CurrentApp().Preferences().SetInt("launchericonsize", d.launcherIconSize)
	d.apply()
}

func (d *deskSettings) setLauncherDisableTaskbar(taskbar bool) {
	d.launcherDisableTaskbar = taskbar
	fyne.CurrentApp().Preferences().SetBool("launcherdisabletaskbar", d.launcherDisableTaskbar)
	d.apply()
}

func (d *deskSettings) setLauncherDisableZoom(zoom bool) {
	d.launcherDisableZoom = zoom
	fyne.CurrentApp().Preferences().SetBool("launcherdisablezoom", d.launcherDisableZoom)
	d.apply()
}

func (d *deskSettings) setLauncherZoomScale(scale float64) {
	d.launcherZoomScale = scale
	fyne.CurrentApp().Preferences().SetFloat("launcherzoomscale", d.launcherZoomScale)
	d.apply()
}

func (d *deskSettings) load() {
	env := os.Getenv("FYNEDESK_BACKGROUND")
	if env != "" {
		d.background = env
	} else {
		d.background = fyne.CurrentApp().Preferences().String("background")
	}

	env = os.Getenv("FYNEDESK_ICONTHEME")
	if env != "" {
		d.iconTheme = env
	} else {
		d.iconTheme = fyne.CurrentApp().Preferences().String("icontheme")
	}
	if d.iconTheme == "" {
		d.iconTheme = "hicolor"
	}

	launcherIcons := fyne.CurrentApp().Preferences().String("launchericons")
	if launcherIcons != "" {
		d.launcherIcons = strings.SplitN(fyne.CurrentApp().Preferences().String("launchericons"), "|", -1)
	}
	if len(d.launcherIcons) == 0 {
		defaultApps := fynedesk.Instance().IconProvider().DefaultApps()
		for _, appData := range defaultApps {
			d.launcherIcons = append(d.launcherIcons, appData.Name())
		}
	}

	d.launcherIconSize = fyne.CurrentApp().Preferences().Int("launchericonsize")
	if d.launcherIconSize == 0 {
		d.launcherIconSize = 48
	}

	d.launcherDisableTaskbar = fyne.CurrentApp().Preferences().Bool("launcherdisabletaskbar")
	d.launcherDisableZoom = fyne.CurrentApp().Preferences().Bool("launcherdisablezoom")

	d.launcherZoomScale = fyne.CurrentApp().Preferences().Float("launcherzoomscale")
	if d.launcherZoomScale == 0.0 {
		d.launcherZoomScale = 2.0
	}
}

// newDeskSettings loads the user's preferences from environment or config
func newDeskSettings() *deskSettings {
	settings := &deskSettings{}
	settings.load()

	return settings
}
