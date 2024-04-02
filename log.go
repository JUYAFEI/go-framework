package go_framework

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type LoggerConfig struct {
	Formatter LoggerFormatter
	out       io.Writer
}

type LoggerFormatter func(params LoggerFormatterParams) string

type LoggerFormatterParams struct {
	Request    *http.Request
	TimeStamp  time.Time
	StatusCode int
	Latency    time.Duration
	ClientIP   string
	Method     string
	Path       string
}

var DefaultWriter = os.Stdout

var defaultLogFormatter = func(param LoggerFormatterParams) string {
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}
	return fmt.Sprintf("%s | %3d | %13v | %15s | %-7s | %#v\n",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		param.StatusCode,
		param.Latency, param.ClientIP, param.Method, param.Path)
}

func LoggerWithConfig(config LoggerConfig, next HandlerFunc) HandlerFunc {
	if config.Formatter == nil {
		config.Formatter = defaultLogFormatter
	}
	out := config.out
	if out == nil {
		out = DefaultWriter
	}
	return func(c *Context) {
		param := LoggerFormatterParams{
			Request: c.R,
		}
		start := time.Now()
		path := c.R.URL.Path
		raw := c.R.URL.RawQuery
		next(c)
		stop := time.Now()
		latency := stop.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(c.R.RemoteAddr))
		clientIp := net.ParseIP(ip)
		method := c.R.Method
		statusCode := c.StatusCode

		if raw != "" {
			path = path + "?" + raw
		}
		param.TimeStamp = time.Now()
		param.StatusCode = statusCode
		param.Latency = latency
		param.ClientIP = clientIp.String()
		param.Method = method
		param.Path = path
		fmt.Fprint(out, config.Formatter(param))
	}
}

func Logging(next HandlerFunc) HandlerFunc {
	return LoggerWithConfig(LoggerConfig{}, next)
}
