package testutil

import (
	"bufio"
	"bytes"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewTestLoger creates a new logger for testing.
// nolint:gocritic
func NewTestLoger() (*zap.SugaredLogger, func()) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writer := bufio.NewWriter(&buf)
	logger := zap.New(
		zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel),
		zap.WithCaller(true),
	).Sugar()

	return logger, func() {
		_ = logger.Sync()

		_ = writer.Flush()
		fmt.Println("******************** LOGS ********************")
		fmt.Print(buf.String())
		fmt.Println("******************** LOGS ********************")
	}
}
