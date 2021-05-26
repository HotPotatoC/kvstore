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
			zapcore.Level(int8(viper.GetInt("log.level"))),
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(jsonEncoder),
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   viper.GetString("log.path"),
				MaxSize:    viper.GetInt("log.max_size"),
				MaxBackups: viper.GetInt("log.max_backups"),
				MaxAge:     viper.GetInt("log.max_age"),
				Compress:   viper.GetBool("log.compress"),
			}),
			zapcore.Level(int8(viper.GetInt("log.level"))),
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
