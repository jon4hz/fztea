package main

import (
	"github.com/flipperdevices/go-flipper"
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
	"strings"
)

var s tcell.Screen
var flipperStyle = tcell.StyleDefault.Background(tcell.ColorDarkOrange).Foreground(tcell.ColorBlack)

func main() {
	var err error
	s, err = tcell.NewScreen()
	if err != nil {
		log.Fatalln("Can't create screen", err)
	}
	if err = s.Init(); err != nil {
		log.Fatalln("Can't init screen", err)
	}

	var port string
	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	if port == "" {
		ports, err := findFlippers()
		if err != nil {
			log.Fatalln("Can't enumerate ports. Please specify manually")
		}
		if len(ports) == 0 {
			log.Fatalln("No Flippers detected . Try specifying manually")
		}
		if len(ports) > 1 {
			log.Fatalf("Multiple Flippers detected: (%s). Please specify manually\n",
				strings.Join(ports, ", "))
		}
		port = ports[0]
	}

	ser, err := initCli(port)
	if err != nil {
		log.Fatalln("Can't init RPC", err)
	}

	f, err := flipper.Connect(ser)
	if err != nil {
		log.Fatalln("Can't connect to RPC", err)
	}

	err = f.Gui.StartScreenStream(renderFrame)
	if err != nil {
		log.Fatalln("Can't start screen stream", err)
	}

	for {
		ev := s.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC {
				s.Fini()
				os.Exit(0)
			}

			key := mapKey(ev)
			if key == -1 {
				continue
			}

			f.Gui.SendInputEvent(key, flipper.InputTypePress)
			f.Gui.SendInputEvent(key, flipper.InputTypeShort)
			f.Gui.SendInputEvent(key, flipper.InputTypeRelease)
		}
	}
}

func renderFrame(frame flipper.ScreenStreamFrame) {
	s.Clear()

	for y := 0; y < 64; y += 2 {
		for x := 0; x < 128; x++ {
			r := tcell.RuneBlock
			if !frame.IsPixelSet(x, y) && frame.IsPixelSet(x, y+1) {
				r = '\u2584'
			}
			if frame.IsPixelSet(x, y) && !frame.IsPixelSet(x, y+1) {
				r = '\u2580'
			}
			if !frame.IsPixelSet(x, y) && !frame.IsPixelSet(x, y+1) {
				r = ' '
			}
			s.SetContent(x, y/2, r, nil, flipperStyle)
		}
	}

	s.Show()
}

func mapKey(ev *tcell.EventKey) flipper.InputKey {
	switch ev.Key() {
	case tcell.KeyUp:
		return flipper.InputKeyUp
	case tcell.KeyDown:
		return flipper.InputKeyDown
	case tcell.KeyRight:
		return flipper.InputKeyRight
	case tcell.KeyLeft:
		return flipper.InputKeyLeft
	case tcell.KeyEscape:
		return flipper.InputKeyBack
	case tcell.KeyBackspace:
		return flipper.InputKeyBack
	case tcell.KeyBackspace2:
		return flipper.InputKeyBack
	case tcell.KeyEnter:
		return flipper.InputKeyOk
	case tcell.KeyRune:
		if ev.Rune() == ' ' {
			return flipper.InputKeyOk
		}
	}
	return -1
}
