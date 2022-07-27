package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jon4hz/flipperzero-tea/flipperzero"
	"github.com/muesli/coral"
)

var rootFlags struct {
	port string
}

var rootCmd = &coral.Command{
	Use:   "flipperzero-tea",
	Short: "TUI to interact with your flipper zero",
	Run:   root,
}

func init() {
	rootCmd.Flags().StringVarP(&rootFlags.port, "port", "p", "", "port to connect to")
	rootCmd.AddCommand(serverCmd)
}

func root(cmd *coral.Command, args []string) {
	fz, err := flipperzero.NewFlipperZero(flipperzero.WithPort(rootFlags.port))
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
