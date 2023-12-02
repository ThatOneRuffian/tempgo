package main

import (
	"fmt"
	"os"
	"strings"

	"tempgo/tempo"
	"tempgo/util"
)

func main() {
	var currentCapDevice tempo.BpmCaptureDevice
	var selectedCaptureIndex int = -1

	currentCapDevice.FirstRun = true
	availableDevices, enumErr := tempo.GetInputDevices()

	if enumErr != nil {
		fmt.Println("Could Not Enumerate Input Devices.")
		os.Exit(0)
	}

	util.ClearTerminal()

	// prompt user to select input device from list
	if len(availableDevices) > 0 {
		for {
			// prompt user for input
			terminalSep := strings.Repeat("-", 90)
			fmt.Println(terminalSep)
			fmt.Println("Select Input Device to Monitor: ")
			fmt.Println(terminalSep)
			// print discovered input devices
			for index, eventCaptureDevice := range availableDevices {
				fmt.Printf("%d) %s\n", index, eventCaptureDevice)
			}

			_, err := fmt.Scan(&selectedCaptureIndex)
			if err != nil {
				fmt.Println("Error reading User Input:", err)
			}

			// input validation
			if selectedCaptureIndex >= 0 && selectedCaptureIndex < len(availableDevices) {
				// set capture device
				currentCapDevice.CurrentCaptureDevice = availableDevices[selectedCaptureIndex]
				break
			} else {
				util.ClearTerminal()
				fmt.Println("Invalid Input Please Choose From the List Below.")
			}
		}
	} else {
		fmt.Println("No Input Capture Devices Found, Exiting.")
		os.Exit(0)
	}
	// todo this input should be selected on first-run or every run or?

	// open selected input device
	file, err := os.Open(currentCapDevice.CurrentCaptureDevice)
	if err != nil {
		fmt.Println("Error opening input event device:", err)
		return
	}
	defer file.Close()

	// initial metronome settings are in the metronome init func
	go tempo.MainMetronome.StartMetronome()
	//tempo.MainMetronome.SetMetronome(144, 4, 4)
	//go tempo.MainMetronome.StartMetronome()

	// attach monitor to input device should this be a go func? need to wait on exit sig...
	currentCapDevice.AttachInputStream(file)

	// fyne
	// need set monitor key? based on click? wizard on start-up
	// need legit layout

	// todo need another way to give user permission to input without allowing all userspace to access
	// todo needs to account for beats per quarter note - needs research...?
	// todo tick volume?

	// notes
	// make note that only non-zero averages are used
}
