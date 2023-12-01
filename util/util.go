package util

import (
	"fmt"
	"os"
	"os/exec"
)

func AdjustWavVolume(wavData []byte) []byte {
	var newWaveData []byte

	return newWaveData
}

func ClearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func CheckCommandExists(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
