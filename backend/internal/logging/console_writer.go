package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	consoleMode       = "console"
	consoleTimeFormat = "2006-01-02 15:04:05"
)

// SetupConsoleWriter 安装符合后端控制台阅读习惯的 logx 输出格式。
func SetupConsoleWriter(mode string) {
	if mode != "" && mode != consoleMode {
		return
	}

	logx.SetWriter(newConsoleWriter(os.Stdout, os.Stderr))
}

type consoleWriter struct {
	stdout io.Writer
	stderr io.Writer
	lock   sync.Mutex
}

func newConsoleWriter(stdout, stderr io.Writer) *consoleWriter {
	return &consoleWriter{
		stdout: stdout,
		stderr: stderr,
	}
}

func (w *consoleWriter) Alert(v any) {
	w.write(w.stderr, "alert", v)
}

func (w *consoleWriter) Close() error {
	return nil
}

func (w *consoleWriter) Debug(v any, fields ...logx.LogField) {
	w.write(w.stdout, "debug", v, fields...)
}

func (w *consoleWriter) Error(v any, fields ...logx.LogField) {
	w.write(w.stderr, "error", v, fields...)
}

func (w *consoleWriter) Info(v any, fields ...logx.LogField) {
	w.write(w.stdout, "info", v, fields...)
}

func (w *consoleWriter) Severe(v any) {
	w.write(w.stderr, "fatal", v)
}

func (w *consoleWriter) Slow(v any, fields ...logx.LogField) {
	w.write(w.stderr, "slow", v, fields...)
}

func (w *consoleWriter) Stack(v any) {
	w.write(w.stderr, "error", v)
}

func (w *consoleWriter) Stat(v any, fields ...logx.LogField) {
	w.write(w.stdout, "stat", v, fields...)
}

func (w *consoleWriter) write(out io.Writer, level string, v any, fields ...logx.LogField) {
	caller, traceID, spanID, extraFields := splitFields(fields)

	var builder strings.Builder
	builder.WriteString(time.Now().Format(consoleTimeFormat))
	builder.WriteByte(' ')
	builder.WriteString(level)
	builder.WriteByte(' ')
	builder.WriteString(formatValue(v))
	for _, field := range extraFields {
		builder.WriteByte(' ')
		builder.WriteString(field)
	}
	if caller != "" {
		builder.WriteByte(' ')
		builder.WriteString(caller)
	}
	builder.WriteByte('\n')

	if traceID != "" || spanID != "" {
		if traceID != "" {
			builder.WriteString("trace: ")
			builder.WriteString(traceID)
		}
		if spanID != "" {
			builder.WriteString("span: ")
			builder.WriteString(spanID)
		}
		builder.WriteByte('\n')
	}

	w.lock.Lock()
	defer w.lock.Unlock()
	_, _ = io.WriteString(out, builder.String())
}

func splitFields(fields []logx.LogField) (caller, traceID, spanID string, extraFields []string) {
	extraFields = make([]string, 0, len(fields))
	for _, field := range fields {
		value := formatValue(field.Value)
		switch field.Key {
		case "caller":
			caller = value
		case "trace":
			traceID = value
		case "span":
			spanID = value
		default:
			if field.Key == "" {
				continue
			}
			extraFields = append(extraFields, fmt.Sprintf("%s=%s", field.Key, value))
		}
	}
	return
}

func formatValue(v any) string {
	if v == nil {
		return "<nil>"
	}
	if sensitive, ok := v.(logx.Sensitive); ok {
		v = sensitive.MaskSensitive()
	}

	switch value := v.(type) {
	case string:
		return value
	case error:
		return value.Error()
	case fmt.Stringer:
		return value.String()
	case []byte:
		return string(value)
	default:
		text, err := marshalJSON(value)
		if err != nil {
			return fmt.Sprintf("%+v", value)
		}
		return text
	}
}

func marshalJSON(v any) (string, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buffer.String(), "\n"), nil
}
