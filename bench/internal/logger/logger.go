package logger

import (
	"github.com/isucon/isucon13/bench/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const loggerName = "isupipe-benchmarker"

// InitZapLogger はzapロガーを初期化します
func InitStaffLogger() (*zap.SugaredLogger, error) {
	c := zap.NewProductionConfig()
	c.Encoding = "console"
	c.DisableCaller = false
	c.DisableStacktrace = true
	c.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	c.OutputPaths = []string{config.StaffLogPath, "stderr"}
	c.ErrorOutputPaths = []string{"stderr"}
	c.Sampling = nil

	l, err := c.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(l.Named("staff-logger"))

	return zap.S(), nil
}

// InitZapLogger はzapロガーを初期化します
func InitTestLogger() (*zap.Logger, error) {
	c := zap.NewProductionConfig()
	c.Encoding = "console"
	c.DisableCaller = false
	c.DisableStacktrace = true
	c.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	c.OutputPaths = []string{"stderr"}
	c.ErrorOutputPaths = []string{"stderr"}
	c.Sampling = nil

	l, err := c.Build()
	if err != nil {
		return nil, err
	}

	return l.Named("test-logger"), nil
}

func InitContestantLogger() (*zap.Logger, error) {
	c := zap.NewProductionConfig()
	c.Encoding = "console"
	c.DisableCaller = true
	c.DisableStacktrace = true
	c.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	c.OutputPaths = []string{config.ContestantLogPath, "stdout"}
	c.ErrorOutputPaths = []string{"stdout"}
	c.Sampling = nil

	l, err := c.Build()
	if err != nil {
		return nil, err
	}

	return l.Named(loggerName), nil
}
