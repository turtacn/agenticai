// internal/logger/logger.go
package logger

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// key names for structured output
const (
	RequestIDKey = "request_id"
	SpanIDKey    = "span_id"
	TraceIDKey   = "trace_id"
	SystemKey    = "system"
)

type ctxKey struct{}

var (
	global *zap.Logger
)

// Init create a new global logger instance
// Must be called once at program bootstrap
func Init(level zapcore.Level, encoding string) error {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var sink zapcore.WriteSyncer
	sink = zapcore.AddSync(os.Stdout)

	core := zapcore.NewCore(
		func() zapcore.Encoder {
			if encoding == "console" {
				return zapcore.NewConsoleEncoder(encoderCfg)
			}
			return zapcore.NewJSONEncoder(encoderCfg)
		}(),
		sink,
		level,
	)
	opts := []zap.Option{
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}
	global = zap.New(core, opts...).Named("agenticai")
	return nil
}

// Sync flush all buffered logs
func Sync() error {
	return global.Sync()
}

// WithCtx returns a contextual logger enriched with trace/span/request ids
func WithCtx(ctx context.Context) *zap.Logger {
	if ctx == nil {
		ctx = context.Background()
	}
	l := global
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		l = l.With(
			zap.String(TraceIDKey, spanCtx.TraceID().String()),
			zap.String(SpanIDKey, spanCtx.SpanID().String()),
		)
	}
	if v := ctx.Value(ctxKey{}); v != nil {
		if rid, ok := v.(string); ok {
			l = l.With(zap.String(RequestIDKey, rid))
		}
	}
	return l.With(zap.String(SystemKey, "agenticai"))
}

// SetRequestID injects request-id into ctx
func SetRequestID(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, ctxKey{}, rid)
}

//
// leveled convenience wrappers
//
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	WithCtx(ctx).Debug(msg, fields...)
}
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	WithCtx(ctx).Info(msg, fields...)
}
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	WithCtx(ctx).Warn(msg, fields...)
}
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	WithCtx(ctx).Error(msg, fields...)
}
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	WithCtx(ctx).Fatal(msg, fields...)
}

// Sugar helpers if required
func Errorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	WithCtx(ctx).Sugar().Errorw(msg, keysAndValues...)
}
//Personal.AI order the ending
