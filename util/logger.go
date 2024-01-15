package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var logger log.Logger
var logFile *os.File

func init() {

	// open log file for r/w
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		log.Println("Error opening log file:", err)
		log.Println("Continuing with stderr output")
	}

	// format log to print in UTC time
	logger = *log.New(logFile, "", log.Ldate|log.Ltime)

	// TODO what happens when logFile/stderr is nil?
	multiWriter := io.MultiWriter(os.Stderr, logFile)

	// write to stderr and log file
	logger.SetOutput(multiWriter)
}

func Log(stringList ...any) {
	logger.Println("(WARN):\t", anyToString(stringList))
}

func LogFatal(stringList ...any) {
	logger.Fatalln("(FATAL):\t", anyToString(stringList))
}

func CloseLogger() {
	logFile.Close()
}

func anyToString(values ...any) string {
	var logText []string
	for _, value := range values {
		stringVal := strings.Trim(fmt.Sprintf("%v", value), "[]")
		logText = append(logText, stringVal)
	}
	return logText[0]
}
