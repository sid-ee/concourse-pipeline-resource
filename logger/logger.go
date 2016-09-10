package logger

import (
	"fmt"
	"io"
)

type Logger struct {
	sink io.Writer
}

func NewLogger(sink io.Writer) *Logger {
	return &Logger{
		sink: sink,
	}
}

func (l Logger) Debugf(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(l.sink, format, a...)
}
