// +build linux

package wm // import "fyne.io/fynedesk/wm"

import (
	"math"

	"fyne.io/fyne"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"

	"fyne.io/fynedesk"
)

type x11ScreensProvider struct {
	screens []*fynedesk.Screen
	active  *fynedesk.Screen
	primary *fynedesk.Screen
	scale   float32
}

// NewX11ScreensProvider returns a screen provider for use in x11 desktop mode
func NewX11ScreensProvider(mgr fynedesk.WindowManager) fynedesk.ScreenList {
	screensProvider := &x11ScreensProvider{}
	x := mgr.(*x11WM)
	screensProvider.setupScreens(x)
	return screensProvider
}

func (xsp *x11ScreensProvider) Screens() []*fynedesk.Screen {
	return xsp.screens
}

func (xsp *x11ScreensProvider) Active() *fynedesk.Screen {
	return xsp.active
}

func (xsp *x11ScreensProvider) Primary() *fynedesk.Screen {
	return xsp.primary
}

func (xsp *x11ScreensProvider) Scale() float32 {
	return xsp.scale
}

func (xsp *x11ScreensProvider) ScreenForWindow(win fynedesk.Window) *fynedesk.Screen {
	fr := win.(*client).frame
	if fr == nil || len(xsp.screens) <= 1 {
		return xsp.screens[0]
	}
	return xsp.ScreenForGeometry(int(fr.x), int(fr.y), int(fr.width), int(fr.height))
}

func (xsp *x11ScreensProvider) ScreenForGeometry(x int, y int, width int, height int) *fynedesk.Screen {
	if len(xsp.screens) <= 1 {
		return xsp.screens[0]
	}
	for i := 0; i < len(xsp.screens); i++ {
		xx, yy, ww, hh := xsp.screens[i].X, xsp.screens[i].Y,
			xsp.screens[i].Width, xsp.screens[i].Height
		middleW := width / 2
		middleH := height / 2
		middleW += x
		middleH += y
		if middleW >= xx && middleH >= yy &&
			middleW <= xx+ww && middleH <= yy+hh {
			return xsp.screens[i]
		}
	}
	return xsp.screens[0]
}

func getScale(widthPx uint16, widthMm uint32) float32 {
	dpi := float32(widthPx) / (float32(widthMm) / 25.4)
	if dpi > 1000 || dpi < 10 {
		dpi = 96
	}

	screenScale := dpi / 96.0 * 10.0
	userScale := fyne.CurrentApp().Settings().Scale()
	if userScale == fyne.SettingsScaleAuto || userScale == 0.0 { // 0.0 is the old auto, or the missing config
		userScale = 1.0
	}
	return float32(math.Round(float64(screenScale*userScale))) / 10.0
}

func (xsp *x11ScreensProvider) setupScreens(x *x11WM) {
	err := randr.Init(x.x.Conn())
	if err != nil {
		fyne.LogError("Could not initialize randr", err)
		xsp.setupSingleScreen(x)
		return
	}

	root := xproto.Setup(x.x.Conn()).DefaultScreen(x.x.Conn()).Root
	resources, err := randr.GetScreenResources(x.x.Conn(), root).Reply()
	if err != nil || len(resources.Outputs) == 0 {
		fyne.LogError("Could not get randr screen resources", err)
		xsp.setupSingleScreen(x)
		return
	}

	var primaryInfo *randr.GetOutputInfoReply
	primary, err := randr.GetOutputPrimary(x.x.Conn(),
		xproto.Setup(x.x.Conn()).DefaultScreen(x.x.Conn()).Root).Reply()
	if err == nil {
		primaryInfo, _ = randr.GetOutputInfo(x.x.Conn(), primary.Output, 0).Reply()
	}
	primaryFound := false
	var firstFoundMmWidth uint32
	var firstFoundWidth uint16
	i := 0
	for _, output := range resources.Outputs {
		outputInfo, err := randr.GetOutputInfo(x.x.Conn(), output, 0).Reply()
		if err != nil {
			fyne.LogError("Could not get randr output", err)
			continue
		}
		if outputInfo.Crtc == 0 || outputInfo.Connection == randr.ConnectionDisconnected {
			continue
		}
		crtcInfo, err := randr.GetCrtcInfo(x.x.Conn(), outputInfo.Crtc, 0).Reply()
		if err != nil {
			fyne.LogError("Could not get randr crtcs", err)
			continue
		}
		if i == 0 {
			firstFoundMmWidth = outputInfo.MmWidth
			firstFoundWidth = crtcInfo.Width
		}
		xsp.screens = append(xsp.screens, &fynedesk.Screen{Name: string(outputInfo.Name),
			X: int(crtcInfo.X), Y: int(crtcInfo.Y), Width: int(crtcInfo.Width), Height: int(crtcInfo.Height)})
		if primaryInfo != nil {
			if string(primaryInfo.Name) == string(outputInfo.Name) {
				primaryFound = true
				xsp.primary = xsp.screens[i]
				xsp.active = xsp.screens[i]
				xsp.scale = getScale(crtcInfo.Width, outputInfo.MmWidth)
			}
		}
		i++
	}
	if !primaryFound {
		xsp.primary = xsp.screens[0]
		xsp.active = xsp.screens[0]
		xsp.scale = getScale(firstFoundWidth, firstFoundMmWidth)
	}
}

func (xsp *x11ScreensProvider) setupSingleScreen(x *x11WM) {
	xsp.screens = append(xsp.screens, &fynedesk.Screen{Name: "Screen0",
		X: xwindow.RootGeometry(x.x).X(), Y: xwindow.RootGeometry(x.x).Y(),
		Width: xwindow.RootGeometry(x.x).Width(), Height: xwindow.RootGeometry(x.x).Height()})
	xsp.primary = xsp.screens[0]
	xsp.active = xsp.screens[0]
}
