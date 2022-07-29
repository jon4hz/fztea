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
	"github.com/jon4hz/fztea/flipperzero"
	"github.com/muesli/coral"
)

var serverFlags struct {
	port           string
	listen         string
	authorizedKeys string
}

var serverCmd = &coral.Command{
	Use:   "server",
	Short: "Start an ssh server serving the flipper zero TUI",
	Run:   server,
}

func init() {
	serverCmd.Flags().StringVarP(&serverFlags.port, "port", "p", "", "port to connect to")
	serverCmd.Flags().StringVarP(&serverFlags.listen, "listen", "l", "127.0.0.1:2222", "address to listen on")
	serverCmd.Flags().StringVarP(&serverFlags.authorizedKeys, "authorized-keys", "k", "", "authorized_keys file for public key authentication")
}

func server(cmd *coral.Command, args []string) {
	fz, err := flipperzero.NewFlipperZero(flipperzero.WithPort(serverFlags.port))
	if err != nil {
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
					flipper: flipperzero.New(fz),
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
