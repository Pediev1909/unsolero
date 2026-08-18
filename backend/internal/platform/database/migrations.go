package database

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationLockID int64 = 7_243_936_275

var migrationFilename = regexp.MustCompile(`^(\d{6})_([a-z0-9_]+)\.sql$`)

type Migration struct {
	Version  int64
	Name     string
	Filename string
	SQL      string
	Checksum string
}

type appliedMigration struct {
	name     string
	checksum string
}

func LoadMigrations(source fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(source, ".")
	if err != nil {
		return nil, fmt.Errorf("read migration directory: %w", err)
	}

	migrations := make([]Migration, 0, len(entries))
	versions := make(map[int64]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == "embed.go" {
			continue
		}

		matches := migrationFilename.FindStringSubmatch(entry.Name())
		if matches == nil {
			if entry.Name() == "README.md" {
				continue
			}
			return nil, fmt.Errorf("invalid migration filename %q", entry.Name())
		}

		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse migration version in %q: %w", entry.Name(), err)
		}
		if existing, exists := versions[version]; exists {
			return nil, fmt.Errorf("duplicate migration version %06d in %q and %q", version, existing, entry.Name())
		}

		contents, err := fs.ReadFile(source, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}
		if len(contents) == 0 {
			return nil, fmt.Errorf("migration %q is empty", entry.Name())
		}

		digest := sha256.Sum256(contents)
		migrations = append(migrations, Migration{
			Version:  version,
			Name:     matches[2],
			Filename: entry.Name(),
			SQL:      string(contents),
			Checksum: hex.EncodeToString(digest[:]),
		})
		versions[version] = entry.Name()
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// SchemaChecker verifies that the connected database exactly matches the
// migration manifest embedded in the running release. It never applies DDL.
type SchemaChecker struct {
	pool        *pgxpool.Pool
	expected    []Migration
	manifestErr error
}

func NewSchemaChecker(pool *pgxpool.Pool, source fs.FS) *SchemaChecker {
	expected, err := LoadMigrations(source)
	return &SchemaChecker{pool: pool, expected: expected, manifestErr: err}
}

func (checker *SchemaChecker) Ready(ctx context.Context) error {
	if checker.manifestErr != nil {
		return fmt.Errorf("load release migration manifest: %w", checker.manifestErr)
	}
	rows, err := checker.pool.Query(ctx, `SELECT version,name,checksum FROM platform.schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("read schema migration state: %w", err)
	}
	defer rows.Close()
	applied := make(map[int64]appliedMigration)
	for rows.Next() {
		var version int64
		var item appliedMigration
		if err := rows.Scan(&version, &item.name, &item.checksum); err != nil {
			return fmt.Errorf("scan schema migration state: %w", err)
		}
		applied[version] = item
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate schema migration state: %w", err)
	}
	if len(applied) != len(checker.expected) {
		return fmt.Errorf("schema migration count %d does not match release manifest %d", len(applied), len(checker.expected))
	}
	for _, migration := range checker.expected {
		item, exists := applied[migration.Version]
		if !exists || item.name != migration.Name || item.checksum != migration.Checksum {
			return fmt.Errorf("schema migration %06d does not match the running release", migration.Version)
		}
	}
	return nil
}

func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool, source fs.FS) error {
	migrations, err := LoadMigrations(source)
	if err != nil {
		return err
	}

	connection, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer connection.Release()

	if _, err := connection.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationLockID); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		unlockContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = connection.Exec(unlockContext, `SELECT pg_advisory_unlock($1)`, migrationLockID)
	}()

	if err := ensureMigrationTable(ctx, connection); err != nil {
		return err
	}

	applied, err := loadAppliedMigrations(ctx, connection)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if existing, exists := applied[migration.Version]; exists {
			if existing.name != migration.Name || existing.checksum != migration.Checksum {
				return fmt.Errorf("applied migration %06d does not match %q; migrations are immutable", migration.Version, migration.Filename)
			}
			continue
		}

		transaction, err := connection.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %q: %w", migration.Filename, err)
		}

		if _, err := transaction.Exec(ctx, migration.SQL); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("execute migration %q: %w", migration.Filename, err)
		}
		if _, err := transaction.Exec(
			ctx,
			`INSERT INTO platform.schema_migrations (version, name, checksum) VALUES ($1, $2, $3)`,
			migration.Version,
			migration.Name,
			migration.Checksum,
		); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("record migration %q: %w", migration.Filename, err)
		}
		if err := transaction.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %q: %w", migration.Filename, err)
		}
	}

	return nil
}

func ensureMigrationTable(ctx context.Context, connection *pgxpool.Conn) error {
	_, err := connection.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS platform;
		CREATE TABLE IF NOT EXISTS platform.schema_migrations (
			version bigint PRIMARY KEY,
			name text NOT NULL,
			checksum text NOT NULL,
			applied_at timestamptz NOT NULL DEFAULT now()
		);`)
	if err != nil {
		return fmt.Errorf("ensure migration table: %w", err)
	}
	return nil
}

func loadAppliedMigrations(ctx context.Context, connection *pgxpool.Conn) (map[int64]appliedMigration, error) {
	rows, err := connection.Query(ctx, `SELECT version, name, checksum FROM platform.schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]appliedMigration)
	for rows.Next() {
		var version int64
		var migration appliedMigration
		if err := rows.Scan(&version, &migration.name, &migration.checksum); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = migration
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}

	return applied, nil
}
