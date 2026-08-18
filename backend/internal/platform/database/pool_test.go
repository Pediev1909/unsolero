package database

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorClass
	}{
		{"canceled", context.Canceled, ErrorCanceled},
		{"timeout", context.DeadlineExceeded, ErrorTimeout},
		{"statement timeout", &pgconn.PgError{Code: "57014"}, ErrorTimeout},
		{"lock timeout", &pgconn.PgError{Code: "55P03"}, ErrorTimeout},
		{"deadlock", &pgconn.PgError{Code: "40P01"}, ErrorDeadlock},
		{"serialization", &pgconn.PgError{Code: "40001"}, ErrorSerialization},
		{"constraint", &pgconn.PgError{Code: "23505"}, ErrorConstraint},
		{"unknown", errors.New("do not expose this value"), ErrorUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ClassifyError(test.err); got != test.want {
				t.Fatalf("ClassifyError()=%q want %q", got, test.want)
			}
		})
	}
}

func TestOpenPoolAppliesOperationalTimeouts(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	pool, err := OpenPool(context.Background(), PoolConfig{
		URL: databaseURL, ApplicationName: "phase7-pool-test", MaxConnections: 2,
		MaxConnectionLifetime: time.Minute, MaxConnectionIdleTime: time.Minute,
		HealthCheckPeriod: time.Minute, ConnectTimeout: 5 * time.Second,
		StatementTimeout: 7 * time.Second, LockTimeout: 3 * time.Second,
		IdleTransactionTimeout: 11 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var applicationName, statementTimeout, lockTimeout, idleTimeout string
	if err := pool.QueryRow(context.Background(), `SELECT current_setting('application_name'),current_setting('statement_timeout'),current_setting('lock_timeout'),current_setting('idle_in_transaction_session_timeout')`).Scan(&applicationName, &statementTimeout, &lockTimeout, &idleTimeout); err != nil {
		t.Fatal(err)
	}
	if applicationName != "phase7-pool-test" || statementTimeout != "7s" || lockTimeout != "3s" || idleTimeout != "11s" {
		t.Fatalf("settings=%q %q %q %q", applicationName, statementTimeout, lockTimeout, idleTimeout)
	}
}

func TestClassifiedErrorDoesNotExposeDatabaseDetails(t *testing.T) {
	err := ClassifiedError(&pgconn.PgError{Code: "23505", Detail: "email=private@example.test"})
	if got := err.Error(); got != "database constraint_violation" {
		t.Fatalf("ClassifiedError()=%q", got)
	}
}

func TestPoolExhaustionHonorsCallerDeadlineAndRecovers(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	pool, err := OpenPool(context.Background(), PoolConfig{
		URL: databaseURL, ApplicationName: "phase8-pool-exhaustion", MaxConnections: 1,
		MaxConnectionLifetime: time.Minute, MaxConnectionIdleTime: time.Minute,
		HealthCheckPeriod: time.Minute, ConnectTimeout: 5 * time.Second,
		StatementTimeout: 5 * time.Second, LockTimeout: 2 * time.Second,
		IdleTransactionTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	held, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	deadlineContext, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	var value int
	err = pool.QueryRow(deadlineContext, `SELECT 1`).Scan(&value)
	if !errors.Is(err, context.DeadlineExceeded) || ClassifyError(err) != ErrorTimeout {
		t.Fatalf("exhausted pool error=%v class=%s", err, ClassifyError(err))
	}
	held.Release()
	if err := pool.QueryRow(context.Background(), `SELECT 1`).Scan(&value); err != nil || value != 1 {
		t.Fatalf("pool did not recover value=%d error=%v", value, err)
	}
}

func TestStatementTimeoutCancelsSlowQuery(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	pool, err := OpenPool(context.Background(), PoolConfig{
		URL: databaseURL, ApplicationName: "phase8-statement-timeout", MaxConnections: 1,
		MaxConnectionLifetime: time.Minute, MaxConnectionIdleTime: time.Minute,
		HealthCheckPeriod: time.Minute, ConnectTimeout: 5 * time.Second,
		StatementTimeout: 100 * time.Millisecond, LockTimeout: 100 * time.Millisecond,
		IdleTransactionTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	started := time.Now()
	_, err = pool.Exec(context.Background(), `SELECT pg_sleep(1)`)
	if ClassifyError(err) != ErrorTimeout {
		t.Fatalf("slow query error=%v class=%s", err, ClassifyError(err))
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("statement timeout took %s", elapsed)
	}
}
