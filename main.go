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

	if len(keys) < 3 {
		fmt.Printf("Did not receive enough keys for a valid reset sequence: %d\n", len(keys))
		return false
	}
	if keys[len(keys)-3] == 5705736 && keys[len(keys)-2] == evdev.KeyM && keys[len(keys)-1] == evdev.KeyM {
		fmt.Println("Home, Menu, Menu has been pressed")
		return true
	}
	fmt.Print("The key sequence is invalid: ")
	fmt.Printf("Keys %d, %d, %d \n", keys[len(keys)-3], keys[len(keys)-2], keys[len(keys)-1])
	return false
}

func pollIR() {
	ir, err := evdev.Open("/dev/input/event0")
	defer ir.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
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
			} else {
				fmt.Println("----------------------------")
				if startTime == 0 {
					startTime = time.Now().Unix()
				} else if int64(startTime+3) < time.Now().Unix() {
					startTime = 0
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
