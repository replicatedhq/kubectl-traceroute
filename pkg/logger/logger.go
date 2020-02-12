package logger

import (
	"fmt"

	"github.com/fatih/color"
)

type Logger struct {
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	if msg == "" {
		fmt.Println("")
		return
	}

	fmt.Println(fmt.Sprintf(msg, args...))
}

func (l *Logger) InfoNoNewLine(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func (l *Logger) Error(err error) {
	c := color.New(color.FgHiRed)
	c.Println(fmt.Sprintf("%#v", err))
}
