package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	flipper       tea.Model
	width, height int
}

// Init is the bubbletea init function.
func (m model) Init() tea.Cmd {
	return m.flipper.Init()
}

// Update is the bubbletea update funciton and handles all tea.Msgs.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	m.flipper, cmd = m.flipper.Update(msg)
	return m, cmd
}

// View is the bubbletea view function.
func (m model) View() string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.flipper.View())
}
