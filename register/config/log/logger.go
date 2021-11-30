package log

import (
	"strings"
	"time"

	"github.com/galaxy-future/schedulx/register/config"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func Init() {
	Logger = createZapLogger().Sugar()
}

func createZapLogger() *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	var recordTimeFormat = "2006-01-02 15:04:05.000"
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(recordTimeFormat))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.TimeKey = "created_at"
	var encoder = zapcore.NewConsoleEncoder(encoderConfig)

	lumberJackLogger := &lumberjack.Logger{
		Filename:   config.GlobalConfig.LogFile,
		MaxSize:    200,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   false,
	}
	writer := zapcore.AddSync(lumberJackLogger)
	logLevel := zap.InfoLevel
	switch strings.ToLower(config.GlobalConfig.LogLevel) {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	}
	zapCore := zapcore.NewCore(encoder, writer, logLevel)
	return zap.New(zapCore, zap.AddCaller())
}
