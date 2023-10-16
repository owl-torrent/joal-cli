package web

import (
	"github.com/go-stomp/stomp/v3"
	"go.uber.org/zap"
)

type zapStompLogger struct {
	delegate *zap.SugaredLogger
}

func (w *zapStompLogger) Debug(msg string) {
	w.delegate.Debug(msg)
}

func (w *zapStompLogger) Info(msg string) {
	w.delegate.Info(msg)
}

func (w *zapStompLogger) Warning(msg string) {
	w.delegate.Warn(msg)
}

func (w *zapStompLogger) Error(msg string) {
	w.delegate.Error(msg)
}

func (w *zapStompLogger) Debugf(template string, args ...interface{}) {
	w.delegate.Debugf(template, args...)
}

func (w *zapStompLogger) Infof(template string, args ...interface{}) {
	w.delegate.Infof(template, args...)
}

func (w *zapStompLogger) Warningf(template string, args ...interface{}) {
	w.delegate.Warnf(template, args...)
}

func (w *zapStompLogger) Errorf(template string, args ...interface{}) {
	w.delegate.Errorf(template, args...)
}

func wrapZapLogger(pointerToLogger *zap.Logger) stomp.Logger {
	logger := pointerToLogger.WithOptions(zap.AddCallerSkip(1))
	return &zapStompLogger{delegate: logger.Sugar()}
}
