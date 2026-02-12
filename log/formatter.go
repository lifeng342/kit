// Custom logger format for minimax server development
// Log format refer to https://vrfi1sk8a0.feishu.cn/wiki/wikcnVnPvFHe9Cn4QszqHibGzCf

package log

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/sirupsen/logrus"
)

const (
	defaultTimestampFormat = "2006-01-02 15:04:05.000"
	levenLen               = 7
	placeholder            = "-"
)

var LevelStr = [7]string{}

func init() {
	for _, l := range logrus.AllLevels {
		level := strings.ToUpper(l.String())
		padding := strings.Repeat(" ", levenLen-len(level))
		LevelStr[l] = padding + level
	}
}

// Formatter implements logrus.Formatter interface.
type Formatter struct {
}

const CustomFieldsKey = "ctx_extra_data"
const TraceIDKey = "trace_id"

// Format building log message.
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	logTime := entry.Time.Format(defaultTimestampFormat)
	level := LevelStr[entry.Level]

	traceId := entry.Data[TraceIDKey]
	if traceId == nil {
		if entry.Context != nil && entry.Context.Value(TraceIDKey) != nil {
			traceId = entry.Context.Value(TraceIDKey)
		} else {
			traceId = placeholder
		}
	}

	depth := 8
	if entry.Context == nil {
		depth = 9
	}
	_, file, line, _ := runtime.Caller(depth)
	caller := fmt.Sprintf("%v:%v", file, line)

	msg := entry.Message
	pid := GetPID()
	gid := GetGID()
	custom := "{}"
	var customMap map[string]string
	if entry.Context != nil {
		if m, ok := entry.Context.Value(CustomFieldsKey).(map[string]string); ok {
			customMap = m
		}
	}
	if customMap != nil {
		bytes, _ := sonic.Marshal(customMap)
		custom = string(bytes)
	}
	// time, level, pid, thread id, trace_id, file_loc, :, context info(opt), msg(opt)
	output := fmt.Sprintf("%v %v %v %v %v %v %v : %v\n", logTime, level, pid, gid, traceId, caller, custom, msg)
	return []byte(output), nil
}
