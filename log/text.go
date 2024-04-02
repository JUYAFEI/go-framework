package log

import (
	"fmt"
	"strings"
	"time"
)

type TextFormatter struct {
}

func (t *TextFormatter) Formatter(param *LoggerFormatter) string {
	now := time.Now()
	var builderField strings.Builder
	var fieldsDisplay = ""
	if param.LoggerFields != nil {
		fieldsDisplay = "| fields: "
		num := len(param.LoggerFields)
		count := 0
		for k, v := range param.LoggerFields {
			fmt.Fprintf(&builderField, "%s=%v ", k, v)
			if count < num-1 {
				fmt.Fprintf(&builderField, ",")
				count++
			}
		}
	}
	if param.IsColor {
		//要带颜色  error的颜色 为红色 info为绿色 debug为蓝色
		levelColor := t.LevelColor(param.Level)
		msgColor := t.MsgColor(param.Level)
		return fmt.Sprintf("%s [go-frameword] %s %s%v%s | level= %s %s %s | msg=%s %#v %s %s %s \n",
			yellow, reset, blue, now.Format("2006/01/02 - 15:04:05"), reset,
			levelColor, param.Level.Level(), reset, msgColor, param.Msg, reset, fieldsDisplay, builderField.String(),
		)
	}
	return fmt.Sprintf("[go-frameword] %v | level=%s | msg= %#v %s %s \n",
		now.Format("2006/01/02 - 15:04:05"),
		param.Level.Level(), param.Msg, fieldsDisplay, builderField.String(),
	)
}

func (t *TextFormatter) LevelColor(level LoggerLevel) string {
	switch level {
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

func (t *TextFormatter) MsgColor(level LoggerLevel) string {
	switch level {
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
