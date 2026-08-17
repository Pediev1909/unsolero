package database

import (
	"testing"
	"testing/fstest"
)

func TestLoadMigrationsOrdersAndChecksumsFiles(t *testing.T) {
	source := fstest.MapFS{
		"000002_second.sql": {Data: []byte("SELECT 2;")},
		"000001_first.sql":  {Data: []byte("SELECT 1;")},
		"README.md":         {Data: []byte("documentation")},
	}

	migrations, err := LoadMigrations(source)
	if err != nil {
		t.Fatalf("LoadMigrations() returned an unexpected error: %v", err)
	}
	if len(migrations) != 2 {
		t.Fatalf("len(migrations) = %d, want 2", len(migrations))
	}
	if migrations[0].Version != 1 || migrations[1].Version != 2 {
		t.Fatalf("migration order = [%d, %d], want [1, 2]", migrations[0].Version, migrations[1].Version)
	}
	if migrations[0].Checksum == "" {
		t.Fatal("expected a migration checksum")
	}
}

func TestLoadMigrationsRejectsInvalidFilename(t *testing.T) {
	source := fstest.MapFS{
		"initial.sql": {Data: []byte("SELECT 1;")},
	}

	if _, err := LoadMigrations(source); err == nil {
		t.Fatal("LoadMigrations() expected an invalid filename error")
	}
}

func TestLoadMigrationsRejectsDuplicateVersion(t *testing.T) {
	source := fstest.MapFS{
		"000001_first.sql":  {Data: []byte("SELECT 1;")},
		"000001_second.sql": {Data: []byte("SELECT 2;")},
	}

	if _, err := LoadMigrations(source); err == nil {
		t.Fatal("LoadMigrations() expected a duplicate version error")
	}
}
