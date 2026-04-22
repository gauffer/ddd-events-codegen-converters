package log

import (
	"io"

	"go.uber.org/zap/zapcore"
)

type options struct {
	Writer    io.Writer
	Level     *zapcore.Level
	Formatter *Formatter
}

type Option func(o *options)
