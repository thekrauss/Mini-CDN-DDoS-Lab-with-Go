package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger(level string, format string) {
	var zapCfg zap.Config

	if format == "json" {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
	}

	switch level {
	case "debug":
		zapCfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "warn":
		zapCfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		zapCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}

	var err error
	Log, err = zapCfg.Build()
	if err != nil {
		panic("⚠️ Logger init failed: " + err.Error())
	}
}
