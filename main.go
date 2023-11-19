package main

import (
	"fmt"
	"os"

	"tempgo/tempo"
	"tempgo/terminal"
)

func main() {
	var currentCapDevice tempo.BpmCaptureDevice
	var selectedCaptureIndex int = -1
	availableDevices, enumErr := tempo.GetInputDevices()

	if enumErr != nil {
		fmt.Println("Could Not Enumerate Input Devices.")
		os.Exit(0)
	}

	terminal.ClearTerminal()

	// prompt user to select input device from list
	if len(availableDevices) > 0 {
		for {
			// print discovered input devices
			for index, eventCaptureDevice := range availableDevices {
				fmt.Printf("%d) %s\n", index, eventCaptureDevice)
			}

			// prompt user for input
			fmt.Print("Select Input Capture Device: ")
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
				terminal.ClearTerminal()
				fmt.Println("Invalid Input Please Choose From the List Below:")
			}
		}
	} else {
		fmt.Println("No Input Capture Devices Found, Exiting.")
		os.Exit(0)
	}

	// open selected input device
	file, err := os.Open(currentCapDevice.CurrentCaptureDevice)
	if err != nil {
		fmt.Println("Error opening input event device:", err)
		return
	}
	defer file.Close()

	// attach monitor to input device should this be a go func
	go tempo.StartMetronome(100, 4, 4)
	tempo.AttachInputStream(file, &currentCapDevice)

	// todo data
	// - discard data outliners
	// - overall rating variance? other stuff.
	// - skip first run to populate the bpm?
	// - need to be able to pass capture device into metronome maybe for recording beats?
	//   how will this be stored?? timestamps how many etc....

	/// metronome
	// fine adjust while playing?
	// start/stop
	// metronome volume
	// measure against a given BPM and give accuracy and +-

	// fyne
	// need set monitor key? can this auto detect... based on click?
	// display BPM (ms per peak) and accurancy

}
