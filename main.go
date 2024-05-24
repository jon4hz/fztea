package main

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jon4hz/fztea/flipperui"
	"github.com/jon4hz/fztea/internal/version"
	"github.com/jon4hz/fztea/recfz"
	"github.com/muesli/coral"
	mcoral "github.com/muesli/mango-coral"
	"github.com/muesli/roff"
)

var rootFlags struct {
	port                 string
	screenshotResolution string
	fgColor              string
	bgColor              string
}

var rootCmd = &coral.Command{
	Use:     "fztea",
	Short:   "TUI to interact with your flipper zero",
	Version: version.Version,
	Run:     root,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootFlags.port, "port", "p", "", "serial port to connect to (default: auto-detected)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.screenshotResolution, "screenshot-resolution", "1024x512", "screenshot resolution")
	rootCmd.PersistentFlags().StringVar(&rootFlags.fgColor, "fg-color", "#000000", "foreground color")
	rootCmd.PersistentFlags().StringVar(&rootFlags.bgColor, "bg-color", "#FF8C00", "background color")

	//rootCmd.AddCommand(serverCmd, versionCmd, manCmd)
	rootCmd.AddCommand(versionCmd, manCmd)
}

func root(cmd *coral.Command, args []string) {
	// parse screenshot resolution
	screenshotResolution, err := parseScreenshotResolution()
	if err != nil {
		log.Fatalf("failed to parse screenshot resolution: %s", err)
	}

	screenUpdates := make(chan flipperui.ScreenMsg)
	fz, err := recfz.NewFlipperZero(
		recfz.WithContext(cmd.Context()),
		recfz.WithPort(rootFlags.port),
		recfz.WithStreamScreenCallback(flipperui.UpdateScreen(screenUpdates)),
		recfz.WithLogger(log.New(io.Discard, "", 0)),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer fz.Close()

	if err := fz.Connect(); err != nil {
		log.Fatal(err)
	}
	m := model{
		flipper: flipperui.New(fz, screenUpdates,
			flipperui.WithScreenshotResolution(screenshotResolution.width, screenshotResolution.height),
			flipperui.WithFgColor(rootFlags.fgColor),
			flipperui.WithBgColor(rootFlags.bgColor),
		),
	}
	if _, err := tea.NewProgram(m, tea.WithMouseCellMotion()).Run(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var manCmd = &coral.Command{
	Use:                   "man",
	Short:                 "generates the manpages",
	SilenceUsage:          true,
	DisableFlagsInUseLine: true,
	Hidden:                true,
	Args:                  coral.NoArgs,
	RunE: func(cmd *coral.Command, args []string) error {
		manPage, err := mcoral.NewManPage(1, rootCmd)
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(os.Stdout, manPage.Build(roff.NewDocument()))
		return err
	},
}

var versionCmd = &coral.Command{
	Use:   "version",
	Short: "Print the version info",
	Run: func(cmd *coral.Command, args []string) {
		fmt.Printf("Version: %s\n", version.Version)
		fmt.Printf("Commit: %s\n", version.Commit)
		fmt.Printf("Date: %s\n", version.Date)
		fmt.Printf("Build by: %s\n", version.BuiltBy)
	},
}

func parseScreenshotResolution() (struct {
	width  int
	height int
}, error) {
	var screenshotResolution struct {
		width  int
		height int
	}
	_, err := fmt.Sscanf(rootFlags.screenshotResolution, "%dx%d", &screenshotResolution.width, &screenshotResolution.height)
	return screenshotResolution, err
}
