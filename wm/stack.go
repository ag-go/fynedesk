package wm

import "github.com/fyne-io/desktop"

type stack struct {
	frames []desktop.Window
}

func (s *stack) addToStack(win desktop.Window) {
	s.frames = append([]desktop.Window{win}, s.frames...)
}

func (s *stack) removeFromStack(win desktop.Window) {
	pos := -1
	for i, w := range s.frames {
		if w == win {
			pos = i
		}
	}

	if pos == -1 {
		return
	}
	s.frames = append(s.frames[:pos], s.frames[pos+1:]...)
}

func (s *stack) AddWindow(win desktop.Window) {
	if win == nil {
		return
	}
	s.addToStack(win)
}

func (s *stack) RemoveWindow(win desktop.Window) {
	s.removeFromStack(win)

	if s.TopWindow() != nil {
		s.TopWindow().Focus()
	}
}

func (s *stack) TopWindow() desktop.Window {
	if len(s.frames) == 0 {
		return nil
	}

	return s.frames[0]
}

func (s *stack) Windows() []desktop.Window {
	return s.frames
}

func (s *stack) RaiseToTop(win desktop.Window) {
	win.(*frame).raiseToTop()

	s.removeFromStack(win)
	s.addToStack(win)
}