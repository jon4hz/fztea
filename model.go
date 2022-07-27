package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	flipper tea.Model
}

func (m model) Init() tea.Cmd {
	return m.flipper.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.flipper, cmd = m.flipper.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.flipper.View()
}
