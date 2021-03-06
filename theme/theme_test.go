package theme

import (
	"testing"

	"fyne.io/fyne"
	_ "fyne.io/fyne/test"
	"fyne.io/fyne/theme"

	"github.com/stretchr/testify/assert"
)

func TestIconResources(t *testing.T) {
	assert.NotNil(t, BatteryIcon.Name())
	assert.NotNil(t, BrightnessIcon.Name())
}

func TestIconTheme(t *testing.T) {
	fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
	battDark := BatteryIcon.Content()

	fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
	assert.NotEqual(t, battDark, BatteryIcon.Content())
}
