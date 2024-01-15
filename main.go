package main

import (
	"tempgo/gui"
	"tempgo/tempo"
)

func init() {
	//InputFile = CurrentCapDevice.PromptCMDInputSelect()  // CLI mode
	gui.InitWindowReources(resourceQuarternotePng, resourcePlayiconPng, resourcePauseiconPng)

	// initial metronome core
	go tempo.MainMetronome.StartMetronome()
}

func main() {
	// start input select monitor
	tempo.CreateNewDevMonitor()

	// show and run fyne app
	gui.TempgoFyneApp.FyneWindow.ShowAndRun()
}
