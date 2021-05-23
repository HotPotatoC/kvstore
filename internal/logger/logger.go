package logger

import (
	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init initializes the global logger
func Init(debug bool) {
	atom := zap.NewAtomicLevel()

	if debug {
		atom.SetLevel(zap.DebugLevel)
	} else {
		atom.SetLevel(zap.InfoLevel)
	}

	encoder := zap.NewDevelopmentEncoderConfig()
	encoder.EncodeLevel = nil

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoder),
		zapcore.AddSync(colorable.NewColorableStdout()),
		atom,
	))
	defer logger.Sync()

	if debug {
		logger = logger.WithOptions(zap.AddCaller())
	}

	zap.ReplaceGlobals(logger)
}

func L() *zap.SugaredLogger {
	return zap.L().Sugar()
}
