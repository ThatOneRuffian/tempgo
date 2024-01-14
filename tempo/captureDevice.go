package tempo

import (
	"fmt"
	"math"
	"os"
	"strings"
	"tempgo/gui"
	"tempgo/util"
	"time"

	"fyne.io/fyne/v2/dialog"
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

func CreateNewCaptureDev() *BpmCaptureDevice {
	var newCapDev BpmCaptureDevice
	newCapDev.FirstRun = true
	go newCapDev.StartInputEventListeners()
	go newCapDev.AttachInputStream()
	return &newCapDev
}

func (cap *BpmCaptureDevice) StartInputEventListeners() {
	// listen for gui and external input signals
	go func() {
		for {
			select {
			// external input signal
			case inputSig := <-MainMetronome.inputCompare.inputSignalTime:
				nanoSeconds := 1e9 * 60 / MainMetronome.CurrentTempo // convert bpm to ns
				tickRate := time.Duration(nanoSeconds)
				MainMetronome.calculateInputDelta(tickRate, inputSig, MainMetronome.LastTickTime)
				if MainMetronome.isPlaying {
					gui.TempgoStatData.OverallRatingString.Set(CalculateInputRating(MainMetronome.inputCompare.inputOffset))
				} else {
					gui.TempgoStatData.OverallRatingString.Set("Start Metronome to Begin")
				}
			// gui input signal
			case guiBtnInputSig := <-gui.TempgoFyneApp.InputChanTime:
				nanoSeconds := 1e9 * 60 / MainMetronome.CurrentTempo // convert bpm to ns
				tickRate := time.Duration(nanoSeconds)
				MainMetronome.calculateInputDelta(tickRate, guiBtnInputSig, MainMetronome.LastTickTime)
				if MainMetronome.isPlaying {
					gui.TempgoStatData.OverallRatingString.Set(CalculateInputRating(MainMetronome.inputCompare.inputOffset))
				} else {
					gui.TempgoStatData.OverallRatingString.Set("Start Metronome to Begin")
				}
				go cap.printStats()
			}
		}
	}()
}

func (cap *BpmCaptureDevice) AttachInputStream() {

	var monitoring bool
	var capDevFile *os.File
	stopSignal := make(chan bool)
	go func() {
		for {
			select {
			case inputDevice := <-gui.TempgoFyneApp.UpdateInputDevChan:
				fmt.Println("got channel info", inputDevice)

				// close current device if open and not same
				if capDevFile != nil {
					if inputDevice != capDevFile.Name() {
						capDevFile.Close()
					} else {
						//reprompt for input key?
					}
				}

				if capDevFile == nil {
					// open new provided device
					var err error
					capDevFile, err = os.Open(inputDevice)
					if err != nil {
						fmt.Println("Error opening input event device:", err)
						// this should prob send a dialog box or something
						continue
					}
					// dialog here
					form := dialog.NewForm("Test Dialog", "This is my confirm message", "dismiss", nil, func(a bool) {}, gui.TempgoFyneApp.FyneWindow)
					// set input monitor button
					// todo should have a time out or something here - revert to nil
					fmt.Println("Press Button to Monitor")
					go form.Show()
					for {
						cap.CurrentCaptureBtn = 255
						_, eventCode, _ := readEventData(capDevFile)
						if cap.CurrentCaptureBtn == 255 && eventCode != 4 {
							cap.CurrentCaptureBtn = eventCode
							form.Hide()
							if monitoring {
								stopSignal <- true
							}
							break
							//util.ClearTerminal()
							// enter metronome monitor loop
						} else {
							fmt.Println("Oh")
						}
					}
				}

				// spawn new go-thread for monitoring this specific input device
				go func(stopSig chan bool) {
					monitoring = true
					for {
						// break out of current loop to create a new non-blocking loop
						if len(stopSig) > 0 {
							<-stopSig
							monitoring = false
							break
						}

						eventType, eventCode, eventValue := readEventData(capDevFile)
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
				}(stopSignal)
			}
		}
	}()
}

func (cap *BpmCaptureDevice) printStats() {
	currentTime := time.Now()
	timeDelta := currentTime.Sub(cap.beatInterval)
	bpm := 60 * time.Second / timeDelta
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
	gui.TempgoStatData.RawInputArrayCV.Set(fmt.Sprintf("%.3f", cov))
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

func readEventData(capDevFile *os.File) (eventType uint16, eventCode uint16, eventValue int32) {
	const eventSize = 24 // size of a single input event struct
	var eventBytes [eventSize]byte
	_, err := capDevFile.Read(eventBytes[:])
	if err != nil {
		fmt.Println("Error reading from input event device:", err)
		return
	}

	// decode the event data
	eventType = uint16(eventBytes[16]) | (uint16(eventBytes[17]) << 8)
	eventCode = uint16(eventBytes[18]) | (uint16(eventBytes[19]) << 8)
	eventValue = int32(eventBytes[20]) | (int32(eventBytes[21]) << 8) | (int32(eventBytes[22]) << 16) | (int32(eventBytes[23]) << 24)

	return eventType, eventCode, eventValue
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

func (cap *BpmCaptureDevice) PromptCMDInputSelect() *os.File {
	var selectedCaptureIndex int = -1
	availableDevices, enumErr := util.GetInputDevices()

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
				cap.CurrentCaptureDevice = availableDevices[selectedCaptureIndex]
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
	// open selected input device
	fmt.Println(cap.CurrentCaptureDevice)
	fmt.Println("hello")

	capDevFile, err := os.Open(cap.CurrentCaptureDevice)

	if err != nil {
		fmt.Println("Error opening input event device:", err)
		return nil
	}

	return capDevFile
}
