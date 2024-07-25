package logs

import (
	"fmt"
	"time"
)

func Info(message string) {
	fmt.Printf("%s INFO %s\n", timeString(), message)
}

func Warn(message string) {
	fmt.Printf("%s WARN %s\n", timeString(), message)
}

func Severe(message string) {
	fmt.Printf("%s SEVERE %s\n", timeString(), message)
}

func Panic(message string) {
	formattedMessage := fmt.Sprintf("%s PANIC %s\n", timeString(), message)
	panic(formattedMessage)
}

func timeString() string {
	now := time.Now()
	return now.Format("02.01.2006 15:04:05.000 MST")
}
