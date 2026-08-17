package domain

import "testing"

func TestContentTypePath(t *testing.T) {
	tests := map[ContentType]string{
		ContentTypeArticle:     "/articles/example",
		ContentTypeGuide:       "/guides/example",
		ContentTypeBuyingGuide: "/guides/example",
		ContentTypeComparison:  "/compare/example",
	}
	for contentType, expected := range tests {
		if actual := contentType.Path("example"); actual != expected {
			t.Fatalf("%s path = %q, want %q", contentType, actual, expected)
		}
	}
}

func TestBlockValidationRejectsArbitraryContentShapes(t *testing.T) {
	if err := (Block{Type: BlockParagraph, Text: "A useful paragraph."}).Validate(); err != nil {
		t.Fatalf("valid paragraph rejected: %v", err)
	}
	if err := (Block{Type: BlockParagraph, Text: "Text", Items: []string{"unexpected"}}).Validate(); err == nil {
		t.Fatal("paragraph with list items should be rejected")
	}
	if err := (Block{Type: "html", Text: "<script>alert(1)</script>"}).Validate(); err == nil {
		t.Fatal("unsupported HTML block should be rejected")
	}
}
