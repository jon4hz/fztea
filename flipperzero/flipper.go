package flipperzero

import (
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flipperdevices/go-flipper"
)

const (
	fullBlock      = '█'
	upperHalfBlock = '▀'
	lowerHalfBlock = '▄'

	flipperScreenHeight = 32
	flipperScreenWidth  = 128

	fzEventCoolDown = time.Millisecond * 10
)

type screenMsg string

var ErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))

type Model struct {
	Style       lipgloss.Style
	viewport    viewport.Model
	updates     chan string
	fz          *FlipperZero
	err         error
	content     string
	lastFZEvent time.Time
	mu          *sync.Mutex
}

func New(fz *FlipperZero) tea.Model {
	m := &Model{
		Style:       lipgloss.NewStyle().Background(lipgloss.Color("#FF8C00")).Foreground(lipgloss.Color("#000000")),
		updates:     make(chan string),
		fz:          fz,
		viewport:    viewport.New(flipperScreenWidth, flipperScreenHeight),
		lastFZEvent: time.Now().Add(-fzEventCoolDown),
		mu:          &sync.Mutex{},
	}
	m.viewport.MouseWheelEnabled = false

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		m.fz.Flipper.Gui.StartScreenStream(m.updateScreen) //nolint:errcheck
		return nil
	}, listenScreenUpdate(m.updates))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return nil, tea.Quit
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

	case screenMsg:
		m.content = string(msg)
		m.viewport.SetContent(m.Style.Render(m.content))
		cmds = append(cmds, listenScreenUpdate(m.updates))
	}

	return m, tea.Batch(cmds...)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *Model) sendFlipperEvent(event flipper.InputKey, isLong bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if time.Since(m.lastFZEvent) < fzEventCoolDown {
		return
	}
	if !isLong {
		m.fz.Flipper.Gui.SendInputEvent(event, flipper.InputTypePress)   //nolint:errcheck
		m.fz.Flipper.Gui.SendInputEvent(event, flipper.InputTypeShort)   //nolint:errcheck
		m.fz.Flipper.Gui.SendInputEvent(event, flipper.InputTypeRelease) //nolint:errcheck
	} else {
		m.fz.Flipper.Gui.SendInputEvent(event, flipper.InputTypePress)   //nolint:errcheck
		m.fz.Flipper.Gui.SendInputEvent(event, flipper.InputTypeLong)    //nolint:errcheck
		m.fz.Flipper.Gui.SendInputEvent(event, flipper.InputTypeRelease) //nolint:errcheck
	}
	m.lastFZEvent = time.Now()
}

func (m Model) View() string {
	if m.err != nil {
		return ErrStyle.Render(m.err.Error())
	}
	return m.viewport.View()
}

func (m Model) updateScreen(frame flipper.ScreenFrame) {
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
		m.updates <- s.String()
	}()
}

func listenScreenUpdate(u <-chan string) tea.Cmd {
	return func() tea.Msg {
		return screenMsg(<-u)
	}
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
