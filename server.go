package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
	"github.com/jon4hz/flipperzero-tea/flipperzero"
	"github.com/muesli/coral"
)

var serverFlags struct {
	port   string
	listen string
}

var serverCmd = &coral.Command{
	Use:   "server",
	Short: "Start an ssh server serving the flipper zero TUI",
	Run:   server,
}

func init() {
	serverCmd.Flags().StringVarP(&serverFlags.port, "port", "p", "", "port to connect to")
	serverCmd.Flags().StringVarP(&serverFlags.listen, "listen", "l", "127.0.0.1:2222", "address to listen on")
}

func server(cmd *coral.Command, args []string) {
	fz, err := flipperzero.NewFlipperZero(flipperzero.WithPort(serverFlags.port))
	if err != nil {
		log.Fatal(err)
	}
	s, err := wish.NewServer(
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
					flipper: flipperzero.New(fz),
				}
				return m, []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
			}),
			lm.Middleware(),
		),
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
