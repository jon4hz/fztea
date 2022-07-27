package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jon4hz/flipperzero-tea/flipperzero"
)

type model struct {
	flipper tea.Model
}

func main() {
	var port string
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	fz, err := flipperzero.NewFlipperZero(flipperzero.WithPort(port))
	if err != nil {
		log.Fatal(err)
	}
	m := model{
		flipper: flipperzero.New(fz),
	}
	if err := tea.NewProgram(m).Start(); err != nil {
		log.Fatalln(err)
	}
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
