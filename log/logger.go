package log

import (
	"fmt"
	"io"
	"os"
	"path"
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
	Formatter    LoggingFormatter
	Level        LoggerLevel
	Outs         []*LoggerWriter
	LoggerFields Fields
	logPath      string
}

type Fields map[string]any

type LoggerFormatter struct {
	Level        LoggerLevel
	IsColor      bool
	LoggerFields Fields
}

type LoggingFormatter interface {
	Formatter(Param *LoggingFormatParam) string
}

type LoggingFormatParam struct {
	Level        LoggerLevel
	IsColor      bool
	LoggerFields Fields
	Msg          any
}

type LoggerWriter struct {
	Level LoggerLevel
	Out   io.Writer
}

func FileWriter(name string) (io.Writer, error) {
	w, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	return w, err
}

func DefaultLogger() *Logger {
	logger := NewLogger()
	logger.Level = LevelDebug
	w := &LoggerWriter{
		Level: LevelDebug,
		Out:   os.Stdout,
	}
	logger.Outs = append(logger.Outs, w)
	logger.Formatter = &TextFormatter{}
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
	param := &LoggingFormatParam{
		Level:        level,
		Msg:          args,
		LoggerFields: l.LoggerFields,
	}
	formatter := l.Formatter.Formatter(param)
	for _, out := range l.Outs {
		if out.Out == os.Stdout {
			param.IsColor = true
			formatter = l.Formatter.Formatter(param)
			fmt.Fprint(out.Out, formatter)
		}
		if out.Level == -1 || out.Level == level {
			fmt.Fprintln(out.Out, formatter)
		}
	}

}

func (l *Logger) SetLogPath(logPath string) {
	l.logPath = logPath
	all, err := FileWriter(path.Join(l.logPath, time.Now().Format("2006-01-02")+"-all.log"))
	if err != nil {
		panic(err)
	}
	l.Outs = append(l.Outs, &LoggerWriter{Level: -1, Out: all})
	debug, err := FileWriter(path.Join(l.logPath, time.Now().Format("2006-01-02")+"-debug.log"))
	if err != nil {
		panic(err)
	}
	l.Outs = append(l.Outs, &LoggerWriter{Level: LevelDebug, Out: debug})
	info, err := FileWriter(path.Join(l.logPath, time.Now().Format("2006-01-02")+"-info.log"))
	if err != nil {
		panic(err)
	}
	l.Outs = append(l.Outs, &LoggerWriter{Level: LevelInfo, Out: info})
	logError, err := FileWriter(path.Join(l.logPath, time.Now().Format("2006-01-02")+"-error.log"))
	if err != nil {
		panic(err)
	}
	l.Outs = append(l.Outs, &LoggerWriter{Level: LevelError, Out: logError})
}

func (l *Logger) WithFields(fields Fields) *Logger {
	return &Logger{
		Formatter:    l.Formatter,
		Outs:         l.Outs,
		Level:        l.Level,
		LoggerFields: fields,
	}
}

func (f *LoggerFormatter) Formatter(param *LoggingFormatParam) string {
	now := time.Now()
	if f.IsColor {
		levelColor := f.LevelColor()
		msgColor := f.MsgColor()
		return fmt.Sprintf("%s [go-framework] %s %s%v%s | level= %s %v %s | msg=%s %#v %s | fields=%v\n",
			yellow, reset, blue, now.Format("2006/01/02 - 15:04:05"), reset,
			levelColor, param.Level, reset, msgColor, param.Msg, reset, param.LoggerFields,
		)
	}
	return fmt.Sprintf("[go-framework] %v | level=%s | msg=%#v fields=%v\n",
		now.Format("2006/01/02 - 15:04:05"),
		f.Level.Level(), param.Msg, param.LoggerFields,
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
