package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

type LoggerLevel int

const (
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

type Logger struct {
	Formatter LoggerFormatter
	Level     LoggerLevel
	Outs      []io.Writer
}

type LoggerFormatter struct {
	Level LoggerLevel
}

func DefaultLogger() *Logger {
	logger := NewLogger()
	out := os.Stdout
	logger.Outs = append(logger.Outs, out)
	logger.Level = LevelDebug
	logger.Formatter = LoggerFormatter{}
	return logger

}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Debug(args any) {
	l.Print(LevelDebug, args)
}

func (l *Logger) Info(args any) {
	l.Print(LevelInfo, args)
}

func (l *Logger) Error(args any) {
	l.Print(LevelError, args)
}

func (l *Logger) Print(level LoggerLevel, args any) {
	if l.Level > level {
		return
	}
	l.Formatter.Level = level
	formatter := l.Formatter.formatter(args)
	for _, out := range l.Outs {
		fmt.Fprint(out, formatter)
	}

}

func (f *LoggerFormatter) formatter(msg any) string {
	now := time.Now()
	return fmt.Sprintf("[go-framework] %v | level=%s | msg=%#v \n",
		now.Format("2006/01/02 - 15:04:05"),
		f.Level.Level(), msg,
	)
}

func (level LoggerLevel) Level() string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}
