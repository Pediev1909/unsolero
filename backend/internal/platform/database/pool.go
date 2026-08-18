package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolConfig struct {
	URL                    string
	ApplicationName        string
	MaxConnections         int32
	MinConnections         int32
	MaxConnectionLifetime  time.Duration
	MaxConnectionIdleTime  time.Duration
	HealthCheckPeriod      time.Duration
	ConnectTimeout         time.Duration
	StatementTimeout       time.Duration
	LockTimeout            time.Duration
	IdleTransactionTimeout time.Duration
}

func OpenPool(ctx context.Context, config PoolConfig) (*pgxpool.Pool, error) {
	parsed, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, errors.New("parse PostgreSQL configuration")
	}
	parsed.MaxConns = config.MaxConnections
	parsed.MinConns = config.MinConnections
	parsed.MaxConnLifetime = config.MaxConnectionLifetime
	parsed.MaxConnIdleTime = config.MaxConnectionIdleTime
	parsed.HealthCheckPeriod = config.HealthCheckPeriod
	parsed.ConnConfig.ConnectTimeout = config.ConnectTimeout
	if parsed.ConnConfig.RuntimeParams == nil {
		parsed.ConnConfig.RuntimeParams = make(map[string]string)
	}
	parsed.ConnConfig.RuntimeParams["application_name"] = config.ApplicationName
	parsed.ConnConfig.RuntimeParams["statement_timeout"] = milliseconds(config.StatementTimeout)
	parsed.ConnConfig.RuntimeParams["lock_timeout"] = milliseconds(config.LockTimeout)
	parsed.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = milliseconds(config.IdleTransactionTimeout)

	pool, err := pgxpool.NewWithConfig(ctx, parsed)
	if err != nil {
		return nil, errors.New("create PostgreSQL pool")
	}
	connectCtx, cancel := context.WithTimeout(ctx, config.ConnectTimeout)
	defer cancel()
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", ClassifiedError(err))
	}
	return pool, nil
}

func milliseconds(value time.Duration) string {
	return strconv.FormatInt(value.Milliseconds(), 10)
}

type ErrorClass string

const (
	ErrorCanceled      ErrorClass = "canceled"
	ErrorTimeout       ErrorClass = "timeout"
	ErrorUnavailable   ErrorClass = "unavailable"
	ErrorDeadlock      ErrorClass = "deadlock"
	ErrorSerialization ErrorClass = "serialization_failure"
	ErrorConstraint    ErrorClass = "constraint_violation"
	ErrorUnknown       ErrorClass = "unknown"
)

func ClassifyError(err error) ErrorClass {
	if errors.Is(err, context.Canceled) {
		return ErrorCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorTimeout
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "40P01":
			return ErrorDeadlock
		case "40001":
			return ErrorSerialization
		case "55P03", "57014":
			return ErrorTimeout
		case "23502", "23503", "23505", "23514", "23P01":
			return ErrorConstraint
		case "57P01", "57P02", "57P03", "08000", "08001", "08003", "08004", "08006", "08007", "08P01":
			return ErrorUnavailable
		}
	}
	var networkError *pgconn.ConnectError
	if errors.As(err, &networkError) {
		return ErrorUnavailable
	}
	return ErrorUnknown
}

type classifiedError struct{ class ErrorClass }

func (err classifiedError) Error() string { return "database " + string(err.class) }

func ClassifiedError(err error) error {
	if err == nil {
		return nil
	}
	return classifiedError{class: ClassifyError(err)}
}
