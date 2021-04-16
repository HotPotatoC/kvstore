package logger

import (
	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new logger
func New() *zap.SugaredLogger {
	atom := zap.NewAtomicLevel()

	atom.SetLevel(zap.InfoLevel)

	encoder := zap.NewDevelopmentEncoderConfig()
	encoder.EncodeLevel = nil

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoder),
		zapcore.AddSync(colorable.NewColorableStdout()),
		atom,
	))
	defer logger.Sync()

	sugar := logger.Sugar()

	return sugar
}
