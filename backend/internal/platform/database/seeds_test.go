package database

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestApplySeedRejectsMissingFileBeforeDatabaseAccess(t *testing.T) {
	if err := ApplySeed(context.Background(), nil, fstest.MapFS{}, "demo.sql"); err == nil {
		t.Fatal("ApplySeed() expected an error for a missing seed file")
	}
}

func TestApplySeedRejectsEmptyFileBeforeDatabaseAccess(t *testing.T) {
	source := fstest.MapFS{"demo.sql": {Data: []byte{}}}
	if err := ApplySeed(context.Background(), nil, source, "demo.sql"); err == nil {
		t.Fatal("ApplySeed() expected an error for an empty seed file")
	}
}
