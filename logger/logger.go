package logger

import (
	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Init initializes the logger and replaces the global zap logger with it
func Init(debug bool) {
	var consoleEncoder, jsonEncoder zapcore.EncoderConfig

	var level zapcore.Level

	if debug {
		consoleEncoder = zap.NewDevelopmentEncoderConfig()
		jsonEncoder = zap.NewDevelopmentEncoderConfig()
		level = zap.DebugLevel
	} else {
		consoleEncoder = zap.NewProductionEncoderConfig()
		jsonEncoder = zap.NewProductionEncoderConfig()

		consoleEncoder.EncodeTime = zapcore.ISO8601TimeEncoder
		jsonEncoder.EncodeTime = zapcore.ISO8601TimeEncoder
		level = zapcore.Level(int8(viper.GetInt("log.level")))
	}

	consoleEncoder.EncodeLevel = nil

	consoleLoggerCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoder),
		zapcore.AddSync(colorable.NewColorableStdout()),
		level,
	)

	fileLoggerCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(jsonEncoder),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   viper.GetString("logger.file_path"),
			MaxSize:    viper.GetInt("log.max_size"),
			MaxBackups: viper.GetInt("log.max_backups"),
			MaxAge:     viper.GetInt("log.max_age"),
			Compress:   viper.GetBool("log.compress"),
		}),
		level,
	)

	core := zapcore.NewTee(consoleLoggerCore, fileLoggerCore)

	logger := zap.New(core, zap.WithClock(zapcore.DefaultClock))
	defer logger.Sync()

	if debug {
		logger = logger.WithOptions(zap.AddCaller())
	}

	zap.ReplaceGlobals(logger)
}

// S is a shortcut for zap.S()
func S() *zap.SugaredLogger {
	return zap.S()
}
