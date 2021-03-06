package fynedesk

// ScreenList provides information about available physical screens for Fyne desktop
type ScreenList interface {
	Screens() []*Screen                                            // Screens returns a Screen type slice of each available physical screen
	Active() *Screen                                               // Active returns the screen index of the currently active screen
	Primary() *Screen                                              // Primary returns the screen index of the primary screen
	Scale() float32                                                // Return the scale calculated for use across screens
	ScreenForWindow(Window) *Screen                                // Return the screen that a window is located on
	ScreenForGeometry(x int, y int, width int, height int) *Screen // Return the screen that a geometry is located on
}

// Screen provides relative information about a single physical screen
type Screen struct {
	Name                string // Name is the randr provided name of the screen
	X, Y, Width, Height int    // Geometry of the screen
}
