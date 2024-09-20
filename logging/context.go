package logging

import (
	"context"
	"log/slog"
)

var ContextKey = contextKeyType{}

type contextKeyType struct{}

func (contextKeyType) String() string {
	return "mvp/logging.ContextKey"
}

func From(ctx context.Context) *slog.Logger {
	if v := ctx.Value(ContextKey); v != nil {
		return v.(*slog.Logger)
	}
	return slog.Default()
}
