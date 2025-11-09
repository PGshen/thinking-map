package callback

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
	once.Do(func() {
		os.MkdirAll("logs", 0755)
		// Use a daily rolling writer with optional size-based rotation to prevent oversized log files.
		// Configure max size via env var LOG_MAX_SIZE_MB (megabytes). If empty or <=0, size-based rotation is disabled.
		maxSizeBytes := int64(100 * 1024 * 1024)
		if v := os.Getenv("LOG_MAX_SIZE_MB"); v != "" {
			if mb, err := parseInt64(v); err == nil && mb > 0 {
				maxSizeBytes = mb * 1024 * 1024
			}
		}
		writer := NewDailyRollingWriter("logs", "callback", maxSizeBytes)
		config := &LogCallbackConfig{
			Detail: true,
			Writer: writer,
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

// DailyRollingWriter implements an io.Writer that writes to log files
// named by day (YYYY-MM-DD). It automatically rotates the file when the day changes.
type DailyRollingWriter struct {
	dir     string
	prefix  string
	mu      sync.Mutex
	curDate string
	file    *os.File
	// size-based rotation
	maxSizeBytes int64
	currentSize  int64
}

// NewDailyRollingWriter creates a new writer under dir with file name prefix.
// Files will be created as: <dir>/<prefix>-YYYY-MM-DD.log
// If maxSizeBytes > 0, a size-based rotation is enabled: when exceeding max size,
// a new file will be created with a time suffix: <prefix>-YYYY-MM-DD-HHMMSS.log
func NewDailyRollingWriter(dir, prefix string, maxSizeBytes int64) *DailyRollingWriter {
	w := &DailyRollingWriter{dir: dir, prefix: prefix, maxSizeBytes: maxSizeBytes}
	// Initialize with current day
	_ = w.rotateIfNeeded(time.Now())
	return w
}

// Write implements io.Writer and rotates the underlying file when day changes.
func (w *DailyRollingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	if err := w.rotateIfNeeded(now); err != nil {
		return 0, err
	}
	// size-based rotation: pre-rotate if this write will exceed the max size
	if w.file != nil && w.maxSizeBytes > 0 && w.currentSize+int64(len(p)) > w.maxSizeBytes {
		if err := w.rotateForSize(now); err != nil {
			return 0, err
		}
	}
	if w.file == nil {
		return 0, fmt.Errorf("log writer not initialized")
	}
	n, err := w.file.Write(p)
	if err == nil {
		w.currentSize += int64(n)
	}
	return n, err
}

// rotateIfNeeded checks the current date and opens a new file if the day changed.
func (w *DailyRollingWriter) rotateIfNeeded(t time.Time) error {
	dateStr := t.Format("2006-01-02")
	if w.file != nil && w.curDate == dateStr {
		return nil
	}
	// Close previous file if open
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}
	// Ensure directory exists
	if err := os.MkdirAll(w.dir, 0755); err != nil {
		return err
	}
	filename := fmt.Sprintf("%s/%s-%s.log", w.dir, w.prefix, dateStr)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	w.file = f
	w.curDate = dateStr
	// set current size
	if fi, err := f.Stat(); err == nil {
		w.currentSize = fi.Size()
	} else {
		w.currentSize = 0
	}
	// If base file already exceeds max size, immediately rotate by size to a new file
	if w.maxSizeBytes > 0 && w.currentSize >= w.maxSizeBytes {
		return w.rotateForSize(t)
	}
	return nil
}

// rotateForSize rotates the file due to size limit, keeping the same date but using a time-suffixed filename
func (w *DailyRollingWriter) rotateForSize(t time.Time) error {
	// Close current file first
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}
	// Ensure directory exists
	if err := os.MkdirAll(w.dir, 0755); err != nil {
		return err
	}
	timeSuffix := t.Format("150405")
	filename := fmt.Sprintf("%s/%s-%s-%s.log", w.dir, w.prefix, w.curDate, timeSuffix)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	w.file = f
	w.currentSize = 0
	return nil
}

// Close closes the current file if open.
func (w *DailyRollingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

// parseInt64 safely parses a string to int64
func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscan(s, &v)
	return v, err
}
