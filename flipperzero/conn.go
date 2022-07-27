package flipperzero

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/flipperdevices/go-flipper"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

const (
	flipperPid             = "5740"
	flipperVid             = "0483"
	startRpcSessionCommand = "start_rpc_session\r"
)

type Opts func(f *FlipperZero)

func WithPort(port string) Opts {
	return func(f *FlipperZero) {
		f.port = port
	}
}

type FlipperZero struct {
	port    string
	Flipper *flipper.Flipper
}

func NewFlipperZero(opts ...Opts) (*FlipperZero, error) {
	f := &FlipperZero{}
	for _, opt := range opts {
		opt(f)
	}

	if f.port == "" {
		p, err := autodetectFlipper()
		if err != nil {
			return nil, fmt.Errorf("could not autodetect flipper: %w", err)
		}
		f.port = p
	}
	conn, err := newConn(f.port)
	if err != nil {
		return nil, fmt.Errorf("could not open serial conn: %w", err)
	}
	fz, err := flipper.ConnectWithTimeout(conn, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("could not connect to flipper: %w", err)
	}
	f.Flipper = fz

	return f, nil
}

func autodetectFlipper() (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", err
	}

	for _, p := range ports {
		if p.PID == flipperPid && p.VID == flipperVid {
			return p.Name, nil
		}
	}
	return "", errors.New("no flipper found")
}

func newConn(port string) (io.ReadWriter, error) {
	ser, err := serial.Open(port, &serial.Mode{})
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(ser)
	_, err = readUntil(br, []byte("\r\n\r\n>: "))
	if err != nil {
		return nil, err
	}
	_, err = ser.Write([]byte(startRpcSessionCommand))
	if err != nil {
		return nil, err
	}
	token, err := br.ReadString('\r')
	if err != nil {
		return nil, err
	}
	if token != startRpcSessionCommand {
		return nil, errors.New(strings.TrimSpace(token))
	}
	return ser, nil
}

type reader interface {
	ReadString(delim byte) (line string, err error)
}

func readUntil(r reader, delim []byte) (line []byte, err error) {
	for {
		s := ""
		s, err = r.ReadString(delim[len(delim)-1])
		if err != nil {
			return
		}

		line = append(line, []byte(s)...)
		if bytes.HasSuffix([]byte(s), delim) {
			return line[:len(line)-len(delim)], nil
		}
	}
}
