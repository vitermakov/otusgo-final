package logger

import (
	"errors"
)

var (
	ErrorUnknownLevel = errors.New("unknown log level")
	ErrorOpenLogFile  = errors.New("error open log file")
)

type Level int

const (
	LevelNone Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelOut
)

// Logger интерфейс логгера.
type Logger interface {
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Fatal(string, ...interface{})
	Debug(string, ...interface{})
}

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	case LevelDebug:
		return "debug"
	case LevelNone, LevelOut:
	}
	return "none"
}

func (l Level) Valid() bool {
	return l > LevelNone && l < LevelOut
}

func ParseLevel(level string) (Level, error) {
	var l Level
	switch level {
	case "info":
		l = LevelInfo
	case "warn":
		l = LevelWarn
	case "error":
		l = LevelError
	case "fatal":
		l = LevelFatal
	case "debug":
		l = LevelDebug
	default:
		return LevelNone, ErrorUnknownLevel
	}
	return l, nil
}

// Config для всех логгеров.
type Config struct {
	Level     Level
	FileName  string
	IsTesting bool
}
