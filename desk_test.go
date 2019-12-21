package desktop

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/test"
	"github.com/stretchr/testify/assert"

	wmTheme "fyne.io/desktop/theme"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
)

type testDesk struct {
	settings DeskSettings
	icons    ApplicationProvider
}

func (*testDesk) Root() fyne.Window {
	return test.NewWindow(nil)
}

func (*testDesk) Run() {
}

func (*testDesk) RunApp(app AppData) error {
	return app.Run([]string{}) // no added env
}

func (td *testDesk) Settings() DeskSettings {
	return td.settings
}

func (*testDesk) ContentSizePixels(screen *Screen) (uint32, uint32) {
	return uint32(320), uint32(240)
}

func (td *testDesk) IconProvider() ApplicationProvider {
	return td.icons
}

func (*testDesk) WindowManager() WindowManager {
	return nil
}

func (*testDesk) Screens() ScreenList {
	return nil
}

type testSettings struct {
	background             string
	iconTheme              string
	launcherIcons          []string
	launcherIconSize       int
	launcherZoomScale      float64
	launcherDisableZoom    bool
	launcherDisableTaskbar bool
}

func (ts *testSettings) IconTheme() string {
	return ts.iconTheme
}

func (ts *testSettings) Background() string {
	return ts.background
}

func (ts *testSettings) LauncherIcons() []string {
	return ts.launcherIcons
}

func (ts *testSettings) LauncherIconSize() int {
	if ts.launcherIconSize == 0 {
		return 32
	}
	return ts.launcherIconSize
}

func (ts *testSettings) LauncherDisableTaskbar() bool {
	return ts.launcherDisableTaskbar
}

func (ts *testSettings) LauncherDisableZoom() bool {
	return ts.launcherDisableZoom
}

func (ts *testSettings) LauncherZoomScale() float64 {
	if ts.launcherZoomScale == 0 {
		return 1.0
	}
	return ts.launcherZoomScale
}

func (*testSettings) AddChangeListener(listener chan DeskSettings) {
	return
}

type testScreensProvider struct {
	screens []*Screen
}

func (tsp testScreensProvider) Screens() []*Screen {
	if tsp.screens == nil {
		tsp.screens = []*Screen{{"Screen0", 0, 0, 2000, 1000}}
	}
	return tsp.screens
}

func (tsp testScreensProvider) Active() *Screen {
	return tsp.Screens()[0]
}

func (tsp testScreensProvider) Primary() *Screen {
	return tsp.Screens()[0]
}

func (tsp testScreensProvider) Scale() float32 {
	return 1.0
}

func (tsp testScreensProvider) ScreenForWindow(win Window) *Screen {
	return tsp.Screens()[0]
}

func (tsp testScreensProvider) ScreenForGeometry(x int, y int, width int, height int) *Screen {
	return tsp.Screens()[0]
}

type testAppData struct {
	name string
}

func (tad *testAppData) Name() string {
	return tad.name
}

func (tad *testAppData) Run([]string) error {
	return nil
}

func (tad *testAppData) Icon(theme string, size int) fyne.Resource {
	if theme == "" {
		return nil
	} else if theme == "Maximize" {
		return wmTheme.MaximizeIcon
	}
	return wmTheme.IconifyIcon
}

type testAppProvider struct {
	screens []*Screen
}

func (tap *testAppProvider) AvailableApps() []AppData {
	return nil
}

func (tap *testAppProvider) AvailableThemes() []string {
	return nil
}

func (tap *testAppProvider) FindAppFromName(appName string) AppData {
	return &testAppData{name: appName}
}

func (tap *testAppProvider) FindAppFromWinInfo(win Window) AppData {
	return &testAppData{}
}

func (tap *testAppProvider) FindAppsMatching(pattern string) []AppData {
	return nil
}

func (tap *testAppProvider) DefaultApps() []AppData {
	return nil
}

func TestDeskLayout_Layout(t *testing.T) {
	l := &deskLayout{}
	l.screens = &testScreensProvider{}
	l.backgrounds = append(l.backgrounds, canvas.NewRectangle(color.White))
	l.bar = canvas.NewRectangle(color.Black)
	l.widgets = canvas.NewRectangle(color.Black)
	deskSize := fyne.NewSize(2000, 1000)

	l.Layout([]fyne.CanvasObject{l.backgrounds[0], l.bar, l.widgets}, deskSize)

	assert.Equal(t, l.backgrounds[0].Size(), deskSize)
	assert.Equal(t, l.widgets.Position().X+l.widgets.Size().Width, deskSize.Width)
	assert.Equal(t, l.widgets.Size().Height, deskSize.Height)
	assert.Equal(t, l.bar.Size().Width, deskSize.Width)
	assert.Equal(t, l.bar.Position().Y+l.bar.Size().Height, deskSize.Height)
}

func TestScaleVars(t *testing.T) {
	l := &deskLayout{}
	l.screens = &testScreensProvider{}
	env := l.scaleVars(1.8)
	assert.Contains(t, env, "QT_SCALE_FACTOR=1.8")
	assert.Contains(t, env, "GDK_SCALE=2")
	assert.Contains(t, env, "ELM_SCALE=1.8")
}

func TestBackgroundChange(t *testing.T) {
	l := &deskLayout{}
	instance = l
	l.screens = &testScreensProvider{}
	l.settings = &testSettings{}
	l.backgrounds = append(l.backgrounds, newBackground())

	workingDir, err := os.Getwd()
	if err != nil {
		fyne.LogError("Could not get current working directory", err)
		t.FailNow()
	}
	assert.Equal(t, wmTheme.Background, l.backgrounds[0].(*canvas.Image).Resource)

	l.settings.(*testSettings).background = filepath.Join(workingDir, "testdata", "fyne.png")
	l.updateBackgrounds(l.Settings().Background())
	assert.Equal(t, l.settings.Background(), l.backgrounds[0].(*canvas.Image).File)
}
