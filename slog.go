package panylcli

import (
	"context"
	"log/slog"
)

type slogContextKey string

const (
	// LoggerCtxKey is the string used to extract logger
	slogLoggerCtxKey slogContextKey = "logger-cli"
)

var emptySLogger *slog.Logger

func SLogCLIToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, slogLoggerCtxKey, logger)
}

func SLogCLIFromContext(ctx context.Context) *slog.Logger {
	v, ok := ctx.Value(slogLoggerCtxKey).(*slog.Logger)
	if ok {
		return v
	}
	return emptySLogger
}

func init() {
	emptySLogger = slog.New(&discardHandler{})
}

type discardHandler struct{}

func (dh discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (dh discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (dh discardHandler) WithAttrs(attrs []slog.Attr) slog.Handler  { return dh }
func (dh discardHandler) WithGroup(name string) slog.Handler        { return dh }
