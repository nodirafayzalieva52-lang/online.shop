package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func (l *Logger) Infof(s string, userRole string, authID int64) {
	panic("unimplemented")
}

func New(devMode bool) (*Logger, error) {
	var cfg zap.Config
	if devMode {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.OutputPaths = []string{"stdout"}
	mainLogger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("mainLogger.Build: %w", err)
	}

	return &Logger{
		mainLogger,
	}, nil
}
