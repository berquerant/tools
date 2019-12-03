/*
Package log provides logger
*/
package log

import (
	"fmt"
	"os"
)

type (
	// Level is log level
	Level int
	// Logger writes log
	Logger struct {
		level Level
	}
	// Option indicates logger option
	Option func(logger *Logger)
)

const (
	Debug Level = iota
	Info
	None
)

var (
	levelToLabel = map[Level]string{
		Debug: "D",
		Info:  "I",
	}
)

// NewLogger makes logger and applies options
func NewLogger(options ...Option) *Logger {
	x := &Logger{}
	for _, opt := range options {
		opt(x)
	}
	return x
}

// WithLevel returns option that set log level
func WithLevel(level Level) Option {
	return Option(func(logger *Logger) {
		logger.level = level
	})
}

func (s *Logger) log(level Level, format string, a ...interface{}) {
	if s.level > level {
		return
	}
	var label string
	if x, ok := levelToLabel[level]; ok {
		label = fmt.Sprintf("%s | ", x)
	}
	_, _ = fmt.Fprintf(os.Stderr, "%s%s\n", label, fmt.Sprintf(format, a...))
}

// Info writes info log
func (s *Logger) Info(format string, a ...interface{}) {
	s.log(Info, format, a...)
}

// Debug writes debug log
func (s *Logger) Debug(format string, a ...interface{}) {
	s.log(Debug, format, a...)
}
