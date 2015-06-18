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
		return false
	}
	if keys[0] == evdev.KeyH && keys[1] == evdev.KeyM && keys[2] == evdev.KeyM {
		return true
	}

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
				return
			}
			switch evt.Code {
			case evdev.KeyH, evdev.KeyM:
				if startTime == 0 {
					startTime = time.Now().Unix()
				}
				if startTime != 0 && int64(startTime+3) > time.Now().Unix() {
					startTime = 0
					receivedKeys = nil
				}
				fmt.Println("On of the reset keys has been pressed")
				receivedKeys = append(receivedKeys, evt.Code)
				if isValidResetSequence(receivedKeys) {
					reset()
					break
				}
			}
		}
	}
}

func main() {
	fmt.Printf("Listening for reset signal")
	pollIR()
}
