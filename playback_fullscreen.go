package subsonic

import (
	"time"

	"github.com/egoist/mygo/ui"
)

type playbackWindow interface {
	IsFullScreen() bool
	SetFullScreen(bool)
	OnEnterFullScreen(func()) func()
	OnLeaveFullScreen(func()) func()
}

type fullscreenTransition uint8

const (
	fullscreenIdle fullscreenTransition = iota
	fullscreenEntering
	fullscreenLeaving
	fullscreenTimeout = 3 * time.Second
	restoreIdleDelay  = 2 * time.Second
)

type playbackFullscreen struct {
	window playbackWindow
	stop   func()
	owned  bool
	target fullscreenTransition
	since  time.Time

	restoreUntil       time.Time
	pointerX, pointerY float32
	pointerNear        bool
}

func (f *playbackFullscreen) bind(window playbackWindow, redraw func()) {
	if f.window == window {
		return
	}
	if f.stop != nil {
		f.restore()
		f.stop()
	}
	*f = playbackFullscreen{window: window}
	offEnter := window.OnEnterFullScreen(func() {
		cancelled := f.target == fullscreenEntering && !f.owned
		f.target = fullscreenIdle
		if cancelled {
			f.request(fullscreenLeaving)
		}
		redraw()
	})
	offLeave := window.OnLeaveFullScreen(func() {
		f.owned, f.target = false, fullscreenIdle
		f.restoreUntil = time.Time{}
		redraw()
	})
	f.stop = func() {
		offEnter()
		offLeave()
	}
}

func (f *playbackFullscreen) request(target fullscreenTransition) {
	f.target, f.since = target, time.Now()
	f.window.SetFullScreen(target == fullscreenEntering)
}

func (f *playbackFullscreen) expire() {
	if f.target != fullscreenIdle && time.Since(f.since) >= fullscreenTimeout {
		f.target = fullscreenIdle
		if !f.window.IsFullScreen() {
			f.owned = false
		}
	}
}

// escape handles fullscreen before the caller dismisses Now Playing.
func (f *playbackFullscreen) escape() bool {
	if f.window == nil {
		return false
	}
	f.expire()
	switch f.target {
	case fullscreenEntering:
		// Some window managers animate entry. Finish the exit once they confirm it.
		f.owned = false
		return true
	case fullscreenLeaving:
		return true
	}
	if !f.window.IsFullScreen() {
		return false
	}
	f.owned = false
	f.request(fullscreenLeaving)
	return true
}

// restore only unwinds fullscreen that was entered through the playback view.
func (f *playbackFullscreen) restore() {
	f.restoreUntil = time.Time{}
	if f.owned {
		f.escape()
	}
}

// Only changes in pointer position renew the timer. Drawing lyrics or hovering
// in place must let the control disappear even when the playback view is busy.
func (f *playbackFullscreen) showRestore(c *ui.Context, zone ui.Element) bool {
	x, y, near := zone.PointerPosition()
	now := c.Now()
	if f.restoreUntil.IsZero() || near && (!f.pointerNear || x != f.pointerX || y != f.pointerY) {
		f.restoreUntil = now.Add(restoreIdleDelay)
	}
	f.pointerX, f.pointerY, f.pointerNear = x, y, near
	if left := f.restoreUntil.Sub(now); left > 0 {
		c.After(left)
		return true
	}
	return false
}

func (f *playbackFullscreen) button(c *ui.Context, w float32, pointerIn, canEnter bool) {
	if f.window == nil {
		return
	}
	f.expire()
	if f.target != fullscreenIdle {
		c.After(fullscreenTimeout - time.Since(f.since))
	}
	top, right := topBar(c), npSideX(w)
	full := f.window.IsFullScreen() || f.target == fullscreenEntering
	if full {
		// The upper-right region tracks motion even while the button is hidden.
		// Its button keeps the same offsets as in the windowed view.
		zone := ui.Box(c.Key("fullscreen-restore-zone")).Absolute().Top(0).Right(0).
			Size(right+100, top+npTabsHeight+24)
		if f.showRestore(c, zone) {
			zone.Children(func() { f.control(c, top, right, true, true) })
		}
		return
	}
	f.restoreUntil = time.Time{}
	if canEnter {
		f.control(c, top, right, pointerIn, false)
	}
}

func (f *playbackFullscreen) control(c *ui.Context, top, right float32, shown, full bool) {
	label, hint := "全屏", "进入全屏"
	if full {
		label, hint = "恢复", "恢复窗口（Esc）"
	}
	b := ui.ButtonBase(c.Key("playback-fullscreen")).Absolute().Top(top).Right(right).
		Size(54, npTabsHeight).Center().Radius(999).FocusRing(false).Label(label).Tooltip(hint)
	if full {
		fade := ui.Animation{Duration: 160 * time.Millisecond}
		b.Transition(ui.ElementTransition{Animation: &fade, Enter: &ui.Motion{}, Exit: &ui.Motion{}})
	}
	bg, fg := ui.Transparent, white.Alpha(0.55)
	if b.Hovered() {
		bg, fg = white.Alpha(0.06), white.Alpha(0.85)
	}
	b.Background(b.AnimateColor("bg", bg, ui.Animation{Duration: 120 * time.Millisecond}))
	reveal(b, shown).Children(func() {
		ui.Text(c, label).FontSize(13).FontWeight(600).TextColor(fg)
	})
	if b.Clicked() {
		if !f.escape() {
			f.owned = true
			f.request(fullscreenEntering)
		}
	}
}
