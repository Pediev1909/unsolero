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

// The CTA block is the only block that can put a paid destination inside an
// article, so its validation is the boundary that decides what an editor can
// publish. Every case below is a way that boundary could be walked through.
func TestCTABlockOnlyNamesAnApprovedPromotion(t *testing.T) {
	valid := Block{
		Type:      BlockCTA,
		Heading:   "If automation is why you are leaving",
		Text:      "Their own comparison is the honest place to start.",
		Label:     "See ActiveCampaign against Mailchimp",
		Promotion: "activecampaign-mailchimp-switch",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid CTA rejected: %v", err)
	}

	// A heading is optional; the explanation is not. A button a reader cannot
	// interpret is the whole failure mode this block exists to avoid.
	noHeading := valid
	noHeading.Heading = ""
	if err := noHeading.Validate(); err != nil {
		t.Fatalf("CTA without a heading rejected: %v", err)
	}
	noText := valid
	noText.Text = ""
	if err := noText.Validate(); err == nil {
		t.Fatal("CTA without its explanation should be rejected")
	}

	// A promotion is a slug. Anything that could steer the redirect handler
	// somewhere other than one row of commerce.affiliate_promotions is not one.
	for name, promotion := range map[string]string{
		"absent":         "",
		"absolute URL":   "https://evil.example/pay-me",
		"path traversal": "../../admin",
		"path separator": "promo/other",
		"query string":   "promo?source=promotion",
		"uppercase":      "ActiveCampaign",
		"leading dash":   "-activecampaign",
		"underscore":     "active_campaign",
	} {
		block := valid
		block.Promotion = promotion
		if err := block.Validate(); err == nil {
			t.Fatalf("CTA with a %s promotion (%q) should be rejected", name, promotion)
		}
	}

	noLabel := valid
	noLabel.Label = ""
	if err := noLabel.Validate(); err == nil {
		t.Fatal("CTA without a label should be rejected")
	}

	// The CTA fields are meaningless on any other block, and a paragraph that
	// carried a promotion slug would be a link the renderer never draws and
	// nobody reviews.
	if err := (Block{Type: BlockParagraph, Text: "Text", Promotion: "some-promo"}).Validate(); err == nil {
		t.Fatal("paragraph carrying a promotion should be rejected")
	}
	if err := (Block{Type: BlockParagraph, Text: "Text", Label: "Buy"}).Validate(); err == nil {
		t.Fatal("paragraph carrying a CTA label should be rejected")
	}
}
