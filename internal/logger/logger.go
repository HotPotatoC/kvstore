package logger

import (
	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Init initializes the global logger
func Init(debug bool) {
	atom := zap.NewAtomicLevel()

	if debug {
		atom.SetLevel(zap.DebugLevel)
	} else {
		atom.SetLevel(zap.InfoLevel)
	}

	var consoleEncoder, jsonEncoder zapcore.EncoderConfig

	if debug {
		consoleEncoder = zap.NewDevelopmentEncoderConfig()
		jsonEncoder = zap.NewDevelopmentEncoderConfig()
	} else {
		consoleEncoder = zap.NewProductionEncoderConfig()
		jsonEncoder = zap.NewProductionEncoderConfig()

		consoleEncoder.EncodeTime = zapcore.ISO8601TimeEncoder
		jsonEncoder.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	consoleEncoder.EncodeLevel = nil

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoder),
			zapcore.AddSync(colorable.NewColorableStdout()),
			atom,
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(jsonEncoder),
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   viper.GetString("log.path"),
				MaxSize:    10,
				MaxBackups: 6,
				MaxAge:     28,
				Compress:   true,
			}),
			zap.DebugLevel,
		),
	)

	logger := zap.New(core)
	defer logger.Sync()

	if debug {
		logger = logger.WithOptions(zap.AddCaller())
	}

	zap.ReplaceGlobals(logger)
}

func S() *zap.SugaredLogger {
	return zap.S()
}
