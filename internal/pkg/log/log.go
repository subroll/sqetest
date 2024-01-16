package log

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	loggerOpts = []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.ErrorLevel)}
	encoder    = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     RFC3339NanoEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	output   zapcore.WriteSyncer
	logLevel zapcore.Level
	log      *zap.Logger
	logMu    sync.Mutex
)

func init() {
	defaultLog()
}

func defaultLog() {
	logMu.Lock()
	defer logMu.Unlock()

	output = os.Stderr
	logLevel = zap.InfoLevel
	createLogger(output, logLevel)
}

func createLogger(o zapcore.WriteSyncer, lvl zapcore.Level) {
	output = o
	core := zapcore.NewCore(encoder, output, lvl)
	log = zap.New(core).WithOptions(loggerOpts...)
}

// Info add log entry with or without fields to info level
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Warn add log entry with or without fields to warn level
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Error add log entry with or without fields to error level
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Fatal add log entry with or without fields to fatal level
func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}
