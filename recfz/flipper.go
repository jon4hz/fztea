package recfz

import (
	"errors"

	"github.com/flipperdevices/go-flipper"
)

// startScreenStream starts a screen stream from the flipper zero device.
// It triggers a callback function for every new screen frame.
func (f *FlipperZero) startScreenStream() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.streamScreenCallback == nil {
		return errors.New("no stream screen callback set")
	}
	if err := f.flipper.Gui.StartScreenStream(f.streamScreenCallback); err != nil {
		return err
	}
	f.logger.Println("started screen streaming...")
	return nil
}

// SendShortPress sends a short press event to the flipper zero device.
// If the flipper zero device is not connected, it will do nothing.
func (f *FlipperZero) SendShortPress(event flipper.InputKey) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.flipper == nil {
		return
	}
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypePress)   //nolint:errcheck
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypeShort)   //nolint:errcheck
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypeRelease) //nolint:errcheck
}

// SendLongPress sends a long press event to the flipper zero device.
// If the flipper zero device is not connected, it will do nothing.
func (f *FlipperZero) SendLongPress(event flipper.InputKey) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.flipper == nil {
		return
	}
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypePress)   //nolint:errcheck
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypeLong)    //nolint:errcheck
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypeRelease) //nolint:errcheck
}
