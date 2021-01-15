package logs

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var encoderConfig = zapcore.EncoderConfig{
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

var log *zap.Logger
var logLevel = zap.NewAtomicLevelAt(zap.InfoLevel)

func init() {
	ws, _, err := zap.Open("stdout")
	if err != nil {
		panic(err)
	}

	log = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
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

// This function replace the globally available logger, it must be called at the beginning of the program to prevent any concurrency issues
func ReplaceLogger(config *LogConfig) error {
	ws, _, err := zap.Open(config.OutputPaths...)
	if err != nil {
		return err
	}

	// replace the current log level
	err = logLevel.UnmarshalText([]byte(config.Level))
	if err != nil {
		return errors.Wrapf(err, "failed to parse log level '%s'", config.Level)
	}

	log = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			ws,
			logLevel,
		),
		zap.Development(),
	)
	return nil
}

type LogConfig struct {
	Level       string   `yaml:"level"`
	OutputPaths []string `yaml:"outputPaths"`
}

func (c LogConfig) Default() *LogConfig {
	return &LogConfig{
		Level: "info",
		OutputPaths: []string{
			"stdout",
		},
	}
}
