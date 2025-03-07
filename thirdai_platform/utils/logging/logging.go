package logging

import (
	"log/slog"
)

type LogCode string

const (
	// SYSTEM EVENTS (SYSTEM*)
	SYSTEM LogCode = "SYSTEM"

	// DATA OPERATIONS (DATA*)
	FILE_VALIDATION LogCode = "FILE_VALIDATION"

	// MODEL OPERATIONS (MODEL*)
	MODEL_INFO    LogCode = "MODEL_INFO"
	MODEL_SAVE    LogCode = "MODEL_SAVE"
	MODEL_LOAD    LogCode = "MODEL_LOAD"
	MODEL_INIT    LogCode = "MODEL_INIT"
	MODEL_PREDICT LogCode = "MODEL_PREDICT"
	MODEL_TRAIN   LogCode = "MODEL_TRAIN"
	MODEL_EVAL    LogCode = "MODEL_EVAL"

	// NDB Specific Operations
	MODEL_INSERT LogCode = "MODEL_INSERT"
	MODEL_DELETE LogCode = "MODEL_DELETE"
	MODEL_RLHF   LogCode = "MODEL_RLHF"
	MODEL_SEARCH LogCode = "MODEL_SEARCH"
)

// VictoriaLogs has fixed field name for time (_time) and message(_msg). This function maps fields msg -> _msg and time -> _time.
func convertKeysToVictoriaLogs(keys []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{Key: "_time", Value: slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05"))}
	}
	if a.Key == slog.MessageKey {
		return slog.Attr{Key: "_msg", Value: a.Value}
	}
	return a
}

func GetVictoriaLogsOptions(addSource bool) *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: convertKeysToVictoriaLogs,
		AddSource:   addSource,
	}
}
