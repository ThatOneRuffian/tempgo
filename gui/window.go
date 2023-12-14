package gui

import (
	"fmt"
	"image/color"
	"log"
	"strconv"
	"tempgo/util"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type metronomeGUI struct {
	fyneTitle           string
	InputChan           chan time.Time // input for GUI tempo button
	PlayMetronomeChan   chan bool
	PauseMetronomeChan  chan bool
	UpdateMetronomeChan chan bool
	fyneApp             fyne.App
	FyneWindow          fyne.Window
	inputDropdown       fyne.Widget
	tempoInputBtn       fyne.Widget
	metronomePlayBtn    fyne.Widget
	metronomePauseBtn   fyne.Widget
	metronomeUpdateBtn  fyne.Widget
	inputDevSelect      fyne.Widget
}

type metronomeStats struct {
	NewTempoFieldVal           binding.String
	NewBeatsPerMeasureFieldVal binding.String
	MetronomeCurrentBPM        binding.String
	MetronomeBeatInterval      binding.String
	MetronomeInputOffset       binding.String
	MetronomeBeatsPerMeasure   binding.String
	RawInputArrayCV            binding.String
	RawInputArray              binding.String
	AverageBPM                 binding.String
	DetectedInterval           binding.String
	OverallRatingString        binding.String
	IntervalRating             binding.String
}

var TempgoFyneApp metronomeGUI
var TempgoStatData metronomeStats

func init() {
	//init tempgo window elements
	TempgoFyneApp.fyneTitle = "Tempgo"
	TempgoFyneApp.InputChan = make(chan time.Time)
	TempgoFyneApp.PlayMetronomeChan = make(chan bool)
	TempgoFyneApp.PauseMetronomeChan = make(chan bool)
	TempgoFyneApp.UpdateMetronomeChan = make(chan bool)
	TempgoFyneApp.fyneApp = app.New()
	TempgoFyneApp.FyneWindow = TempgoFyneApp.fyneApp.NewWindow(TempgoFyneApp.fyneTitle)

	//todo need to query capture devices to populate this list and then add a selection event
	availableInputDevs, err := util.GetInputDevices()
	if err != nil {
		fmt.Println("Could Not Query Input Devices.")
	}
	inputDevSelect := widget.NewSelect(availableInputDevs, func(value string) {
		log.Println("Select set to", value)
	})

	// allocate button resources
	btnIcon, err := fyne.LoadResourceFromPath("./resources/quarter_note.png")
	if err != nil {
		fmt.Println("Unable to load button image resource.")
	}

	playIcon, err := fyne.LoadResourceFromPath("./resources/play_icon.png")
	if err != nil {
		fmt.Println("Unable to load button image resource.")
	}

	pauseIcon, err := fyne.LoadResourceFromPath("./resources/pause_icon.png")
	if err != nil {
		fmt.Println("Unable to load button image resource.")
	}

	// init gui buttons
	TempgoFyneApp.tempoInputBtn = widget.NewButtonWithIcon("", btnIcon, func() {
		TempgoFyneApp.InputChan <- time.Now()
	})

	TempgoFyneApp.metronomePlayBtn = widget.NewButtonWithIcon("", playIcon, func() {
		TempgoFyneApp.PlayMetronomeChan <- true
	})

	TempgoFyneApp.metronomePauseBtn = widget.NewButtonWithIcon("", pauseIcon, func() {
		TempgoFyneApp.PauseMetronomeChan <- true
	})

	TempgoFyneApp.metronomeUpdateBtn = widget.NewButton("Update Metronome", func() {
		TempgoFyneApp.UpdateMetronomeChan <- true
	})

	metronomeActionBtns := container.New(layout.NewGridLayoutWithColumns(2), TempgoFyneApp.metronomePlayBtn, TempgoFyneApp.metronomePauseBtn)

	// init all bounded gui elements
	TempgoStatData.OverallRatingString = binding.NewString()
	TempgoStatData.OverallRatingString.Set("Start Metronome to Begin")

	TempgoStatData.AverageBPM = binding.NewString()
	TempgoStatData.AverageBPM.Set("0 BPM")

	TempgoStatData.DetectedInterval = binding.NewString()
	TempgoStatData.DetectedInterval.Set("0ms")

	TempgoStatData.RawInputArray = binding.NewString()
	TempgoStatData.RawInputArray.Set("[0, 0, 0, 0, 0, 0, 0, 0, 0,0]")

	TempgoStatData.MetronomeInputOffset = binding.NewString()
	TempgoStatData.MetronomeInputOffset.Set("Start Metronome to Begin")

	TempgoStatData.IntervalRating = binding.NewString()
	TempgoStatData.IntervalRating.Set("Tap to Begin")

	TempgoStatData.RawInputArrayCV = binding.NewString()
	TempgoStatData.RawInputArrayCV.Set("0")

	TempgoStatData.NewTempoFieldVal = binding.NewString()
	TempgoStatData.MetronomeCurrentBPM = binding.NewString()
	TempgoStatData.MetronomeBeatInterval = binding.NewString()
	TempgoStatData.MetronomeBeatsPerMeasure = binding.NewString()
	TempgoStatData.NewBeatsPerMeasureFieldVal = binding.NewString()

	// init input area of window
	inputSelectionContainer := container.New(layout.NewFormLayout(), widget.NewLabel("Input Devices:"), inputDevSelect)
	inputAreaContainer := container.New(layout.NewStackLayout(), TempgoFyneApp.tempoInputBtn, inputSelectionContainer)

	// init metronome setting area of window
	metronomeForm := container.NewGridWithColumns(4, widget.NewLabel("New Tempo:"), widget.NewEntryWithData(TempgoStatData.NewTempoFieldVal), widget.NewLabel("New Beats Per Measure:"), widget.NewEntryWithData(TempgoStatData.NewBeatsPerMeasureFieldVal))
	updateMetronomeBtn := container.NewGridWithColumns(1, TempgoFyneApp.metronomeUpdateBtn)

	// init metronome stats area of window
	metronomeSettingContainer := container.NewGridWithRows(1,
		container.New(layout.NewFormLayout(), widget.NewLabel("Current Tempo:"), widget.NewLabelWithData(TempgoStatData.MetronomeCurrentBPM)),
		container.New(layout.NewFormLayout(), widget.NewLabel("Beat Interval:"), widget.NewLabelWithData(TempgoStatData.MetronomeBeatInterval)),
		container.New(layout.NewFormLayout(), widget.NewLabel("Beats Per Measure:"), widget.NewLabelWithData(TempgoStatData.MetronomeBeatsPerMeasure)),
	)

	metronomeStatsContainer := container.NewGridWithColumns(1,
		container.New(layout.NewFormLayout(), widget.NewLabel("Metronome Input Offset:"), widget.NewLabelWithData(TempgoStatData.MetronomeInputOffset)),
		container.New(layout.NewFormLayout(), widget.NewLabel("Metronome Rating:"), widget.NewLabelWithData(TempgoStatData.OverallRatingString)))

	// init raw stats area
	rawInputStatsContainer := container.NewVBox(
		widget.NewLabelWithStyle("Raw Input Stats:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.New(layout.NewFormLayout(), widget.NewLabel("Detected BPM Array:"), widget.NewLabelWithData(TempgoStatData.RawInputArray)),
		container.New(layout.NewFormLayout(), widget.NewLabel("Average BPM:"), widget.NewLabelWithData(TempgoStatData.AverageBPM)),
		container.New(layout.NewFormLayout(), widget.NewLabel("Detected Beat Interval:"), widget.NewLabelWithData(TempgoStatData.DetectedInterval)),
		container.New(layout.NewFormLayout(), widget.NewLabel("Overall Raiting (Coefficient of Variation | Lower is Better):"), widget.NewLabelWithData(TempgoStatData.RawInputArrayCV)),
		container.New(layout.NewFormLayout(), widget.NewLabel("Interval Rating:"), widget.NewLabelWithData(TempgoStatData.IntervalRating)))

	// finalize metronome container
	metronomeContainer := container.NewVBox(
		widget.NewLabelWithStyle("Metronome Settings:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		metronomeSettingContainer,
		metronomeActionBtns,
		canvas.NewLine((color.Black)),
		widget.NewLabelWithStyle("Update Metronome:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		metronomeForm,
		updateMetronomeBtn,
		canvas.NewLine((color.Black)),
		widget.NewLabelWithStyle("Metronome Stats:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		metronomeStatsContainer,
		canvas.NewLine((color.Black)),
		rawInputStatsContainer)

	// assemble final window layout
	windowContainer := container.NewGridWithColumns(2, inputAreaContainer, metronomeContainer)
	TempgoFyneApp.FyneWindow.SetContent(windowContainer)
	TempgoFyneApp.FyneWindow.Resize(fyne.NewSize(100, 100))
}

func IntArrayToString(intArray [10]int) string {
	// Create a slice of strings to store the string representations of integers
	strArray := make([]string, len(intArray))

	// Convert each integer to its string representation
	for i, v := range intArray {
		strArray[i] = strconv.Itoa(v)
	}

	// Concatenate the string representations
	result := "["
	for index, str := range strArray {

		result += str
		if index != len(strArray)-1 {
			result += ", "
		}

	}
	result = result + "]"

	return result
}
