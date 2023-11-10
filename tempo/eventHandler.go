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
	CurrentCaptureDevice string
	CurrentCaptureBtn    uint16
	AvgBPM               float64
	keyInterval          time.Time
	bpmSamples           [10]int       // bpm sample storage
	lastDelta            time.Duration // this stores the delta between the
}

func AttachInputStream(file *os.File, captureDevice *BpmCaptureDevice) {
	const eventSize = 24 // size of a single input event struct
	var eventBytes [eventSize]byte
	captureDevice.CurrentCaptureBtn = 255
	currentSampleIndex := 0
	runCount := 0
	fmt.Println("Press Button to Monitor.")
	for {
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
				captureDevice.bpmSamples[currentSampleIndex] = int(bpm)
				captureDevice.keyInterval = currentTime
				avgBpm := 0
				bpmSum := 0
				for _, bpmSample := range captureDevice.bpmSamples {
					bpmSum += bpmSample
				}
				avgBpm = bpmSum / len(captureDevice.bpmSamples)
				terminal.ClearTerminal()

				// print current stats
				fmt.Println(captureDevice.bpmSamples)
				fmt.Printf("Average BPM: %d\n", avgBpm)
				fmt.Printf("Detected Interval: %dms\n", int(timeDelta.Milliseconds()))
				fmt.Printf("Last Detected BPM: %d\n", bpm)
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

				fmt.Printf("Beat Offset: %s%dms\n", inputSign, int(precision))
				fmt.Printf("Rating: %s", rating)

				captureDevice.lastDelta = timeDelta
				if currentSampleIndex+1 < len(captureDevice.bpmSamples) {
					currentSampleIndex += 1
				} else {
					currentSampleIndex = 0
				}
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
