package recfz

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/flipperdevices/go-flipper"
	"go.bug.st/serial"
)

const (
	flipperPid             = "5740"
	flipperVid             = "0483"
	startRPCSessionCommand = "start_rpc_session\r"
)

// Opts represents an optional configuration for the flipper zero.
type Opts func(f *FlipperZero)

func WithPort(port string) Opts {
	return func(f *FlipperZero) {
		f.port = port
	}
}

// WithContext sets the context for the flipper zero.
func WithContext(ctx context.Context) Opts {
	return func(f *FlipperZero) {
		f.parentCtx = ctx
	}
}

// WithStreamScreenCallback sets the callback for the screen stream.
func WithStreamScreenCallback(cb func(frame flipper.ScreenFrame)) Opts {
	return func(f *FlipperZero) {
		f.streamScreenCallback = cb
	}
}

// WithLogger sets the logger for the flipper zero.
func WithLogger(l *log.Logger) Opts {
	return func(f *FlipperZero) {
		f.logger = l
	}
}

// FlipperZero represents the flipper zero device.
type FlipperZero struct {
	parentCtx            context.Context
	ctx                  context.Context
	cancel               context.CancelFunc
	port                 string
	conn                 serial.Port
	flipper              *flipper.Flipper
	reconnCh             chan struct{}
	connecting           bool
	mu                   sync.Mutex
	staticPort           bool
	streamScreenCallback func(frame flipper.ScreenFrame)
	logger               *log.Logger
	isClosing            bool
}

// NewFlipperZero creates a new flipper zero device.
// If the port is not static, it will try to autodetect the flipper.
func NewFlipperZero(opts ...Opts) (*FlipperZero, error) {
	f := &FlipperZero{
		reconnCh:  make(chan struct{}),
		logger:    log.Default(),
		parentCtx: context.Background(),
	}
	for _, opt := range opts {
		opt(f)
	}
	f.ctx, f.cancel = context.WithCancel(f.parentCtx)

	if f.port == "" {
		p, err := f.autodetectFlipper()
		if err != nil {
			return nil, fmt.Errorf("could not autodetect flipper: %w", err)
		}
		f.port = p
	} else {
		f.staticPort = true
	}
	return f, nil
}

// Close closes the connection to the flipper zero.
func (f *FlipperZero) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.isClosing {
		return
	}
	f.isClosing = true
	f.cancel()
	close(f.reconnCh)
	f.conn.Close()
}

func (f *FlipperZero) getClosing() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.isClosing
}

// Connected returns true if the flipper zero is connected.
func (f *FlipperZero) Connected() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.flipper != nil
}

// SetFlipper can be used to set a flipper instance.
func (f *FlipperZero) SetFlipper(fz *flipper.Flipper) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flipper = fz
}

// SetConn sets a serial connection to the flipper zero.
func (f *FlipperZero) SetConn(c serial.Port) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.conn = c
}

func (f *FlipperZero) getConn() serial.Port {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.conn
}

// GetFlipper returns the flipper instance.
// If the flipper is not connected, it returns an error.
func (f *FlipperZero) GetFlipper() (*flipper.Flipper, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.flipper == nil {
		return nil, fmt.Errorf("flipper is not connected")
	}
	return f.flipper, nil
}
