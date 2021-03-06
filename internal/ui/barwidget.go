package ui

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	deskDriver "fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"fyne.io/fynedesk"
)

//bar is the main widget housing the icon launcher
type bar struct {
	widget.BaseWidget

	desk          fynedesk.Desktop    // The desktop instance we are holding icons for
	children      []fyne.CanvasObject // Icons that are laid out by the bar
	mouseInside   bool                // Is the mouse inside of the bar?
	mousePosition fyne.Position       // The current coordinates of the mouse cursor

	iconSize       int
	iconScale      float32
	disableTaskbar bool
	disableZoom    bool
	icons          []*barIcon
}

//MouseIn alerts the widget that the mouse has entered
func (b *bar) MouseIn(*deskDriver.MouseEvent) {
	if b.desk.Settings().LauncherDisableZoom() {
		return
	}
	b.mouseInside = true
	b.Refresh()
}

//MouseOut alerts the widget that the mouse has left
func (b *bar) MouseOut() {
	if b.desk.Settings().LauncherDisableZoom() {
		return
	}
	b.mouseInside = false
	b.Refresh()
}

//MouseMoved alerts the widget that the mouse has changed position
func (b *bar) MouseMoved(event *deskDriver.MouseEvent) {
	if b.desk.Settings().LauncherDisableZoom() {
		return
	}
	b.mousePosition = event.Position
	b.Refresh()
}

//append adds an object to the end of the widget
func (b *bar) append(object fyne.CanvasObject) {
	if b.Hidden && object.Visible() {
		object.Hide()
	}
	b.children = append(b.children, object)

	b.Refresh()
}

//appendSeparator adds a separator between the default icons and the taskbar
func (b *bar) appendSeparator() {
	line := canvas.NewRectangle(theme.TextColor())
	b.append(line)
}

//removeFromTaskbar removes an object from the taskbar area of the widget
func (b *bar) removeFromTaskbar(object fyne.CanvasObject) {
	if b.Hidden && object.Visible() {
		object.Hide()
	}
	for i, fycon := range b.children {
		if fycon == object {
			b.children = append(b.children[:i], b.children[i+1:]...)
		}
	}

	b.Refresh()
}

//CreateRenderer creates the renderer that will be responsible for painting the widget
func (b *bar) CreateRenderer() fyne.WidgetRenderer {
	return &barRenderer{objects: b.children, layout: newBarLayout(b), appBar: b}
}

//newAppBar returns a horizontal list of icons for an icon launcher
func newAppBar(desk fynedesk.Desktop, children ...fyne.CanvasObject) *bar {
	bar := &bar{desk: desk, children: children}
	bar.ExtendBaseWidget(bar)
	bar.iconSize = desk.Settings().LauncherIconSize()
	bar.iconScale = float32(desk.Settings().LauncherZoomScale())
	bar.disableTaskbar = desk.Settings().LauncherDisableTaskbar()

	return bar
}

//barRenderer provides the renderer functions for the bar Widget
type barRenderer struct {
	layout barLayout

	appBar  *bar
	objects []fyne.CanvasObject
}

//MinSize returns the layout's Min Size
func (b *barRenderer) MinSize() fyne.Size {
	return b.layout.MinSize(b.objects)
}

//Layout recalculates the widget
func (b *barRenderer) Layout(size fyne.Size) {
	b.layout.setPointerInside(b.appBar.mouseInside)
	b.layout.setPointerPosition(b.appBar.mousePosition)
	b.layout.Layout(b.objects, size)
}

//BackgroundColor returns the background color of the widget
func (b *barRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

//Objects returns the objects associated with the widget
func (b *barRenderer) Objects() []fyne.CanvasObject {
	return b.objects
}

//Refresh will recalculate the widget and repaint it
func (b *barRenderer) Refresh() {
	b.objects = b.appBar.children
	b.Layout(b.appBar.Size())

	canvas.Refresh(b.appBar)
}

//Destroy destroys the renderer
func (b *barRenderer) Destroy() {

}
