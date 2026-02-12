package log

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/pkg/klog"
)

func (hl *HLogger) SetLevel(level hlog.Level) {
	var lv klog.Level
	switch level {
	case hlog.LevelTrace:
		lv = klog.LevelTrace
	case hlog.LevelDebug:
		lv = klog.LevelDebug
	case hlog.LevelInfo:
		lv = klog.LevelInfo
	case hlog.LevelWarn:
		lv = klog.LevelWarn
	case hlog.LevelError:
		lv = klog.LevelError
	case hlog.LevelFatal:
		lv = klog.LevelFatal
	default:
		lv = klog.LevelWarn
	}
	hl.Logger.SetLevel(lv)
}

func WithKitex() {
	klog.SetLogger(logger)
}

type HLogger struct {
	*Logger
}

func WithHertz() {
	hl := &HLogger{
		Logger: logger,
	}
	hlog.SetLogger(hl)
}
