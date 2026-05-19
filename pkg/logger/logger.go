package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

var globalLogger *Logger

func Init(environment string) *Logger {
	var config zap.Config
	if environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	l, err := config.Build()
	if err != nil {
		panic(err)
	}

	globalLogger = &Logger{l}
	return globalLogger
}

func L() *Logger {
	if globalLogger == nil {
		globalLogger = &Logger{zap.NewNop()}
	}
	return globalLogger
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.Logger.Sugar().Infow(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	l.Logger.Sugar().Errorw(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.Logger.Sugar().Debugw(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.Logger.Sugar().Warnw(msg, fields...)
}

func Sync() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
}