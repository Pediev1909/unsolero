package observability

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestJSONLoggerRedactsSecretsAndClassifiesErrors(t *testing.T) {
	var output bytes.Buffer
	logger := NewJSONLogger(&output, slog.LevelDebug)
	logger.Error("operation failed",
		"session_token", "secret-session-value",
		"email", "private@example.test",
		"database_url", "postgres://user:password@database/private",
		"api_key", "provider-key-value",
		"error", &pgconn.PgError{Code: "23505", Detail: "private@example.test"})
	logged := output.String()
	for _, forbidden := range []string{"secret-session-value", "private@example.test", "provider-key-value", "postgres://", "23505"} {
		if strings.Contains(logged, forbidden) {
			t.Fatalf("log exposed %q: %s", forbidden, logged)
		}
	}
	for _, expected := range []string{"[redacted]", "database.constraint_violation"} {
		if !strings.Contains(logged, expected) {
			t.Fatalf("log missing %q: %s", expected, logged)
		}
	}
}

func TestJSONLoggerDoesNotRenderArbitraryErrorText(t *testing.T) {
	var output bytes.Buffer
	NewJSONLogger(&output, slog.LevelInfo).Error("failed", "error", errors.New("password=unsafe"))
	if strings.Contains(output.String(), "password=unsafe") {
		t.Fatalf("arbitrary error text leaked: %s", output.String())
	}
}
