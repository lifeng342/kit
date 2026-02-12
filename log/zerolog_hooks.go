package log

import (
	"github.com/rs/zerolog"
)

// customFieldsHook implements zerolog.Hook to add custom fields from context
type customFieldsHook struct{}

func (h customFieldsHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}

	// Extract custom fields from context
	if customData, ok := ctx.Value(CustomFieldsKey).(map[string]string); ok {
		e.Interface(CustomFieldsKey, customData)
	}
}
