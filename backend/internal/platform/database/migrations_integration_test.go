package database

import (
	"context"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/migrations"
)

func TestFailedMigrationRollsBackAndIsNotRecorded(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	const version = 999998
	_, _ = pool.Exec(ctx, `DELETE FROM platform.schema_migrations WHERE version=$1`, version)
	_, _ = pool.Exec(ctx, `DROP TABLE IF EXISTS platform.phase7_failed_migration_probe`)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM platform.schema_migrations WHERE version=$1`, version)
		_, _ = pool.Exec(cleanupCtx, `DROP TABLE IF EXISTS platform.phase7_failed_migration_probe`)
	})

	source := fstest.MapFS{"999998_failure_probe.sql": {Data: []byte(`
		CREATE TABLE platform.phase7_failed_migration_probe(id integer);
		SELECT phase7_function_that_does_not_exist();`)}}
	if err := ApplyMigrations(ctx, pool, source); err == nil {
		t.Fatal("ApplyMigrations() accepted a failing migration")
	}
	var tableExists, recordExists bool
	if err := pool.QueryRow(ctx, `SELECT to_regclass('platform.phase7_failed_migration_probe') IS NOT NULL,
		EXISTS(SELECT 1 FROM platform.schema_migrations WHERE version=$1)`, version).Scan(&tableExists, &recordExists); err != nil {
		t.Fatal(err)
	}
	if tableExists || recordExists {
		t.Fatalf("failed migration leaked table=%v record=%v", tableExists, recordExists)
	}
}

func TestSchemaCheckerFailsClosedForIncompatibleReleaseManifest(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := NewSchemaChecker(pool, migrations.Files).Ready(ctx); err != nil {
		t.Fatalf("current release schema rejected: %v", err)
	}

	entries, err := fs.ReadDir(migrations.Files, ".")
	if err != nil {
		t.Fatal(err)
	}
	incompatible := fstest.MapFS{}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "embed.go" {
			continue
		}
		contents, readErr := fs.ReadFile(migrations.Files, entry.Name())
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.HasPrefix(entry.Name(), "000001_") {
			contents = append(contents, []byte("\n-- incompatible release fixture\n")...)
		}
		incompatible[entry.Name()] = &fstest.MapFile{Data: contents}
	}
	if err := NewSchemaChecker(pool, incompatible).Ready(ctx); err == nil {
		t.Fatal("schema checker accepted a migration checksum mismatch")
	}

	missing := fstest.MapFS{}
	for name, file := range incompatible {
		if !strings.HasPrefix(name, "000017_") {
			missing[name] = file
		}
	}
	if err := NewSchemaChecker(pool, missing).Ready(ctx); err == nil {
		t.Fatal("schema checker accepted a database newer than the release manifest")
	}
}
