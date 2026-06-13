package log

import (
	"go.uber.org/zap"
)

func NewLogger() (*zap.SugaredLogger, func()) {
	config := zap.NewProductionConfig()
	
	logger, _ := config.Build(zap.AddStacktrace(zap.DPanicLevel))
	

	sugaredLogger := logger.Sugar()

	cleanup := func() {
		_ = sugaredLogger.Sync()
	}

	return sugaredLogger, cleanup
}