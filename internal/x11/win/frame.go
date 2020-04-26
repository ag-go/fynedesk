// +build linux

package win

import (
	"image"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/tools/playground"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xwindow"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal/x11"
	wmTheme "fyne.io/fynedesk/theme"
	"fyne.io/fynedesk/wm"
)

type frame struct {
	x, y                                int16
	width, height                       uint16
	childWidth, childHeight             uint16
	resizeStartWidth, resizeStartHeight uint16
	mouseX, mouseY                      int16
	resizeStartX, resizeStartY          int16
	resizeBottom                        bool
	resizeLeft, resizeRight             bool
	moveOnly                            bool

	borderTop, borderTopRight xproto.Pixmap
	borderTopWidth            uint16

	client *client
}

func newFrame(c *client) *frame {
	attrs, err := xproto.GetGeometry(c.wm.Conn(), xproto.Drawable(c.win)).Reply()
	if err != nil {
		fyne.LogError("Get Geometry Error", err)
		return nil
	}

	f, err := xwindow.Generate(c.wm.X())
	if err != nil {
		fyne.LogError("Generate Window Error", err)
		return nil
	}
	x, y, w, h := attrs.X, attrs.Y, attrs.Width, attrs.Height
	full := c.Fullscreened()
	decorated := c.Properties().Decorated()
	maximized := c.Maximized()
	screen := fynedesk.Instance().Screens().ScreenForGeometry(int(x), int(y), int(w), int(h))
	borderWidth := x11.ScaleToPixels(wmTheme.BorderWidth, screen)
	titleHeight := x11.ScaleToPixels(wmTheme.TitleHeight, screen)
	if full || maximized {
		activeHead := fynedesk.Instance().Screens().ScreenForGeometry(int(attrs.X), int(attrs.Y), int(attrs.Width), int(attrs.Height))
		x = int16(activeHead.X)
		y = int16(activeHead.Y)
		if full {
			w = uint16(activeHead.Width)
			h = uint16(activeHead.Height)
		} else {
			maxWidth, maxHeight := fynedesk.Instance().ContentSizePixels(activeHead)
			w = uint16(maxWidth)
			h = uint16(maxHeight)
		}
	} else if decorated {
		x -= int16(borderWidth)
		y -= int16(titleHeight)
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}
		if !maximized {
			w = attrs.Width + borderWidth*2
			h = attrs.Height + borderWidth + titleHeight
		}
	}
	framed := &frame{client: c}
	framed.x = x
	framed.y = y
	values := []uint32{xproto.EventMaskStructureNotify | xproto.EventMaskSubstructureNotify |
		xproto.EventMaskSubstructureRedirect | xproto.EventMaskExposure |
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease | xproto.EventMaskButtonMotion |
		xproto.EventMaskKeyPress | xproto.EventMaskPointerMotion | xproto.EventMaskFocusChange |
		xproto.EventMaskPropertyChange}
	err = xproto.CreateWindowChecked(c.wm.Conn(), c.wm.X().Screen().RootDepth, f.Id, c.wm.X().RootWin(),
		x, y, w, h, 0, xproto.WindowClassInputOutput, c.wm.X().Screen().RootVisual,
		xproto.CwEventMask, values).Check()
	if err != nil {
		fyne.LogError("Create Window Error", err)
		return nil
	}
	c.id = f.Id

	framed.width = w
	framed.height = h
	if full || !decorated {
		framed.childWidth = w
		framed.childHeight = h
	} else {
		framed.childWidth = w - borderWidth*2
		framed.childHeight = h - borderWidth - titleHeight
	}

	var offsetX, offsetY int16 = 0, 0
	if !full && decorated {
		offsetX = int16(borderWidth)
		offsetY = int16(titleHeight)
		xproto.ReparentWindow(c.wm.Conn(), c.win, c.id, int16(borderWidth), int16(titleHeight))
		ewmh.FrameExtentsSet(c.wm.X(), c.win, &ewmh.FrameExtents{Left: int(borderWidth), Right: int(borderWidth),
			Top: int(titleHeight), Bottom: int(borderWidth)})
	} else {
		xproto.ReparentWindow(c.wm.Conn(), c.win, c.id, attrs.X, attrs.Y)
		ewmh.FrameExtentsSet(c.wm.X(), c.win, &ewmh.FrameExtents{Left: 0, Right: 0, Top: 0, Bottom: 0})
	}

	if full || maximized {
		err = xproto.ConfigureWindowChecked(c.wm.Conn(), c.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
			[]uint32{uint32(offsetX), uint32(offsetY), uint32(framed.childWidth), uint32(framed.childHeight)}).Check()
		if err != nil {
			fyne.LogError("Configure Window Error", err)
		}
	}

	windowStateSet(c.wm.X(), c.win, icccm.StateNormal)
	framed.show()
	framed.applyTheme(true)
	framed.notifyInnerGeometry()

	return framed
}

func (f *frame) addBorder() {
	borderWidth := x11.BorderWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))
	x := int16(borderWidth)
	y := int16(titleHeight)
	w := f.width
	h := f.height
	if !f.client.maximized {
		w := f.childWidth + borderWidth*2
		h := f.childHeight + borderWidth + titleHeight
		f.x -= x
		f.y -= y
		if f.x < 0 {
			f.x = 0
		}
		if f.y < 0 {
			f.y = 0
		}
		f.width = w
		f.height = h
	}
	f.applyTheme(true)

	err := xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(x), uint32(y), uint32(f.childWidth), uint32(f.childHeight)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
	err = xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(f.x), uint32(f.y), uint32(w), uint32(h)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}

	ewmh.FrameExtentsSet(f.client.wm.X(), f.client.win, &ewmh.FrameExtents{Left: int(borderWidth), Right: int(borderWidth), Top: int(titleHeight), Bottom: int(borderWidth)})
	f.notifyInnerGeometry()
}

func (f *frame) applyBorderlessTheme() {
	err := xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(0), uint32(0), uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
}

func (f *frame) applyTheme(force bool) {
	if f.client.Fullscreened() || !f.client.Properties().Decorated() {
		f.applyBorderlessTheme()
		return
	}

	f.checkScale()
	f.decorate(force)
}

func (f *frame) checkScale() {
	titleHeight := x11.TitleHeight(x11.XWin(f.client))
	if f.height - titleHeight != f.childHeight {
		f.updateGeometry(f.x, f.y, f.width, f.height, true)
		f.notifyInnerGeometry()
	}
}

func (f *frame) copyDecorationPixels(width, height, xoff, yoff uint32, img image.Image, pid xproto.Pixmap, draw xproto.Gcontext, depth byte) {
	// DATA is BGRx
	data := make([]byte, width*height*4)
	i := uint32(0)
	for y := uint32(0); y < height; y++ {
		for x := uint32(0); x < width; x++ {
			r, g, b, _ := img.At(int(xoff+x), int(yoff+y)).RGBA()

			data[i] = byte(b)
			data[i+1] = byte(g)
			data[i+2] = byte(r)
			data[i+3] = 0

			i += 4
		}
	}
	err := xproto.PutImageChecked(f.client.wm.Conn(), xproto.ImageFormatZPixmap, xproto.Drawable(pid), draw,
		uint16(width), uint16(height), 0, int16(yoff), 0, depth, data).Check()
	if err != nil {
		fyne.LogError("Put image error", err)
	}
}

func (f *frame) createPixmaps(depth byte) error {
	iconPix := x11.TitleHeight(x11.XWin(f.client))
	iconAndBorderPix := iconPix + x11.BorderWidth(x11.XWin(f.client))*2
	f.borderTopWidth = f.width - iconAndBorderPix

	pid, err := xproto.NewPixmapId(f.client.wm.Conn())
	if err != nil {
		return err
	}

	xproto.CreatePixmap(f.client.wm.Conn(), depth, pid,
		xproto.Drawable(f.client.wm.X().Screen().Root), f.borderTopWidth, iconPix)
	f.borderTop = pid

	pid, err = xproto.NewPixmapId(f.client.wm.Conn())
	if err != nil {
		return err
	}

	xproto.CreatePixmap(f.client.wm.Conn(), depth, pid,
		xproto.Drawable(f.client.wm.X().Screen().Root), iconAndBorderPix, iconPix)
	f.borderTopRight = pid

	return nil
}

func (f *frame) decorate(force bool) {
	depth := f.client.wm.X().Screen().RootDepth
	refresh := force

	if f.borderTop == 0 || refresh {
		err := f.createPixmaps(depth)
		if err != nil {
			fyne.LogError("New Pixmap Error", err)
			return
		}
		refresh = true
	}

	backR, backG, backB, _ := theme.ButtonColor().RGBA()
	if f.client.Focused() {
		backR, backG, backB, _ = theme.BackgroundColor().RGBA()
	}
	bgColor := backR<<16 | backG<<8 | backB

	drawTop, _ := xproto.NewGcontextId(f.client.wm.Conn())
	xproto.CreateGC(f.client.wm.Conn(), drawTop, xproto.Drawable(f.borderTop), xproto.GcForeground, []uint32{bgColor})
	drawTopRight, _ := xproto.NewGcontextId(f.client.wm.Conn())
	xproto.CreateGC(f.client.wm.Conn(), drawTopRight, xproto.Drawable(f.borderTopRight), xproto.GcForeground, []uint32{bgColor})

	if refresh {
		f.drawDecoration(f.borderTop, drawTop, f.borderTopRight, drawTopRight, depth)
	}

	iconSizePix := x11.TitleHeight(x11.XWin(f.client))
	draw, _ := xproto.NewGcontextId(f.client.wm.Conn())
	xproto.CreateGC(f.client.wm.Conn(), draw, xproto.Drawable(f.client.id), xproto.GcForeground, []uint32{bgColor})
	rect := xproto.Rectangle{X: 0, Y: int16(iconSizePix), Width: f.width, Height: f.height - iconSizePix}
	xproto.PolyFillRectangleChecked(f.client.wm.Conn(), xproto.Drawable(f.client.id), draw, []xproto.Rectangle{rect})

	iconAndBorderSizePix := iconSizePix + x11.BorderWidth(x11.XWin(f.client))*2
	if f.borderTopWidth+iconAndBorderSizePix < f.width {
		rect := xproto.Rectangle{X: int16(f.borderTopWidth), Y: 0,
			Width: f.width - f.borderTopWidth - iconAndBorderSizePix, Height: iconSizePix}
		xproto.PolyFillRectangleChecked(f.client.wm.Conn(), xproto.Drawable(f.client.id), draw, []xproto.Rectangle{rect})
	}

	xproto.CopyArea(f.client.wm.Conn(), xproto.Drawable(f.borderTop), xproto.Drawable(f.client.id), drawTop,
		0, 0, 0, 0, f.borderTopWidth, iconSizePix)
	xproto.CopyArea(f.client.wm.Conn(), xproto.Drawable(f.borderTopRight), xproto.Drawable(f.client.id), drawTopRight,
		0, 0, int16(f.width-iconAndBorderSizePix), 0, iconAndBorderSizePix, iconSizePix)
}

func (f *frame) drawDecoration(pidTop xproto.Pixmap, drawTop xproto.Gcontext, pidTopRight xproto.Pixmap, drawTopRight xproto.Gcontext, depth byte) {
	screen := fynedesk.Instance().Screens().ScreenForWindow(f.client)
	scale := screen.CanvasScale()

	canvas := playground.NewSoftwareCanvas()
	canvas.SetScale(scale)
	canvas.SetPadded(false)
	canMaximize := true
	if windowSizeFixed(f.client.wm.X(), f.client.win) ||
		!windowSizeCanMaximize(f.client.wm.X(), f.client) {
		canMaximize = false
	}
	canvas.SetContent(wm.NewBorder(f.client, f.client.Properties().Icon(), canMaximize))

	heightPix := x11.TitleHeight(x11.XWin(f.client))
	iconBorderPixWidth := heightPix + x11.BorderWidth(x11.XWin(f.client))*2
	widthPix := f.borderTopWidth + iconBorderPixWidth
	canvas.Resize(fyne.NewSize(int(float32(widthPix)/scale)+1, wmTheme.TitleHeight))
	img := canvas.Capture()

	// TODO just copy the label minSize - smallest possible but maybe bigger than window width
	// Draw in pixel rows so we don't overflow count usable by PutImageChecked
	for i := uint16(0); i < heightPix; i++ {
		f.copyDecorationPixels(uint32(f.borderTopWidth), 1, 0, uint32(i), img, pidTop, drawTop, depth)
	}

	f.copyDecorationPixels(uint32(iconBorderPixWidth), uint32(heightPix), uint32(f.borderTopWidth), 0, img, pidTopRight, drawTopRight, depth)
}

func (f *frame) getGeometry() (int16, int16, uint16, uint16) {
	return f.x, f.y, f.width, f.height
}

func (f *frame) getInnerWindowCoordinates(w uint16, h uint16) (uint32, uint32, uint32, uint32) {
	if f.client.Fullscreened() || !f.client.Properties().Decorated() {
		constrainW, constrainH := w, h
		if !f.client.Properties().Decorated() {
			adjustedW, adjustedH := windowSizeWithIncrement(f.client.wm.X(), f.client.win, w, h)
			constrainW, constrainH = windowSizeConstrain(f.client.wm.X(), f.client.win,
				adjustedW, adjustedH)
		}
		f.width = constrainW
		f.height = constrainH
		f.height = constrainH
		return 0, 0, uint32(constrainW), uint32(constrainH)
	}

	borderWidth := x11.BorderWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))

	extraWidth := 2 * borderWidth
	extraHeight := borderWidth + titleHeight
	w -= extraWidth
	h -= extraHeight
	adjustedW, adjustedH := windowSizeWithIncrement(f.client.wm.X(), f.client.win, w, h)
	constrainW, constrainH := windowSizeConstrain(f.client.wm.X(), f.client.win,
		adjustedW, adjustedH)
	f.width = constrainW + extraWidth
	f.height = constrainH + extraHeight

	return uint32(borderWidth), uint32(titleHeight), uint32(constrainW), uint32(constrainH)
}

func (f *frame) hide() {
	f.client.RaiseToTop() // Lets ensure this client is on top of the stack so we can walk backwards to find the next window to focus
	stack := f.client.wm.Windows()
	for i := len(stack) - 1; i >= 0; i-- {
		if !stack[i].Iconic() {
			stack[i].RaiseToTop()
			stack[i].Focus()
		}
	}
	xproto.ReparentWindow(f.client.wm.Conn(), f.client.win, f.client.wm.X().RootWin(), f.x, f.y)
	xproto.UnmapWindow(f.client.wm.Conn(), f.client.win)
}

func (f *frame) maximizeApply() {
	if windowSizeFixed(f.client.wm.X(), f.client.win) ||
		!windowSizeCanMaximize(f.client.wm.X(), f.client) {
		return
	}
	f.client.restoreWidth = f.width
	f.client.restoreHeight = f.height
	f.client.restoreX = f.x
	f.client.restoreY = f.y

	head := fynedesk.Instance().Screens().ScreenForWindow(f.client)
	maxWidth, maxHeight := fynedesk.Instance().ContentSizePixels(head)
	if f.client.Fullscreened() {
		maxWidth = uint32(head.Width)
		maxHeight = uint32(head.Height)
	}
	f.updateGeometry(int16(head.X), int16(head.Y), uint16(maxWidth), uint16(maxHeight), true)
	f.notifyInnerGeometry()
	f.applyTheme(true)
}

func (f *frame) mouseDrag(x, y int16) {
	if f.client.Maximized() || f.client.Fullscreened() {
		return
	}
	moveDeltaX := x - f.mouseX
	moveDeltaY := y - f.mouseY
	f.mouseX = x
	f.mouseY = y
	if moveDeltaX == 0 && moveDeltaY == 0 {
		return
	}

	if f.resizeBottom || f.resizeLeft || f.resizeRight && !windowSizeFixed(f.client.wm.X(), f.client.win) {
		deltaX := x - f.resizeStartX
		deltaY := y - f.resizeStartY
		x := f.x
		width := f.resizeStartWidth
		height := f.resizeStartHeight
		if f.resizeBottom {
			height += uint16(deltaY)
		}
		if f.resizeLeft {
			x += moveDeltaX
			width -= uint16(deltaX)
		} else if f.resizeRight {
			width += uint16(deltaX)
		}
		f.updateGeometry(x, f.y, width, height, false)
	} else if f.moveOnly {
		f.updateGeometry(f.x+moveDeltaX, f.y+moveDeltaY, f.width, f.height, false)
	}
}

func (f *frame) mouseMotion(x, y int16) {
	borderWidth := x11.BorderWidth(x11.XWin(f.client))
	buttonWidth := x11.ButtonWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))

	relX := x - f.x
	relY := y - f.y
	cursor := x11.DefaultCursor
	if relY <= int16(titleHeight) { // title bar
		if relX > int16(borderWidth) && relX <= int16(borderWidth+buttonWidth) {
			cursor = x11.CloseCursor
		}
		err := xproto.ChangeWindowAttributesChecked(f.client.wm.Conn(), f.client.id, xproto.CwCursor,
			[]uint32{uint32(cursor)}).Check()
		if err != nil {
			fyne.LogError("Set Cursor Error", err)
		}
		return
	}
	if f.client.Maximized() || f.client.Fullscreened() || windowSizeFixed(f.client.wm.X(), f.client.win) {
		return
	}

	if relY >= int16(f.height-buttonWidth) { // bottom
		if relX < int16(buttonWidth) {
			cursor = x11.ResizeBottomLeftCursor
		} else if relX >= int16(f.width-buttonWidth) {
			cursor = x11.ResizeBottomRightCursor
		} else {
			cursor = x11.ResizeBottomCursor
		}
	} else { // center (sides)
		if relX < int16(f.width-buttonWidth) {
			cursor = x11.ResizeLeftCursor
		} else if relX >= int16(f.width-buttonWidth) {
			cursor = x11.ResizeRightCursor
		}
	}

	err := xproto.ChangeWindowAttributesChecked(f.client.wm.Conn(), f.client.id, xproto.CwCursor,
		[]uint32{uint32(cursor)}).Check()
	if err != nil {
		fyne.LogError("Set Cursor Error", err)
	}
}

func (f *frame) mousePress(x, y int16) {
	if !f.client.Focused() {
		f.client.RaiseToTop()
		f.client.Focus()
		return
	}
	if f.client.Maximized() || f.client.Fullscreened() {
		return
	}

	buttonWidth := x11.ButtonWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))
	f.mouseX = x
	f.mouseY = y
	f.resizeStartX = x
	f.resizeStartY = y

	relX := x - f.x
	relY := y - f.y
	f.resizeStartWidth = f.width
	f.resizeStartHeight = f.height
	f.resizeBottom = false
	f.resizeLeft = false
	f.resizeRight = false
	f.moveOnly = false

	if relY >= int16(titleHeight) && !windowSizeFixed(f.client.wm.X(), f.client.win) {
		if relY >= int16(f.height-buttonWidth) {
			f.resizeBottom = true
		}
		if relX < int16(buttonWidth) {
			f.resizeLeft = true
		} else if relX >= int16(f.width-buttonWidth) {
			f.resizeRight = true
		}
	} else if relY < int16(titleHeight) {
		f.moveOnly = true
	}

	f.client.wm.RaiseToTop(f.client)
}

func (f *frame) mouseRelease(x, y int16) {
	borderWidth := x11.BorderWidth(x11.XWin(f.client))
	buttonWidth := x11.ButtonWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))

	relX := x - f.x
	relY := y - f.y
	barYMax := int16(titleHeight)
	if relY > barYMax {
		return
	}
	if relX >= int16(borderWidth) && relX < int16(borderWidth+buttonWidth) {
		f.client.Close()
	}
	if relX >= int16(borderWidth)+int16(theme.Padding())+int16(buttonWidth) &&
		relX < int16(borderWidth)+int16(theme.Padding()*2)+int16(buttonWidth*2) {
		if f.client.Maximized() {
			f.client.Unmaximize()
		} else {
			f.client.Maximize()
		}
	} else if relX >= int16(borderWidth)+int16(theme.Padding()*2)+int16(buttonWidth*2) &&
		relX < int16(borderWidth)+int16(theme.Padding()*2)+int16(buttonWidth*3) {
		f.client.Iconify()
	}

	f.resizeBottom = false
	f.resizeLeft = false
	f.resizeRight = false
	f.moveOnly = false
	f.updateGeometry(f.x, f.y, f.width, f.height, false)
}

// Notify the child window that it's geometry has changed to update menu positions etc.
// This should be used sparingly as it can impact performance on the child window.
func (f *frame) notifyInnerGeometry() {
	innerX, innerY, innerW, innerH := f.getInnerWindowCoordinates(f.width, f.height)
	ev := xproto.ConfigureNotifyEvent{Event: f.client.win, Window: f.client.win, AboveSibling: 0,
		X: int16(f.x + int16(innerX)), Y: int16(f.y + int16(innerY)), Width: uint16(innerW), Height: uint16(innerH),
		BorderWidth: x11.BorderWidth(x11.XWin(f.client)), OverrideRedirect: false}
	xproto.SendEvent(f.client.wm.Conn(), false, f.client.win, xproto.EventMaskStructureNotify, string(ev.Bytes()))
}

func (f *frame) removeBorder() {
	borderWidth := x11.BorderWidth(x11.XWin(f.client))
	titleHeight := x11.TitleHeight(x11.XWin(f.client))

	if !f.client.maximized {
		f.x += int16(borderWidth)
		f.y += int16(titleHeight)
		f.width = f.childWidth
		f.height = f.childHeight
	}
	f.applyTheme(true)

	err := xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(f.x), uint32(f.y), uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
	err = xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{0, 0, uint32(f.childWidth), uint32(f.childHeight)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}

	ewmh.FrameExtentsSet(f.client.wm.X(), f.client.win, &ewmh.FrameExtents{Left: 0, Right: 0, Top: 0, Bottom: 0})
	f.notifyInnerGeometry()
}

func (f *frame) show() {
	c := f.client
	xproto.MapWindow(c.wm.Conn(), c.id)

	xproto.ChangeSaveSet(c.wm.Conn(), xproto.SetModeInsert, c.win)
	xproto.MapWindow(c.wm.Conn(), c.win)
	c.wm.BindKeys(c)
	xproto.GrabButton(f.client.wm.Conn(), true, f.client.id,
		xproto.EventMaskButtonPress, xproto.GrabModeSync, xproto.GrabModeSync,
		f.client.wm.X().RootWin(), xproto.CursorNone, xproto.ButtonIndex1, xproto.ModMaskAny)

	c.RaiseToTop()
	c.Focus()
}

func (f *frame) unFrame() {
	c := f.client
	c.wm.RemoveWindow(c)

	if f != nil {
		xproto.ReparentWindow(c.wm.Conn(), c.win, c.wm.X().RootWin(), f.x, f.y)
	}
	xproto.ChangeSaveSet(c.wm.Conn(), xproto.SetModeDelete, c.win)
	xproto.UnmapWindow(c.wm.Conn(), c.id)
}

func (f *frame) unmaximizeApply() {
	if windowSizeFixed(f.client.wm.X(), f.client.win) ||
		!windowSizeCanMaximize(f.client.wm.X(), f.client) {
		return
	}
	if f.client.restoreWidth == 0 && f.client.restoreHeight == 0 {
		screen := fynedesk.Instance().Screens().ScreenForWindow(f.client)
		f.client.restoreWidth = uint16(screen.Width / 2)
		f.client.restoreHeight = uint16(screen.Height / 2)
	}
	f.updateGeometry(f.client.restoreX, f.client.restoreY, f.client.restoreWidth, f.client.restoreHeight, true)
	f.notifyInnerGeometry()
	f.applyTheme(true)
}

func (f *frame) updateGeometry(x, y int16, w, h uint16, force bool) {
	var move, resize bool
	if !force {
		resize = w != f.width || h != f.height
		move = x != f.x || y != f.y
		if !move && !resize {
			return
		}
	}

	currentScreen := fynedesk.Instance().Screens().ScreenForWindow(f.client)

	f.x = x
	f.y = y

	innerX, innerY, innerW, innerH := f.getInnerWindowCoordinates(w, h)

	f.childWidth = uint16(innerW)
	f.childHeight = uint16(innerH)

	err := xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.id, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{uint32(f.x), uint32(f.y), uint32(f.width), uint32(f.height)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}
	err = xproto.ConfigureWindowChecked(f.client.wm.Conn(), f.client.win, xproto.ConfigWindowX|xproto.ConfigWindowY|
		xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
		[]uint32{innerX, innerY, uint32(f.childWidth), uint32(f.childHeight)}).Check()
	if err != nil {
		fyne.LogError("Configure Window Error", err)
	}

	newScreen := fynedesk.Instance().Screens().ScreenForWindow(f.client)
	if newScreen != currentScreen {
		f.updateScale()
	}
}

func (f *frame) updateScale() {
	xproto.FreePixmap(f.client.wm.Conn(), f.borderTop)
	f.borderTop = 0
	xproto.FreePixmap(f.client.wm.Conn(), f.borderTopRight)
	f.borderTopRight = 0

	f.updateGeometry(f.x, f.y, f.width, f.height, true)
	f.applyTheme(true)
}