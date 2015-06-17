package main

import "fmt"
import "github.com/gvalkov/golang-evdev"
import "os"

func reset() {
	os.Remove("/mnt/stateful_partition/encrypted.block")
	os.Remove("/mnt/stateful_partition/encrypted.key")
}

func pollIR() {
	resetCodes := map[int]bool{
		evdev.KEY_H: true,
		evdev.KEY_B: true,
	}
	ir, err := evdev.Open("/dev/input/event0")
	resetEvents := make([]KeyEvent, 3, 3)
	for {
		event, err := evdev.ReadOne()
		if event.Type == evdev.ByEventType.Key {
			// We have a key event.
			keyEvent = evdev.NewKeyEvent(event)
			if resetCodes[keyEvent.Keycode] {
				fmt.Printf("Found reset keycode\n")
				fmt.Printf(keyEvent.String())
			}
		}
	}
}

func main() {
	fmt.Printf("Listening for reset signal")
	pollIR()
}
