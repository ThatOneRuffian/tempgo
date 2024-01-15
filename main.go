package main

import (
	"flag"
	"tempgo/gui"
	"tempgo/tempo"
)

func init() {
	gui.InitWindowReources(resourceQuarternotePng, resourcePlayiconPng, resourcePauseiconPng)
}

func main() {
	// parse args
	consoleMode := flag.Bool("console", false, "Run the application in console mode.")
	consoleSound := flag.Bool("sound", false, "Enable metronome sound when using the console.")
	metronomeBPM := flag.Int("bpm", 77, "Run the application in console mode.")
	metronomeCount := flag.Int("count", 4, "Run the application in console mode.")
	flag.Parse()

	// start input select monitor
	bpmCapDev := tempo.CreateNewDevMonitor()
	if !*consoleMode {
		// initialize metronome core
		go tempo.MainMetronome.StartMetronome()
		// show and run fyne app
		gui.TempgoFyneApp.FyneWindow.ShowAndRun()
	} else {
		// if tempo sound, or non-default tempo/bpm then play
		if *consoleSound || *metronomeBPM != 77 || *metronomeCount != 4 {
			// initialize metronome core
			go tempo.MainMetronome.StartMetronome()
			// set tempo
			bpmCapDev.PlayCMDMetronome(*metronomeBPM, *metronomeCount)
		}
		// CLI mode
		bpmCapDev.PromptCMDInputSelect()
		for {
		}
	}
}
