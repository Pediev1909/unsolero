package localimages

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
)

const testProductID catalog.ProductID = "12345678-1234-4234-8234-123456789abc"

var testPNG = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

func TestStoreLifecycleAndPathValidation(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	name, created, err := store.Save(context.Background(), testProductID, testPNG, ".png")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if !created || !strings.HasSuffix(name, ".png") || !strings.HasPrefix(name, string(testProductID)+"_") {
		t.Fatalf("name = %q", name)
	}
	data, contentType, err := store.Open(context.Background(), name)
	if err != nil || !bytes.Equal(data, testPNG) || contentType != "image/png" {
		t.Fatalf("Open() = (%q, %q, %v)", data, contentType, err)
	}
	duplicate, duplicateCreated, err := store.Save(context.Background(), testProductID, testPNG, ".png")
	if err != nil || duplicate != name || duplicateCreated {
		t.Fatalf("duplicate Save() = (%q, %t, %v)", duplicate, duplicateCreated, err)
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

func TestStoreFailsClosedWhenStorageBecomesUnavailable(t *testing.T) {
	directory := t.TempDir()
	store, err := New(directory)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(directory); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(directory, []byte("storage unavailable fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Save(context.Background(), testProductID, testPNG, ".png"); err == nil {
		t.Fatal("Save() succeeded after storage directory became unavailable")
	}
	if _, _, err := store.Open(context.Background(), "00000000000000000000000000000000.webp"); err == nil {
		t.Fatal("Open() succeeded after storage directory became unavailable")
	}
}

func TestStoreRejectsMismatchedExecutableAndOversizedContent(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fixtures := []struct {
		data      []byte
		extension string
	}{
		{[]byte("<script>alert(1)</script>"), ".png"},
		{testPNG, ".svg"},
		{make([]byte, maximumImageBytes+1), ".png"},
	}
	for _, fixture := range fixtures {
		if _, _, err := store.Save(context.Background(), testProductID, fixture.data, fixture.extension); err == nil {
			t.Fatalf("Save() accepted extension=%q size=%d", fixture.extension, len(fixture.data))
		}
	}
}
