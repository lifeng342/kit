package log

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog"
)

// customWriter wraps an io.Writer and formats zerolog JSON output
// into the custom format: time level pid gid trace_id caller custom : msg
type customWriter struct {
	out          io.Writer
	enableMetric bool
}

func newCustomWriter(w io.Writer) *customWriter {
	return &customWriter{
		out:          w,
		enableMetric: false,
	}
}

func (w *customWriter) enableMetrics() {
	w.enableMetric = true
}

func (w *customWriter) Write(p []byte) (n int, err error) {
	// Parse zerolog JSON output using sonic (faster than encoding/json)
	var logEntry map[string]interface{}
	if err = sonic.Unmarshal(p, &logEntry); err != nil {
		// If parsing fails, write original content
		return w.out.Write(p)
	}

	// Extract fields from JSON
	logTime := formatTime(logEntry[zerolog.TimestampFieldName])
	level := formatLevel(logEntry[zerolog.LevelFieldName])
	msg := getString(logEntry[zerolog.MessageFieldName])

	// Try multiple possible trace ID field names
	traceId := getString(logEntry[TraceIDKey])
	if traceId == "" {
		traceId = placeholder
	}

	// Use caller from zerolog (configured with CallerWithSkipFrameCount)
	caller := getString(logEntry[zerolog.CallerFieldName])
	if caller == "" {
		caller = placeholder
	}

	pid := GetPID()
	gid := GetGID()

	// Extract custom fields
	custom := "{}"
	if customData, ok := logEntry[CustomFieldsKey]; ok {
		if bytes, err := sonic.Marshal(customData); err == nil {
			custom = string(bytes)
		}
	}

	// Update metrics if enabled
	if w.enableMetric {
		levelStr := getString(logEntry[zerolog.LevelFieldName])
		w.updateMetrics(levelStr)
	}

	// Format output
	output := fmt.Sprintf("%v %v %v %v %v %v %v : %v\n",
		logTime, level, pid, gid, traceId, caller, custom, msg)

	return w.out.Write([]byte(output))
}

func (w *customWriter) updateMetrics(levelStr string) {
	// Count error, warn, fatal, panic logs
	switch levelStr {
	case "warn", "error", "fatal", "panic":
		errLogCounter.WithLabelValues(strings.ToUpper(levelStr)).Add(1)
	}
}

// formatTime formats the timestamp from zerolog
func formatTime(t interface{}) string {
	if t == nil {
		return time.Now().Format(defaultTimestampFormat)
	}

	switch v := t.(type) {
	case string:
		// Parse ISO8601 format from zerolog
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			return parsed.Format(defaultTimestampFormat)
		}
		return v
	case float64:
		// Unix timestamp
		return time.Unix(int64(v), 0).Format(defaultTimestampFormat)
	default:
		return time.Now().Format(defaultTimestampFormat)
	}
}

// formatLevel formats the log level with padding
func formatLevel(l interface{}) string {
	if l == nil {
		return "   INFO"
	}

	var levelStr string
	switch v := l.(type) {
	case string:
		levelStr = strings.ToUpper(v)
	case float64:
		// zerolog uses numeric levels: trace=-1, debug=0, info=1, warn=2, error=3, fatal=4, panic=5
		levels := []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC"}
		idx := int(v) + 1
		if idx >= 0 && idx < len(levels) {
			levelStr = levels[idx]
		} else {
			levelStr = "UNKNOWN"
		}
	default:
		levelStr = "INFO"
	}

	// No padding - let level length be variable
	return levelStr
}

// getString safely extracts string from interface{}
func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
