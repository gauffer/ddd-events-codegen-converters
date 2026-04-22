// Package log — структурированный логгер на базе zap
package log

import (
	"context"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Formatter string
)

const (
	FormatterText Formatter = "text"
	FormatterJSON Formatter = "json"
)

const envLogLevel = "LOG_LEVEL"

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	WithContext(ctx context.Context) Logger
	// Zap — доступ к *zap.Logger для совместимости с kafka-go
	Zap() *zap.Logger
}

type logger struct {
	l      *zap.Logger
	fields []zap.Field
}

func NewLogger(o ...Option) Logger {
	var opts options
	for _, opt := range o {
		opt(&opts)
	}

	var encConfig zapcore.EncoderConfig
	if opts.Formatter == nil || *opts.Formatter == FormatterJSON {
		encConfig = zap.NewProductionEncoderConfig()
		encConfig.MessageKey = "message"
		encConfig.TimeKey = "@timestamp"
		encConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		encConfig.CallerKey = "line_number"
		encConfig.EncodeLevel = func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
			if level == zapcore.WarnLevel {
				encoder.AppendString("warning")
				return
			}
			zapcore.LowercaseLevelEncoder(level, encoder)
		}
	} else {
		encConfig = zap.NewDevelopmentEncoderConfig()
		encConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var encoder zapcore.Encoder
	if opts.Formatter == nil || *opts.Formatter == FormatterJSON {
		encoder = newNoDuplicateEncoder(zapcore.NewJSONEncoder(encConfig),
			"message", "@timestamp", "line_number", "level",
			"_id", "_index", "_score", "_type", "source_type", "stream",
			"infra_index", "agent", "region", "host",
			"kubernetes.pod.name", "kubernetes.node.name", "kubernetes.namespace")
	} else {
		encoder = zapcore.NewConsoleEncoder(encConfig)
	}

	var writer io.Writer
	if opts.Writer != nil {
		writer = opts.Writer
	} else {
		writer = os.Stdout
	}

	var level zapcore.Level
	if opts.Level != nil {
		level = *opts.Level
	} else if l, err := zapcore.ParseLevel(os.Getenv(envLogLevel)); err == nil {
		level = l
	} else {
		level = zap.InfoLevel
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(writer),
		level,
	)
	zapLogger := zap.
		New(core).
		WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))

	return &logger{l: zapLogger, fields: []zap.Field{}}
}

type ctxFields struct{}

func NewContext(parent context.Context, fields ...zap.Field) context.Context {
	val := parent.Value(ctxFields{})
	if val == nil {
		return context.WithValue(parent, ctxFields{}, fields)
	}

	ctxValue, ok := val.([]zap.Field)
	if !ok {
		return context.WithValue(parent, ctxFields{}, fields)
	}

	return context.WithValue(parent, ctxFields{}, append(ctxValue, fields...))
}

func (l *logger) Info(msg string, fields ...zap.Field) {
	l.l.Info(msg, fields...)
}

func (l *logger) Error(msg string, fields ...zap.Field) {
	l.l.Error(msg, fields...)
}

func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.l.Debug(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.l.Warn(msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...zap.Field) {
	l.l.Fatal(msg, fields...)
}

func (l *logger) With(fields ...zap.Field) Logger {
	return &logger{l: l.l.With(fields...), fields: append(l.fields, fields...)}
}

func (l *logger) WithContext(ctx context.Context) Logger {
	val := ctx.Value(ctxFields{})
	if val == nil {
		return l
	}

	fields, ok := val.([]zap.Field)
	if !ok {
		return l
	}

	return &logger{l: l.l, fields: append(fields, l.fields...)}
}

func (l *logger) Zap() *zap.Logger {
	return l.l
}
