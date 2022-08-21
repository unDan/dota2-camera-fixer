package logger

import (
	"fmt"
	"log"
)

type Logger struct {
	showLogInfo bool
}

func NewLogger(showLogInfo bool) Logger {
	return Logger{
		showLogInfo: showLogInfo,
	}
}

func (l Logger) Print(msg string, a ...any) {
	logMsg := fmt.Sprintf("%s", fmt.Sprintf(msg, a...))
	log.Println(logMsg)
}

func (l Logger) Info(msg string, a ...any) {
	if !l.showLogInfo {
		return
	}

	logMsg := fmt.Sprintf("[INFO] %s", fmt.Sprintf(msg, a...))
	log.Println(logMsg)
}

func (l Logger) Error(msg string, a ...any) {
	logMsg := fmt.Sprintf("[ERROR] %s", fmt.Sprintf(msg, a...))
	log.Fatalln(logMsg)
}
