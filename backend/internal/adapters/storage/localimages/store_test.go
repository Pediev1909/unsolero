package localimages

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestStoreLifecycleAndPathValidation(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	name, err := store.Save(context.Background(), []byte("image-data"), ".webp")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if !strings.HasSuffix(name, ".webp") {
		t.Fatalf("name = %q", name)
	}
	data, contentType, err := store.Open(context.Background(), name)
	if err != nil || string(data) != "image-data" || contentType != "image/webp" {
		t.Fatalf("Open() = (%q, %q, %v)", data, contentType, err)
	}
	if _, _, err := store.Open(context.Background(), "../secret"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("path traversal error = %v, want os.ErrNotExist", err)
	}
	if err := store.Delete(context.Background(), name); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, _, err := store.Open(context.Background(), name); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Open() after delete error = %v", err)
	}
}
