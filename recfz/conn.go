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

// Connect connects to the flipper zero device.
// It will indefinitely try to reconnect if the connection is lost.
func (f *FlipperZero) Connect() error {
	if err := f.reconnect(); err != nil {
		return err
	}
	go f.reconnLoop()
	return nil
}

// reconnect starts a new connection to the flipper zero device.
func (f *FlipperZero) reconnect() error {
	conn, err := f.newConn()
	if err != nil {
		return fmt.Errorf("could not open serial conn: %w", err)
	}
	fz, err := flipper.ConnectWithTimeout(conn, 10*time.Second)
	if err != nil {
		return fmt.Errorf("could not connect to flipper: %w", err)
	}
	f.logger.Println("successfully connected to flipper")

	f.SetFlipper(fz)
	f.SetConn(conn)

	return f.startScreenStream()
}

// newConn opens a new serial connection to the flipper zero device.
// If the port is not static, it will try to autodetect the flipper zero device.
// If the connection is already open, it will be closed and a new one will be opened.
// If the connection is openend successfully, it will start an rpc session over serial.
func (f *FlipperZero) newConn() (serial.Port, error) {
	port := f.port
	if !f.staticPort {
		var err error
		port, err = f.autodetectFlipper()
		if err != nil {
			return nil, err
		}
	}
	if conn := f.getConn(); conn != nil {
		conn.Close()
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
	f.logger.Println("successfully opened serial connection to flipper")
	return ser, nil
}

// autodetectFlipper tries to automatically detect the flipper zero device.
func (f *FlipperZero) autodetectFlipper() (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", err
	}

	for _, p := range ports {
		if p.PID == flipperPid && p.VID == flipperVid {
			f.logger.Printf("found flipper on %s", p.Name)
			return p.Name, nil
		}
	}
	return "", errors.New("no flipper found")
}

// checkConnLoop checks if the connection is still alive by sending an empty message every 2 seconds.
// If the connection is lost, it will trigger a reconnect.
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

// reconnLoop tries to reconnect to the flipper zero device if the connection is lost.
// If a reconnect fails, it will indefinitely try again after 1 second.
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
