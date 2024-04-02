package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
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
	Level   LoggerLevel
	IsColor bool
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
		if out == os.Stdout {
			l.Formatter.IsColor = true
			formatter = l.Formatter.formatter(args)
		}
		fmt.Fprint(out, formatter)
	}

}

func (f *LoggerFormatter) formatter(msg any) string {
	now := time.Now()
	if f.IsColor {
		levelColor := f.LevelColor()
		msgColor := f.MsgColor()
		return fmt.Sprintf("%s [msgo] %s %s%v%s | level= %s %s %s | msg=%s %#v %s \n",
			yellow, reset, blue, now.Format("2006/01/02 - 15:04:05"), reset,
			levelColor, f.Level.Level(), reset, msgColor, msg, reset,
		)
	}
	return fmt.Sprintf("[go-framework] %v | level=%s | msg=%#v \n",
		now.Format("2006/01/02 - 15:04:05"),
		f.Level.Level(), msg,
	)
}

func (f *LoggerFormatter) LevelColor() string {
	switch f.Level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}

func (f *LoggerFormatter) MsgColor() string {
	switch f.Level {
	case LevelDebug:
		return ""
	case LevelInfo:
		return ""
	case LevelError:
		return red
	default:
		return cyan
	}
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
