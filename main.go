package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/jteeuwen/evdev"
)

func reset() {
	os.Remove("/mnt/stateful_partition/encrypted.block")
	os.Remove("/mnt/stateful_partition/encrypted.key")
	cmd := exec.Command("reboot")
	cmd.Run()
}

func isValidResetSequence(keys []int32) bool {

	if len(keys) != 3 {
		fmt.Println("Did not receive enough keys for a valid reset sequence")
		return false
	}
	if keys[0] == 5705736 && keys[1] == evdev.KeyM && keys[2] == evdev.KeyM {
		fmt.Println("Home, Menu, Menu has been pressed")
		return true
	}
	fmt.Println("The key sequence is invalid")
	return false
}

func eventToString(event evdev.Event) string {
	var name string

	switch event.Type {
	case evdev.EvSync:
		name = "Sync Events"
	case evdev.EvKeys:
		name = "Keys or Buttons"
	case evdev.EvRelative:
		name = "Relative Axes"
	case evdev.EvAbsolute:
		name = "Absolute Axes"
	case evdev.EvMisc:
		name = "Miscellaneous"
	case evdev.EvLed:
		name = "LEDs"
	case evdev.EvSound:
		name = "Sounds"
	case evdev.EvRepeat:
		name = "Repeat"
	case evdev.EvForceFeedback,
		evdev.EvForceFeedbackStatus:
		name = "Force Feedback"
	case evdev.EvPower:
		name = "Power Management"
	case evdev.EvSwitch:
		name = "Binary switches"
	default:
		name = fmt.Sprintf("Unknown (0x%02x)", event.Type)
	}
	return name
}

func pollIR() {
	ir, err := evdev.Open("/dev/input/event0")
	defer ir.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	//resetEvents := make([]int, 3, 3)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	var startTime int64 = 0
	receivedKeys := make([]int32, 3, 3)
	for {
		select {
		case <-signals:
			return
		case evt := <-ir.Inbox:
			if evt.Type != evdev.EvKeys && evt.Type != evdev.EvMisc {
				// Not a key event
				fmt.Print("Received not a relevant event: ")
				fmt.Println(eventToString(evt))
				fmt.Printf("%+v\n", evt)
			} else {
				fmt.Println("----------------------------")
				if startTime == 0 {
					fmt.Println("Setting start time, so the sequence has to be entered within 3 seconds")
					startTime = time.Now().Unix()
				}
				if int64(startTime+3) > time.Now().Unix() {
					fmt.Println("Resetting starttime and received key slice since the user waited more than 3 seconds")
					startTime = time.Now().Unix()
					receivedKeys = nil
				}
				var value int32 = 0
				if evt.Type == evdev.EvKeys {
					value = int32(evt.Code)
				} else if evt.Type == evdev.EvMisc {
					value = evt.Value
				}
				receivedKeys = append(receivedKeys, value)
				if isValidResetSequence(receivedKeys) {
					fmt.Println("The received key sequence is valid, initiating reset")
					reset()
					return
				}
			}
		}
	}
}

func main() {
	fmt.Printf("Listening for reset signal")
	pollIR()
}
