package recfz

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/flipperdevices/go-flipper"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

func (f *FlipperZero) Connect() error {
	if err := f.reconnect(); err != nil {
		return err
	}
	go f.reconnLoop()
	return nil
}

func (f *FlipperZero) reconnect() error {
	conn, err := f.newConn()
	if err != nil {
		return fmt.Errorf("could not open serial conn: %w", err)
	}
	fz, err := flipper.ConnectWithTimeout(conn, 10*time.Second)
	if err != nil {
		return fmt.Errorf("could not connect to flipper: %w", err)
	}

	f.SetFlipper(fz)
	f.SetConn(conn)

	return f.startScreenStream()
}

func (f *FlipperZero) newConn() (serial.Port, error) {
	port := f.port
	if !f.staticPort {
		var err error
		port, err = autodetectFlipper()
		if err != nil {
			return nil, err
		}
	}
	ser, err := serial.Open(port, &serial.Mode{})
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(ser)
	_, err = readUntil(br, []byte("\r\n\r\n>: "))
	if err != nil {
		return nil, err
	}
	_, err = ser.Write([]byte(startRPCSessionCommand))
	if err != nil {
		return nil, err
	}

	token, err := br.ReadString('\r')
	if err != nil {
		return nil, err
	}
	if token != startRPCSessionCommand {
		return nil, errors.New(strings.TrimSpace(token))
	}

	go f.checkConnLoop(ser)

	return ser, nil
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

func (f *FlipperZero) checkConnLoop(r io.Writer) {
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	for {
		select {
		case <-f.ctx.Done():
			return
		case <-ticker.C:
			if !f.Connected() {
				continue
			}
			_, err := r.Write(nil)
			if err != nil {
				if f.getClosing() {
					return
				}
				f.logger.Printf("could not read from flipper: %s", err)
				f.reconnCh <- struct{}{}
				return
			}
		}
	}
}

func (f *FlipperZero) reconnLoop() {
	for {
		select {
		case _, ok := <-f.reconnCh:
			if !ok {
				return
			}
			if f.connecting || !f.Connected() {
				continue
			}
			f.connecting = true
			f.SetFlipper(nil)
			for {
				if err := f.reconnect(); err != nil {
					f.logger.Printf("could not reconnect: %v", err)
					time.Sleep(time.Second)
					continue
				}
				break
			}
			f.connecting = false
		case <-f.ctx.Done():
			return
		}
	}
}
