package logger

import (
	"github.com/isucon/isucon13/bench/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const loggerName = "isupipe-benchmarker"

// InitZapLogger はzapロガーを初期化します
func InitZapLogger() (*zap.SugaredLogger, error) {
	c := zap.NewProductionConfig()
	c.Encoding = "console"
	c.DisableCaller = true
	c.DisableStacktrace = true
	c.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	c.OutputPaths = []string{config.LogPath, "stderr"}
	c.ErrorOutputPaths = []string{"stderr"}

	l, err := c.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(l.Named(loggerName))

	return zap.S(), nil
}
