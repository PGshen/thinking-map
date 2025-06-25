package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/cloudwego/eino/callbacks"
)

var LogCbHandler callbacks.Handler

var once sync.Once

func init() {
	var err error
	once.Do(func() {
		os.MkdirAll("logs", 0755)
		var f *os.File
		f, err = os.OpenFile("logs/callback.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return
		}
		config := &LogCallbackConfig{
			Detail: true,
			Writer: f,
		}
		if os.Getenv("DEBUG") == "true" {
			config.Debug = true
		}
		LogCbHandler = LogCallback(config)
	})
}

type LogCallbackConfig struct {
	Detail bool
	Debug  bool
	Writer io.Writer
}

// LogCallback 日志回调
func LogCallback(config *LogCallbackConfig) callbacks.Handler {
	if config == nil {
		config = &LogCallbackConfig{
			Detail: true,
			Writer: os.Stdout,
		}
	}
	if config.Writer == nil {
		config.Writer = os.Stdout
	}
	builder := callbacks.NewHandlerBuilder()
	builder.OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
		fmt.Fprintf(config.Writer, "[%s] start [%s:%s:%s]\n", time.Now(), info.Component, info.Name, info.Type)
		if config.Detail {
			var b []byte
			if config.Debug {
				b, _ = json.MarshalIndent(input, "", "  ")
			} else {
				b, _ = json.Marshal(input)
			}
			fmt.Fprintf(config.Writer, "[%s] input: %s\n", time.Now(), string(b))
		}
		return ctx
	})
	builder.OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		fmt.Fprintf(config.Writer, "[%s] end [%s:%s:%s]\n", time.Now(), info.Component, info.Name, info.Type)
		if config.Detail {
			var b []byte
			if config.Debug {
				b, _ = json.MarshalIndent(output, "", "  ")
			} else {
				b, _ = json.Marshal(output)
			}
			fmt.Fprintf(config.Writer, "[%s] output: %s\n", time.Now(), string(b))
		}
		return ctx
	})
	return builder.Build()
}
