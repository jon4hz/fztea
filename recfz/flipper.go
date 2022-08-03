package recfz

import (
	"errors"

	"github.com/flipperdevices/go-flipper"
)

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

func (f *FlipperZero) SendShortPress(event flipper.InputKey) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypePress)   //nolint:errcheck
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypeShort)   //nolint:errcheck
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypeRelease) //nolint:errcheck
}

func (f *FlipperZero) SendLongPress(event flipper.InputKey) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypePress)   //nolint:errcheck
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypeLong)    //nolint:errcheck
	f.flipper.Gui.SendInputEvent(event, flipper.InputTypeRelease) //nolint:errcheck
}
