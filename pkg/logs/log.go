package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger
var logLevel zap.AtomicLevel

func init() {
	logLevel = zap.NewAtomicLevel()

	config := zap.Config{
		Level:       logLevel,
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:  "ts",
			LevelKey: "level",
			NameKey:  "logger",
			//CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			//EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	log, _ = config.Build()
}

func GetLogger() *zap.Logger {
	return log
}

func SetLevel(level zapcore.Level) {
	logLevel.SetLevel(level)
}
