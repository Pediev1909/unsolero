package mediaobject

import (
	"strings"
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
)

func TestDeterministicObjectDescriptionAndOwnership(t *testing.T) {
	productID := catalog.ProductID("12345678-1234-4234-8234-123456789abc")
	data := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	first, err := Describe(productID, data, ".png")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Describe(productID, data, ".png")
	if err != nil || first != second {
		t.Fatalf("descriptions differ: %#v %#v %v", first, second, err)
	}
	if !strings.HasPrefix(first.ObjectKey, "products/"+string(productID)+"/") || !BelongsTo(productID, first.Name) {
		t.Fatalf("description = %#v", first)
	}
	if _, err := Parse("../" + first.Name); err == nil {
		t.Fatal("path traversal name accepted")
	}
}

func TestParseObjectKeyRejectsUnexpectedNamespaces(t *testing.T) {
	description, err := Describe("12345678-1234-4234-8234-123456789abc", []byte("\x89PNG\r\n\x1a\n12345678"), ".png")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseObjectKey(description.ObjectKey)
	if err != nil || parsed.Name != description.Name {
		t.Fatalf("ParseObjectKey() = (%+v, %v)", parsed, err)
	}
	for _, key := range []string{"incoming/file.png", "products/other/file.png", description.ObjectKey + "/extra", "products\\escape"} {
		if _, err := ParseObjectKey(key); err == nil {
			t.Fatalf("ParseObjectKey(%q) succeeded", key)
		}
	}
}
