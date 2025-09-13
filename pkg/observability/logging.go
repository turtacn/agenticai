// pkg/observability/logging.go
package observability

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var baseLog *zap.SugaredLogger

func InitLog(level string) (func(), error) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if level == "debug" {
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	baseLog = l.Sugar()
	return func() { _ = l.Sync() }, nil
}

func FromCtx(ctx context.Context) *zap.SugaredLogger {
	if baseLog == nil {
		l, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}
		baseLog = l.Sugar()
	}
	return baseLog.With(zap.String("traceID", getTrace(ctx)))
}

func getTrace(ctx context.Context) string {
	return "todo/trace"
}
//Personal.AI order the ending
