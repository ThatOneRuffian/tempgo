package main

import (
	"tempgo/gui"
	"tempgo/tempo"
)

func init() {

	//InputFile = CurrentCapDevice.PromptCMDInputSelect()  // CLI mode

	// initial metronome core
	go tempo.MainMetronome.StartMetronome()

}

func main() {
	tempo.CreateNewDevMonitor()
	gui.TempgoFyneApp.FyneWindow.ShowAndRun()

	// time consuming but possible features
	// todo need another way to give user permission to input without allowing all userspace to access

	// todo
	// need to embed images into app

	// notes
	// make note that only non-zero averages are used
}
