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

func isValidResetSequence(keys []uint16) bool {

	if len(keys) != 3 {
		fmt.Println("Did not receive enough keys for a valid reset sequence")
		return false
	}
	if keys[0] == evdev.KeyH && keys[1] == evdev.KeyM && keys[2] == evdev.KeyM {
		fmt.Println("Home, Menu, Menu has been pressed")
		return true
	}
	fmt.Println("The key sequence is invalid")
	return false
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
	receivedKeys := make([]uint16, 3, 3)
	for {
		select {
		case <-signals:
			return
		case evt := <-ir.Inbox:
			if evt.Type != evdev.EvKeys {
				// Not a key event
				fmt.Println("Received not a key event")
				fmt.Printf("%+v\n", evt)
			} else {
				switch evt.Code {
				case evdev.KeyH, evdev.KeyM:
					fmt.Println("On of the reset keys has been pressed")
					if startTime == 0 {
						fmt.Println("Setting start time, so the sequence has to be entered within 3 seconds")
						startTime = time.Now().Unix()
					}
					if startTime != 0 && int64(startTime+3) > time.Now().Unix() {
						fmt.Println("Resetting starttime and received key slice since the user waited more than 3 seconds")
						startTime = 0
						receivedKeys = nil
					} else {
						fmt.Println("The key event is used and added to the buffer slice")
						receivedKeys = append(receivedKeys, evt.Code)
						if isValidResetSequence(receivedKeys) {
							fmt.Println("The received key sequence is valid, initiating reset")
							reset()
							return
						}
					}
				}
			}
		}
	}
}

func main() {
	fmt.Printf("Listening for reset signal")
	pollIR()
}
