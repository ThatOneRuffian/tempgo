package tempo

import (
	"fmt"
	"math"
	"os"
	"tempgo/gui"
	"tempgo/util"
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
	FirstRun             bool
}

func (cap *BpmCaptureDevice) AttachInputStream(file *os.File) {
	const eventSize = 24 // size of a single input event struct
	var eventBytes [eventSize]byte
	cap.CurrentCaptureBtn = 255
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
		if cap.CurrentCaptureBtn == 255 && eventCode != 4 {
			if runCount >= 4 { // filter the first few reads before accepting user input for cli
				cap.CurrentCaptureBtn = eventCode
				util.ClearTerminal()
			}
		}

		if eventCode == cap.CurrentCaptureBtn {
			// check for key press and release events
			if eventType == 1 && eventValue == 1 { // key press event
				// send input timestamp to metronome
				MainMetronome.inputCompare.inputSignalTime <- time.Now()
				go cap.printStats()
			} else if eventType == 0 && eventValue == 0 { // key release event
				//fmt.Printf("Key Release: Code=%d\n", eventCode)
			}
		}
	}
}

func (cap *BpmCaptureDevice) printStats() {
	currentTime := time.Now()
	timeDelta := currentTime.Sub(cap.beatInterval)
	bpm := 60 * time.Second / timeDelta
	//todo need to skip this using a first run flag or something - this can also be handled at the end of the loop?
	cap.bpmSamples[cap.currentSampleIndex] = int(bpm)
	cap.beatInterval = currentTime
	avgBpm := 0
	bpmSum := 0
	zeroCount := 0
	for _, bpmSample := range cap.bpmSamples {
		if bpmSample != 0 {
			bpmSum += bpmSample
		} else {
			zeroCount += 1
		}
	}

	if bpmSum > 0 {
		avgBpm = bpmSum / (len(cap.bpmSamples) - zeroCount)
	}

	precision := float64(timeDelta.Milliseconds() - cap.lastDelta.Milliseconds())

	// determine if the input was early or late
	inputSign := "+"
	if precision < 0 {
		inputSign = "-"
	}
	precision = math.Abs(precision)

	detectedInterval := int(timeDelta.Milliseconds())
	currentRawRating := CalculateInputRating(int64(precision))

	cov := calculateStandardDeviation(cap.bpmSamples, avgBpm) / float64(avgBpm)

	// print current raw input stats
	util.ClearTerminal()
	fmt.Println(cap.bpmSamples)
	fmt.Println("===========================================================================")
	fmt.Printf("Average BPM: %d\n", avgBpm)
	fmt.Printf("Detected Interval: %dms\n", detectedInterval)
	fmt.Printf("Last Detected BPM: %d\n", bpm)
	fmt.Printf("Detected Beat Offset: %s%dms\n", inputSign, int(precision))
	fmt.Printf("Overall Rating (Coefficient of Variation): +/-%.1f BPM\n", cov)
	fmt.Printf("Interval Rating: %s\n", currentRawRating)

	// print metronome compare stats
	fmt.Println("===========================================================================")
	fmt.Println("Metronome Stats:")
	fmt.Printf("Current BPM is: %d BPM\n", MainMetronome.CurrentTempo)
	fmt.Printf("Current Beat Interval is: %dms\n", int(60000.0/float64(MainMetronome.CurrentTempo)))
	fmt.Printf("Interval Compare Result: %s%dms\n", MainMetronome.inputCompare.inputOffsetSign, MainMetronome.inputCompare.inputOffset)
	fmt.Printf("Metronome Rating: %s\n", CalculateInputRating(MainMetronome.inputCompare.inputOffset))

	var inputOffset string
	if MainMetronome.isPlaying {
		inputOffset = fmt.Sprintf("%s%dms", MainMetronome.inputCompare.inputOffsetSign, MainMetronome.inputCompare.inputOffset)
	} else {
		inputOffset = "Start Metronome to Begin"
	}
	gui.TempgoStatData.RawInputArrayCV.Set(fmt.Sprint(cov))
	gui.TempgoStatData.MetronomeInputOffset.Set(inputOffset)
	gui.TempgoStatData.RawInputArray.Set(gui.IntArrayToString(cap.bpmSamples))
	gui.TempgoStatData.AverageBPM.Set(fmt.Sprintf("%d BPM", avgBpm))
	gui.TempgoStatData.DetectedInterval.Set(fmt.Sprintf("%sms", fmt.Sprint(detectedInterval)))
	gui.TempgoStatData.MetronomeCurrentBPM.Set(fmt.Sprint(MainMetronome.CurrentTempo))
	gui.TempgoStatData.IntervalRating.Set(currentRawRating)
	cap.lastDelta = timeDelta
	if cap.currentSampleIndex+1 < len(cap.bpmSamples) && !cap.FirstRun {
		cap.currentSampleIndex += 1
	} else {
		cap.currentSampleIndex = 0
	}
	cap.FirstRun = false
}

func calculateStandardDeviation(data [10]int, mean int) float64 {
	sumSquaredDiff := 0.0
	for _, value := range data {
		if value != 0 {
			diff := value - mean
			sumSquaredDiff += float64(diff * diff)
		}
	}
	variance := sumSquaredDiff / float64(len(data)-1)
	return math.Sqrt(variance)
}
