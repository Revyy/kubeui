package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// NewZapLogger creates a zap  in development mode that logs to the file specified by filepath.
// It uses a lumberjack rolling file logger with the following settings:
// MaxSize: 500
// MaxBackups: 3
// MaxAge: 28
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
