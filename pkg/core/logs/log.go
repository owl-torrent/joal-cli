package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	level            zapcore.Level
	OutputPaths      []string
	ErrorOutputPaths []string
}

var log *zap.Logger
var logLevel = zap.NewAtomicLevelAt(zap.InfoLevel)

func init() {
	conf := zapcore.EncoderConfig{
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
	}

	ws, _, err := zap.Open("stdout")
	if err != nil {
		panic(err)
	}

	log = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(conf),
			ws,
			logLevel,
		),
		zap.Development(),
	)
}

func GetLogger() *zap.Logger {
	return log
}

func SetLevel(level zapcore.Level) {
	logLevel.SetLevel(level)
}
