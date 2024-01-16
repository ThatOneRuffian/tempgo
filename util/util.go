package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

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

func GetInputDevices() ([]string, error) {
	dir := "/dev/input/by-id/"
	pattern := "*event*"

	devicePaths, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, err
	}

	return devicePaths, nil
}

func IntArrayToString(intArray [10]int) string {
	strArray := make([]string, len(intArray))

	// integer to string
	for i, v := range intArray {
		strArray[i] = strconv.Itoa(v)
	}

	// create string representation of list
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
