package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

func NewZapLogger(filepath string) *zap.Logger {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
		w,
		zap.InfoLevel,
	)
	return zap.New(core)
}
