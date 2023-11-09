package tempo

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"tempgo/terminal"
	"time"
)

type BpmCaptureDevice struct {
	CurrentBPM           float64
	CurrentCaptureDevice string
	CurrentCaptureBtn    uint16
	keyInterval          time.Time
	lastDelta            time.Duration
}

func AttachInputStream(file *os.File, captureDevice *BpmCaptureDevice) {
	const eventSize = 24 // size of a single input event struct
	var eventBytes [eventSize]byte
	captureDevice.CurrentCaptureBtn = 255
	runCount := 0
	fmt.Println("Press Button to Monitor.")
	time.Sleep(time.Second)
	for {
		// sample rate? max bpm?
		_, err := file.Read(eventBytes[:])
		if err != nil {
			fmt.Println("Error reading from input event device:", err)
			return
		}

		// decode the event data
		eventType := uint16(eventBytes[16]) | (uint16(eventBytes[17]) << 8)
		eventCode := uint16(eventBytes[18]) | (uint16(eventBytes[19]) << 8)
		eventValue := int32(eventBytes[20]) | (int32(eventBytes[21]) << 8) | (int32(eventBytes[22]) << 16) | (int32(eventBytes[23]) << 24)
		runCount += 1

		// set input monitor button
		if captureDevice.CurrentCaptureBtn == 255 && eventCode != 4 {
			if runCount >= 4 { // have to filter the first few reads before accepting user input
				captureDevice.CurrentCaptureBtn = eventCode
			}
		}

		if eventCode == captureDevice.CurrentCaptureBtn {
			// check for key press and release events
			if eventType == 1 && eventValue == 1 { // key press event
				currentTime := time.Now()
				timeDelta := currentTime.Sub(captureDevice.keyInterval)
				bpm := 60 * time.Second / timeDelta
				captureDevice.keyInterval = currentTime
				terminal.ClearTerminal()

				// print interface
				fmt.Println(timeDelta)
				fmt.Printf("Projected BPM: %d\n", bpm)
				precision := float64(timeDelta.Milliseconds() - captureDevice.lastDelta.Milliseconds())
				rating := ""

				// determine if the input was early or late
				inputSign := "+"
				if precision < 0 {
					inputSign = "-"
				}
				precision = math.Abs(precision)

				// determine input accuracy
				if precision >= 0 && precision <= 20 {
					rating = "Tight"
				} else if precision >= 21 && precision <= 40 {
					rating = "Acceptable"
				} else if precision >= 41 {
					rating = "Noticable"
				}

				fmt.Printf("Precison: %s%dms\n", inputSign, int(precision))
				fmt.Printf("Rating: %s", rating)

				captureDevice.lastDelta = timeDelta
			} else if eventType == 0 && eventValue == 0 { // key release event
				fmt.Printf("Key Release: Code=%d\n", eventCode)
			}
		}
	}
}

func GetInputDevices() ([]string, error) {
	dir := "/dev/input/by-id/"
	pattern := "*event*"

	devicePaths, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, err
	}

	return devicePaths, nil
}
