package logger

import (
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const loggerName = "isupipe-benchmarker"

// InitZapLogger はzapロガーを初期化します
func InitZapLogger() (*zap.SugaredLogger, error) {
	nowUnix := time.Now().Unix()
	outputPath := filepath.Join("/tmp", fmt.Sprintf("isupipe-benchmarker-%d.log", nowUnix))
	errorOutputPath := filepath.Join("/tmp", fmt.Sprintf("isupipe-benchmarker-%d.err", nowUnix))

	config := zap.NewProductionConfig()
	config.Encoding = "console"
	config.DisableCaller = true
	config.DisableStacktrace = true
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// config.OutputPaths = []string{"stderr"}
	config.OutputPaths = []string{outputPath, "stderr"}
	// config.ErrorOutputPaths = []string{"stderr"}
	config.ErrorOutputPaths = []string{errorOutputPath, "stderr"}

	l, err := config.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(l.Named(loggerName))

	return zap.S(), nil
}
