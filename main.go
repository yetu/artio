package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"github.com/jteeuwen/evdev"
)

func reset() {
	os.Remove("/mnt/stateful_partition/encrypted.block")
	os.Remove("/mnt/stateful_partition/encrypted.key")
	cmd := exec.Command("reboot")
	cmd.Run()
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
			case evdev.KeyH, evdev.KeyB:
				fmt.Println("On of the reset keys has been pressed")
				//reset()
			}
		}
	}
}

func main() {
	fmt.Printf("Listening for reset signal")
	pollIR()
}
