package observability

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"rigmark/internal/platform/database"
)

func NewJSONLogger(destination io.Writer, level slog.Leveler) *slog.Logger {
	handler := slog.NewJSONHandler(destination, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: replaceSensitiveAttribute,
	})
	return slog.New(handler)
}

func replaceSensitiveAttribute(_ []string, attribute slog.Attr) slog.Attr {
	key := strings.ToLower(attribute.Key)
	for _, fragment := range []string{
		"password", "secret", "token", "authorization", "cookie", "email", "raw_ip",
		"user_agent", "destination", "webhook_body", "payload", "free_text", "recovery_code",
		"database_url", "affiliate_url", "api_key", "private_key", "credential", "dsn",
	} {
		if strings.Contains(key, fragment) {
			return slog.String(attribute.Key, "[redacted]")
		}
	}
	if attribute.Value.Kind() == slog.KindAny {
		if err, ok := attribute.Value.Any().(error); ok {
			if class := database.ClassifyError(err); class != database.ErrorUnknown {
				return slog.String(attribute.Key, "database."+string(class))
			}
			if err == context.Canceled || err == context.DeadlineExceeded {
				return slog.String(attribute.Key, err.Error())
			}
			return slog.String(attribute.Key, fmt.Sprintf("%T", err))
		}
	}
	return attribute
}
