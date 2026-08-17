package database

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const seedLockID int64 = 7_243_936_276

func ApplySeed(ctx context.Context, pool *pgxpool.Pool, source fs.FS, filename string) error {
	contents, err := fs.ReadFile(source, filename)
	if err != nil {
		return fmt.Errorf("read seed %q: %w", filename, err)
	}
	if len(contents) == 0 {
		return fmt.Errorf("seed %q is empty", filename)
	}

	connection, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire seed connection: %w", err)
	}
	defer connection.Release()

	if _, err := connection.Exec(ctx, `SELECT pg_advisory_lock($1)`, seedLockID); err != nil {
		return fmt.Errorf("acquire seed lock: %w", err)
	}
	defer func() {
		unlockContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = connection.Exec(unlockContext, `SELECT pg_advisory_unlock($1)`, seedLockID)
	}()

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	if _, err := transaction.Exec(ctx, string(contents)); err != nil {
		_ = transaction.Rollback(ctx)
		return fmt.Errorf("execute seed %q: %w", filename, err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed %q: %w", filename, err)
	}

	return nil
}
