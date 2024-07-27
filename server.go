package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/jon4hz/fztea/flipperui"
	"github.com/jon4hz/fztea/recfz"
	"github.com/muesli/coral"
)

var serverFlags struct {
	listen         string
	authorizedKeys string
}

var serverCmd = &coral.Command{
	Use:   "server",
	Short: "Start an ssh server serving the flipper zero TUI",
	Run:   server,
}

func init() {
	serverCmd.Flags().StringVarP(&serverFlags.listen, "listen", "l", "127.0.0.1:2222", "address to listen on")
	serverCmd.Flags().StringVarP(&serverFlags.authorizedKeys, "authorized-keys", "k", "", "authorized_keys file for public key authentication")
}

func server(cmd *coral.Command, _ []string) {
	// parse screenshot resolution
	screenshotResolution, err := parseScreenshotResolution()
	if err != nil {
		log.Fatalf("failed to parse screenshot resolution: %s", err)
	}

	screenUpdates := make(chan flipperui.ScreenMsg)
	fz, err := recfz.NewFlipperZero(
		recfz.WithPort(rootFlags.port),
		recfz.WithStreamScreenCallback(flipperui.UpdateScreen(screenUpdates)),
		recfz.WithContext(cmd.Context()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer fz.Close()
	if err := fz.Connect(); err != nil {
		log.Fatal(err)
	}

	cl := newConnLimiter(1)

	sshOpts := []ssh.Option{
		wish.WithAddress(serverFlags.listen),
		wish.WithHostKeyPath(".ssh/flipperzero_tea_ed25519"),
		wish.WithMiddleware(
			bm.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				_, _, active := s.Pty()
				if !active {
					wish.Fatalln(s, "no active terminal, skipping")
					return nil, nil
				}
				m := model{
					flipper: flipperui.New(fz, screenUpdates,
						flipperui.WithScreenshotResolution(screenshotResolution.width, screenshotResolution.height),
						flipperui.WithFgColor(rootFlags.fgColor),
						flipperui.WithBgColor(rootFlags.bgColor)),
				}
				return m, []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
			}),
			lm.Middleware(),
			connLimit(cl),
		),
	}

	if serverFlags.authorizedKeys != "" {
		sshOpts = append(sshOpts, wish.WithAuthorizedKeys(serverFlags.authorizedKeys))
	}

	s, err := wish.NewServer(
		sshOpts...,
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s", serverFlags.listen)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
