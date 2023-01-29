package flipperui

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/disintegration/imaging"
	"github.com/flipperdevices/go-flipper"
	"github.com/jon4hz/fztea/recfz"
)

const (
	// building blocks to draw the flipper screen in the terminal.
	fullBlock      = '█'
	upperHalfBlock = '▀'
	lowerHalfBlock = '▄'

	// screen size of the flipper
	flipperScreenHeight = 32
	flipperScreenWidth  = 128

	// fzEventCoolDown is the time that must pass between two events that are sent to the flipper.
	// That poor serial connection can handle only so much :(
	fzEventCoolDown = time.Millisecond * 10

	// some default colors
	colorBg = lipgloss.Color("#FF8C00")
	colorFg = lipgloss.Color("#000000")
)

type (
	// ScreenMsg is a message that is sent when the flipper sends a screen update.
	ScreenMsg struct {
		screen string
		image  image.Image
	}
)

// ErrStyle is the style of the error message
var ErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))

// Model represents the flipper model.
// It also implements the bubbletea.Model interface.
type Model struct {
	// Style is the style of the flipper screen
	Style lipgloss.Style
	// viewport is used to handle resizing easily
	viewport viewport.Model
	// fz is the flipper zero device
	fz *recfz.FlipperZero
	// err represents the last error that occurred. It will be displayed for a few seconds.
	err error
	// errTime is the time when the last error occurred
	errTime time.Time
	// content is the current screen of the flipper as a string
	content string
	// lastFZEvent is the time of the last event that was sent to the flipper.
	lastFZEvent time.Time
	// screenUpdate is a channel that receives screen updates from the flipper
	screenUpdate <-chan ScreenMsg
	// currentScreen is the last screen that was received from the flipper
	currentScreen image.Image
	// mutex to ensure that only one goroutine can send events to the flipper at a time
	mu *sync.Mutex
	// resolution of the screenshots
	screenshotResolution struct {
		width  int
		height int
	}
}

var _ tea.Model = (*Model)(nil)

// New constructs a new flipper model.
func New(fz *recfz.FlipperZero, screenUpdate <-chan ScreenMsg, opts ...FlipperOpts) tea.Model {
	m := Model{
		Style:        lipgloss.NewStyle().Background(colorBg).Foreground(colorFg),
		fz:           fz,
		viewport:     viewport.New(flipperScreenWidth, flipperScreenHeight),
		lastFZEvent:  time.Now().Add(-fzEventCoolDown),
		screenUpdate: screenUpdate,
		mu:           &sync.Mutex{},
		screenshotResolution: struct {
			width  int
			height int
		}{
			width:  1024,
			height: 512,
		},
	}
	m.viewport.MouseWheelEnabled = false

	for _, opt := range opts {
		opt(&m)
	}

	return &m
}

// Init is the bubbletea init function.
// the initial listenScreenUpdate command is started here.
func (m Model) Init() tea.Cmd {
	return listenScreenUpdate(m.screenUpdate)
}

// listenScreenUpdate listens for screen updates from the flipper and returns them as tea.Cmds.
func listenScreenUpdate(u <-chan ScreenMsg) tea.Cmd {
	return func() tea.Msg {
		return <-u
	}
}

// Update is the bubbletea update funciton and handles all tea.Msgs.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return nil, tea.Quit
		case tea.KeyCtrlS:
			m.saveImage()
			return m, nil
		default:
			key, getlong := mapKey(msg)
			if key != -1 {
				m.sendFlipperEvent(key, getlong)
			}
		}

	case tea.MouseMsg:
		event := mapMouse(msg)
		if event != -1 {
			m.sendFlipperEvent(event, false)
		}

	case tea.WindowSizeMsg:
		m.viewport.Width = min(msg.Width, flipperScreenWidth)
		m.viewport.Height = min(msg.Height, flipperScreenHeight)
		m.viewport.SetContent(m.Style.Render(m.content))

	case ScreenMsg:
		m.content = msg.screen
		m.currentScreen = msg.image
		m.viewport.SetContent(m.Style.Render(m.content))
		cmds = append(cmds, listenScreenUpdate(m.screenUpdate))
	}

	return m, tea.Batch(cmds...)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// mapKey maps a tea.KeyMsg to a flipper.InputKey
func mapKey(key tea.KeyMsg) (flipper.InputKey, bool) {
	switch key.String() {
	case "w", "up":
		return flipper.InputKeyUp, false
	case "a", "left":
		return flipper.InputKeyLeft, false
	case "s", "down":
		return flipper.InputKeyDown, false
	case "d", "right":
		return flipper.InputKeyRight, false
	case "o", "enter", " ":
		return flipper.InputKeyOk, false
	case "b", "backspace", "esc":
		return flipper.InputKeyBack, false
	case "W", "shift+up":
		return flipper.InputKeyUp, true
	case "A", "shift+left":
		return flipper.InputKeyLeft, true
	case "S", "shift+down":
		return flipper.InputKeyDown, true
	case "D", "shift+right":
		return flipper.InputKeyRight, true
	case "O":
		return flipper.InputKeyOk, true
	case "B":
		return flipper.InputKeyBack, true
	}
	return -1, false
}

// mapMouse maps a tea.MouseMsg to a flipper.InputKey
func mapMouse(event tea.MouseMsg) flipper.InputKey {
	switch event.Type {
	case tea.MouseWheelUp:
		return flipper.InputKeyUp
	case tea.MouseWheelDown:
		return flipper.InputKeyDown
	}
	return -1
}

// sendFlipperEvent sends an event to the flipper. It ensures that at most one event is sent every fzEventCoolDown.
func (m *Model) sendFlipperEvent(event flipper.InputKey, isLong bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if time.Since(m.lastFZEvent) < fzEventCoolDown {
		return
	}
	if !isLong {
		m.fz.SendShortPress(event)
	} else {
		m.fz.SendLongPress(event)
	}
	m.lastFZEvent = time.Now()
}

// View renders the flipper screen or an error message if there was an error.
func (m Model) View() string {
	if m.err != nil && time.Since(m.errTime) < time.Second*4 {
		return ErrStyle.Render(fmt.Sprintf("%d %s", int((time.Second*4 - time.Since(m.errTime)).Seconds()), m.err))
	}
	return m.viewport.View()
}

// UpdateScreen renders the terminal screen based on the flipper screen.
// It also returns the flipper screen as an image.
// This function is intended to be used as a callback for the flipper.
func UpdateScreen(updates chan<- ScreenMsg) func(frame flipper.ScreenFrame) {
	return func(frame flipper.ScreenFrame) {
		var s strings.Builder
		for y := 0; y < 64; y += 2 {
			var l strings.Builder
			for x := 0; x < 128; x++ {
				r := fullBlock
				if !frame.IsPixelSet(x, y) && frame.IsPixelSet(x, y+1) {
					r = lowerHalfBlock
				}
				if frame.IsPixelSet(x, y) && !frame.IsPixelSet(x, y+1) {
					r = upperHalfBlock
				}
				if !frame.IsPixelSet(x, y) && !frame.IsPixelSet(x, y+1) {
					r = ' '
				}
				l.WriteRune(r)
			}
			s.WriteString(l.String())

			// if not last line
			if y < 62 {
				s.WriteRune('\n')
			}
		}
		// make sure we don't block
		go func() {
			updates <- ScreenMsg{
				screen: s.String(),
				image:  frame.ToImage(colorFg, colorBg),
			}
		}()
	}
}

// saveImage saves the current screen as a png image.
func (m *Model) saveImage() {
	resImg := imaging.Resize(m.currentScreen, m.screenshotResolution.width, m.screenshotResolution.height, imaging.Box)

	out, err := os.Create(fmt.Sprintf("flipper_%s.png", time.Now().Format("20060102150405")))
	if err != nil {
		m.setError(err)
		return
	}
	defer out.Close()

	if err := png.Encode(out, resImg); err != nil {
		m.setError(err)
	}
}

// setError sets the error message and the time when it occurred.
func (m *Model) setError(err error) {
	m.err = err
	m.errTime = time.Now()
}
