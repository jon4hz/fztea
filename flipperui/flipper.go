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
	fullBlock      = '█'
	upperHalfBlock = '▀'
	lowerHalfBlock = '▄'

	flipperScreenHeight = 32
	flipperScreenWidth  = 128

	fzEventCoolDown = time.Millisecond * 10

	colorOrange = lipgloss.Color("#FF8C00")
	colorWhite  = lipgloss.Color("#000000")
)

type (
	ScreenMsg struct {
		screen string
		image  image.Image
	}
)

var ErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))

type Model struct {
	Style         lipgloss.Style
	viewport      viewport.Model
	fz            *recfz.FlipperZero
	err           error
	errTime       time.Time
	content       string
	lastFZEvent   time.Time
	screenUpdate  <-chan ScreenMsg
	currentScreen image.Image
	mu            *sync.Mutex
}

func New(fz *recfz.FlipperZero, screenUpdate <-chan ScreenMsg) tea.Model {
	m := &Model{
		Style:        lipgloss.NewStyle().Background(colorOrange).Foreground(colorWhite),
		fz:           fz,
		viewport:     viewport.New(flipperScreenWidth, flipperScreenHeight),
		lastFZEvent:  time.Now().Add(-fzEventCoolDown),
		screenUpdate: screenUpdate,
		mu:           &sync.Mutex{},
	}
	m.viewport.MouseWheelEnabled = false

	return m
}

func (m Model) Init() tea.Cmd {
	return listenScreenUpdate(m.screenUpdate)
}

func listenScreenUpdate(u <-chan ScreenMsg) tea.Cmd {
	return func() tea.Msg {
		return <-u
	}
}

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

func mapMouse(event tea.MouseMsg) flipper.InputKey {
	switch event.Type {
	case tea.MouseWheelUp:
		return flipper.InputKeyUp
	case tea.MouseWheelDown:
		return flipper.InputKeyDown
	}
	return -1
}

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

func (m Model) View() string {
	if m.err != nil && time.Since(m.errTime) < time.Second*4 {
		return ErrStyle.Render(fmt.Sprintf("%d %s", int((time.Second*4 - time.Since(m.errTime)).Seconds()), m.err))
	}
	return m.viewport.View()
}

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
		go func() {
			updates <- ScreenMsg{
				screen: s.String(),
				image:  frame.ToImage(colorWhite, colorOrange),
			}
		}()
	}
}

func (m *Model) saveImage() {
	resImg := imaging.Resize(m.currentScreen, 1024, 512, imaging.Box)

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

func (m *Model) setError(err error) {
	m.err = err
	m.errTime = time.Now()
}
