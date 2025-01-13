package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
}

type logger struct {
	logger *zap.Logger
}

func NewLogger(logfile string, verbose bool) Logger {

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	logFile, _ := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)

	var core zapcore.Core

	switch verbose {
	case true:
		core = zapcore.NewTee(
			zapcore.NewCore(fileEncoder, writer, zapcore.InfoLevel),
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		)
	case false:
		core = zapcore.NewCore(fileEncoder, writer, zapcore.WarnLevel)
	}

	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return &logger{
		logger: l,
	}
}

func (l *logger) Debug(msg string) {
	l.logger.Debug(msg)
}

func (l *logger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *logger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *logger) Error(msg string) {
	l.logger.Error(msg)
}

func (l *logger) Fatal(msg string) {
	l.logger.Fatal(msg)
}
