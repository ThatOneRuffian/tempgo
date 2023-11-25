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
	beatInterval         time.Time
	bpmSamples           [10]int // bpm sample storage for raw tests
	currentSampleIndex   int
	lastDelta            time.Duration // this stores the delta between inputs
}

func AttachInputStream(file *os.File, captureDevice *BpmCaptureDevice) {
	const eventSize = 24 // size of a single input event struct
	var eventBytes [eventSize]byte
	captureDevice.CurrentCaptureBtn = 255
	runCount := 0

	fmt.Println("Press Key to Monitor.")
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
			if runCount >= 4 { // have to filter the first few reads before accepting user input for cli
				captureDevice.CurrentCaptureBtn = eventCode
				terminal.ClearTerminal()
			}
		}

		if eventCode == captureDevice.CurrentCaptureBtn {
			// check for key press and release events
			if eventType == 1 && eventValue == 1 { // key press event
				// send input timestamp to metronome
				MainMetronome.inputCompare.inputSignalTime <- time.Now()
				go printStats(captureDevice)
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

func printStats(captureDevice *BpmCaptureDevice) {
	currentTime := time.Now()
	timeDelta := currentTime.Sub(captureDevice.beatInterval)
	bpm := 60 * time.Second / timeDelta
	captureDevice.bpmSamples[captureDevice.currentSampleIndex] = int(bpm)
	captureDevice.beatInterval = currentTime
	avgBpm := 0
	bpmSum := 0
	for _, bpmSample := range captureDevice.bpmSamples {
		bpmSum += bpmSample
	}

	// what is a good delta? 20% diff? exlude result?
	avgBpm = bpmSum / len(captureDevice.bpmSamples) // this gives average nano seconds since unix epoch

	// todo need to put stats from MainMetronome

	precision := float64(timeDelta.Milliseconds() - captureDevice.lastDelta.Milliseconds())

	// determine if the input was early or late
	inputSign := "+"
	if precision < 0 {
		inputSign = "-"
	}
	precision = math.Abs(precision)

	// print current raw input stats
	terminal.ClearTerminal()
	fmt.Println(captureDevice.bpmSamples)
	fmt.Println("===========================================================================")
	fmt.Printf("Average BPM: %d\n", time.Duration(avgBpm))
	fmt.Printf("Detected Interval: %dms\n", int(timeDelta.Milliseconds()))
	fmt.Printf("Last Detected BPM: %d\n", bpm)
	fmt.Printf("Detected Beat Offset: %s%dms\n", inputSign, int(precision))
	fmt.Printf("Rating: %s\n", CalculateInputRating(int64(precision)))

	// print metronome compare stats
	fmt.Println("===========================================================================")
	fmt.Println("Metronome Stats:")
	fmt.Printf("Current BPM is: %d BPM\n", MainMetronome.currentBpm)
	fmt.Printf("Current Beat Interval is: %dms\n", int(60000.0/float64(MainMetronome.currentBpm)))
	fmt.Printf("Interval Compare Result: %s%dms\n", MainMetronome.inputCompare.inputOffsetSign, MainMetronome.inputCompare.inputOffset)
	fmt.Printf("Metronome Rating: %s\n", CalculateInputRating(MainMetronome.inputCompare.inputOffset))

	captureDevice.lastDelta = timeDelta
	if captureDevice.currentSampleIndex+1 < len(captureDevice.bpmSamples) {
		captureDevice.currentSampleIndex += 1
	} else {
		captureDevice.currentSampleIndex = 0
	}
}
