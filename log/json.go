package log

import (
	"encoding/json"
	"fmt"
	"time"
)

type JsonFormatter struct {
	TimeDisplay bool
}

func (j *JsonFormatter) Formatter(param *LoggingFormatParam) string {
	now := time.Now()
	if param.LoggerFields == nil {
		param.LoggerFields = make(Fields)
	}

	if j.TimeDisplay {
		timeNow := now.Format("2006-01-02 - 15:04:05")
		param.LoggerFields["log_time"] = timeNow
	}
	param.LoggerFields["msg"] = param.Msg
	marshal, _ := json.Marshal(param.LoggerFields)
	return fmt.Sprint(string(marshal))

}
