package flipperzero

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flipperdevices/go-flipper"
)

const (
	fullBlock      = '█'
	upperHalfBlock = '▀'
	lowerHalfBlock = '▄'
)

type screenMsg string

var ErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))

type Model struct {
	Style    lipgloss.Style
	viewport viewport.Model
	ready    bool
	updates  chan string
	port     string
	fz       *FlipperZero
	err      error
	content  string
}

func New(fz *FlipperZero) tea.Model {
	m := &Model{
		Style:   lipgloss.NewStyle().Background(lipgloss.Color("#FF8C00")).Foreground(lipgloss.Color("#000000")),
		updates: make(chan string),
		fz:      fz,
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		m.fz.Flipper.Gui.StartScreenStream(m.updateScreen)
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
			key := mapKey(msg)
			if key != -1 {
				m.fz.Flipper.Gui.SendInputEvent(key, flipper.InputTypePress)
				m.fz.Flipper.Gui.SendInputEvent(key, flipper.InputTypeShort)
				m.fz.Flipper.Gui.SendInputEvent(key, flipper.InputTypeRelease)
			}
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.SetContent(m.Style.Render(m.content))
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}

	case screenMsg:
		m.content = string(msg)
		if m.ready {
			m.viewport.SetContent(m.Style.Render(m.content))
		}
		cmds = append(cmds, listenScreenUpdate(m.updates))
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
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
		s.WriteByte('\n')
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

func mapKey(key tea.KeyMsg) flipper.InputKey {
	switch key.Type {
	case tea.KeyUp:
		return flipper.InputKeyUp
	case tea.KeyDown:
		return flipper.InputKeyDown
	case tea.KeyRight:
		return flipper.InputKeyRight
	case tea.KeyLeft:
		return flipper.InputKeyLeft
	case tea.KeyEscape:
		return flipper.InputKeyBack
	case tea.KeyBackspace:
		return flipper.InputKeyBack
	case tea.KeyEnter, tea.KeySpace:
		return flipper.InputKeyOk
	}
	return -1
}
