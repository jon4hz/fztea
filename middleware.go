package main

import (
	"errors"
	"sync"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

type connLimiter struct {
	sync.Mutex
	conns    int
	maxConns int
}

func newConnLimiter(maxConns int) *connLimiter {
	return &connLimiter{
		maxConns: maxConns,
	}
}

func (u *connLimiter) Add() error {
	u.Lock()
	defer u.Unlock()
	if u.conns >= u.maxConns {
		return errors.New("max connections reached")
	}
	u.conns++
	return nil
}

func (u *connLimiter) Remove() {
	u.Lock()
	defer u.Unlock()
	u.conns--
	if u.conns < 0 {
		u.conns = 0
	}
}

func connLimit(limiter *connLimiter) wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			if err := limiter.Add(); err != nil {
				wish.Fatalf(s, "max connections reached\n")
				return
			}
			sh(s)
			limiter.Remove()
		}
	}
}
