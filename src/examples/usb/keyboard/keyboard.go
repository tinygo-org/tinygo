package main

import (
	"machine"
	"machine/usb"
	"time"
)

var keyboard = machine.HID0.Keyboard()

func main() {

	println("USB HID keyboard demo")

	for {
		time.Sleep(5 * time.Second)

		// Open a new text editor
		keyboard.Down(usb.KeyModifierAlt)
		keyboard.Press(usb.KeySpace)
		keyboard.Up(usb.KeyModifierAlt)
		time.Sleep(2 * time.Second)
		keyboard.Write([]byte("kate"))
		time.Sleep(time.Second)
		keyboard.Press(usb.KeyEnter)

		time.Sleep(5 * time.Second)

		// Use the io.Writer interface
		keyboard.Write([]byte("TinyGo USB Keyboard Control Test\n"))
		time.Sleep(2 * time.Second)

		// Or manually specify keycodes and Unicode codepoints
		testKeys([]Key{
			// Print alphabet out-of-order
			{Press: usb.KeyX},
			{Press: usb.KeyY},
			{Press: usb.KeyZ},
			{Press: usb.KeyG},
			{Press: usb.KeyH},
			{Press: usb.KeyI},
			{Press: usb.KeyJ},
			{Press: usb.KeyK},
			{Press: usb.KeyL},
			{Press: usb.KeyM},
			{Press: usb.KeyN},
			{Press: usb.KeyO},
			{Press: usb.KeyP},
			{Press: usb.KeyQ},
			{Press: usb.KeyR},
			{Press: usb.KeyS},
			{Press: usb.KeyT},
			{Press: usb.KeyA},
			{Press: usb.KeyB},
			{Press: usb.KeyC},
			{Press: usb.KeyD},
			{Press: usb.KeyE},
			{Press: usb.KeyF},
			{Press: usb.KeyU},
			{Press: usb.KeyV},
			{Press: usb.KeyW},
			// Pause 1 second
			{Time: time.Second},
			// Move cursor left x3
			{Press: usb.KeyLeft},
			{Press: usb.KeyLeft},
			{Press: usb.KeyLeft},
			// Pause 1 second
			{Time: time.Second},
			// Highlight 6 symbols to the left
			{Down: usb.KeyModifierShift},
			{Press: usb.KeyLeft},
			{Press: usb.KeyLeft},
			{Press: usb.KeyLeft},
			{Press: usb.KeyLeft},
			{Press: usb.KeyLeft},
			{Press: usb.KeyLeft},
			{Up: usb.KeyModifierShift},
			// Pause 1 second
			{Time: time.Second},
			// Use Ctrl-X to cut
			{Down: usb.KeyModifierCtrl, Press: usb.KeyX, Up: usb.KeyModifierCtrl},
			// Pause 1 second
			{Time: time.Second},
			// Move to beginning of line
			{Press: usb.KeyHome},
			// Pause 1 second
			{Time: time.Second},
			// Use Ctrl-V to paste
			{Down: usb.KeyModifierCtrl, Press: usb.KeyV, Up: usb.KeyModifierCtrl},
			// Pause 1 second
			{Time: time.Second},
			// Highlight 3 symbols to the right
			{Down: usb.KeyModifierShift},
			{Press: usb.KeyRight},
			{Press: usb.KeyRight},
			{Press: usb.KeyRight},
			{Up: usb.KeyModifierShift},
			// Use Ctrl-X to cut
			{Down: usb.KeyModifierCtrl, Press: usb.KeyX, Up: usb.KeyModifierCtrl},
			// Pause 1 second
			{Time: time.Second},
			// Move to end of line
			{Press: usb.KeyEnd},
			// Pause 1 second
			{Time: time.Second},
			// Use Ctrl-V to paste
			{Down: usb.KeyModifierCtrl, Press: usb.KeyV, Up: usb.KeyModifierCtrl},
			// Pause 1 second
			{Time: time.Second},
			// Newline
			{Press: usb.KeyEnter},
			{Press: usb.KeyEnter},
		}, 150*time.Millisecond)

		// Highlight all text and delete
		keyboard.Down(usb.KeyModifierCtrl)
		keyboard.Press(usb.KeyA)
		keyboard.Up(usb.KeyModifierCtrl)
		time.Sleep(time.Second)
		keyboard.Press(usb.KeyDelete)
		time.Sleep(time.Second)

		// Close window
		keyboard.Down(usb.KeyModifierCtrl)
		keyboard.Press(usb.KeyQ)
		keyboard.Up(usb.KeyModifierCtrl)
		time.Sleep(2 * time.Second)
		// Confirm discard file
		keyboard.Down(usb.KeyModifierAlt)
		keyboard.Press(usb.KeyD)
		keyboard.Up(usb.KeyModifierAlt)

		time.Sleep(5 * time.Second)

		// Open a new terminal
		keyboard.Down(usb.KeyModifierAlt)
		keyboard.Press(usb.KeySpace)
		keyboard.Up(usb.KeyModifierAlt)
		time.Sleep(2 * time.Second)
		keyboard.Write([]byte("konsole"))
		keyboard.Press(usb.KeyEnter)

		time.Sleep(5 * time.Second)

		// Open serial connection
		keyboard.Write([]byte("screen /dev/ttyACM0 115200"))
		time.Sleep(2 * time.Second)
		keyboard.Press(usb.KeyEnter)
		time.Sleep(4 * time.Second)

		// Write to UART
		keyboard.Write([]byte("hello!"))
		time.Sleep(time.Second)
		keyboard.Press(usb.KeyEnter)
		time.Sleep(2 * time.Second)
		keyboard.Write([]byte("NO U"))
		time.Sleep(time.Second)
		keyboard.Press(usb.KeyEnter)
		time.Sleep(2 * time.Second)

		// Close serial connection
		keyboard.Down(usb.KeyModifierCtrl)
		keyboard.Press(usb.KeyX)
		keyboard.Up(usb.KeyModifierCtrl)
		time.Sleep(time.Second)
		keyboard.Press(usb.KeyBackslash)
		time.Sleep(time.Second)
		keyboard.Press(usb.KeyY)
		keyboard.Press(usb.KeyEnter)
		time.Sleep(2 * time.Second)

		// Close terminal
		keyboard.Down(usb.KeyModifierCtrl)
		keyboard.Press(usb.KeyD)
		keyboard.Up(usb.KeyModifierCtrl)

		time.Sleep(25 * time.Second)
	}
}

type Key struct {
	Press usb.Keycode
	Down  usb.Keycode
	Up    usb.Keycode
	Time  time.Duration
}

func testKeys(key []Key, delay time.Duration) {
	for _, k := range key {
		if 0 != k.Down {
			keyboard.Down(k.Down)
		}
		if 0 != k.Press {
			keyboard.Press(k.Press)
		}
		if 0 != k.Up {
			keyboard.Up(k.Up)
		}
		if 0 != k.Time {
			time.Sleep(k.Time)
		} else {
			time.Sleep(delay)
		}
	}
}

func testInternationalLayout() {
	// International keyboard layouts also supported
	keyboard.Write([]byte("TinyGo USB Keyboard Layout Test\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Lowercase:  abcdefghijklmnopqrstuvwxyz\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Uppercase:  ABCDEFGHIJKLMNOPQRSTUVWXYZ\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Numbers:    0123456789\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Symbols1:   !\"#$%&'()*+,-./\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Symbols2:   :;<=>?[\\]^_`{|}~\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Symbols3:   ¡¢£¤¥¦§¨©ª«¬­®¯°±\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Symbols4:   ²³´µ¶·¸¹º»¼½¾¿×÷\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Grave:      ÀÈÌÒÙàèìòù\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Acute:      ÁÉÍÓÚÝáéíóúý\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Circumflex: ÂÊÎÔÛâêîôû\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Tilde:      ÃÑÕãñõ\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Diaeresis:  ÄËÏÖÜäëïöüÿ\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Cedilla:    Çç\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Ring Above: Åå\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("AE:         Ææ\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Thorn:      Þþ\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Sharp S:    ß\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("O-Stroke:   Øø\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Eth:        Ðð\n"))
	time.Sleep(250 * time.Millisecond)
	keyboard.Write([]byte("Euro:       €\n"))
}
