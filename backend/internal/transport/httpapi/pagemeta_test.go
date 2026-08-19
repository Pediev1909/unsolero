package httpapi

import (
	"strings"
	"testing"
	"unicode/utf8"
)

const testShell = `<!doctype html>
<html>
  <head>
    <!--PAGE_META_START-->
    <title>Default</title>
    <meta name="description" content="Default description" />
    <!--PAGE_META_END-->
    <script src="/assets/index-abc123.js"></script>
  </head>
  <body><div id="root"></div></body>
</html>`

func TestRenderShellReplacesOnlyTheMetadataBlock(t *testing.T) {
	rendered, ok := renderShell(testShell, pageMetadata{
		Title:        "ClickUp Unlimited — ClickUp | UNSOLERO",
		Description:  "Project and task tracking.",
		CanonicalURL: "https://unsolero.com/products/clickup-unlimited",
		Indexable:    true,
	})
	if !ok {
		t.Fatal("renderShell reported failure on a shell containing both markers")
	}

	if !strings.Contains(rendered, "<title>ClickUp Unlimited — ClickUp | UNSOLERO</title>") {
		t.Fatalf("rendered shell is missing the route title:\n%s", rendered)
	}
	if strings.Contains(rendered, "Default description") {
		t.Fatal("the shell's default metadata survived the replacement")
	}
	// The asset tags outside the markers carry the hashed bundle names; losing
	// them would serve a blank page.
	if !strings.Contains(rendered, "/assets/index-abc123.js") {
		t.Fatal("renderShell dropped content outside the metadata markers")
	}
	if !strings.Contains(rendered, `<link rel="canonical" href="https://unsolero.com/products/clickup-unlimited" />`) {
		t.Fatal("rendered shell is missing the canonical link")
	}
}

// Product names and descriptions are catalog data. A quote in one of them must
// not be able to close an attribute and add markup of its own.
func TestRenderShellEscapesMetadataValues(t *testing.T) {
	rendered, ok := renderShell(testShell, pageMetadata{
		Title:       `Tool" onload="alert(1)`,
		Description: `<img src=x onerror=alert(1)>`,
		Indexable:   true,
	})
	if !ok {
		t.Fatal("renderShell reported failure")
	}
	if strings.Contains(rendered, `onload="alert(1)`) {
		t.Fatalf("a quote in the title escaped its attribute:\n%s", rendered)
	}
	if strings.Contains(rendered, "<img src=x") {
		t.Fatalf("markup in the description was not escaped:\n%s", rendered)
	}
}

// JSON-LD is written inside a script element, where the only escape needed is
// the one that stops a literal </script> from closing the block early.
func TestRenderShellEscapesStructuredDataClosingTag(t *testing.T) {
	rendered, ok := renderShell(testShell, pageMetadata{
		Title:     "Product",
		Indexable: true,
		StructuredData: map[string]any{
			"@type": "Product",
			"name":  `Bad</script><script>alert(1)</script>`,
		},
	})
	if !ok {
		t.Fatal("renderShell reported failure")
	}
	if strings.Contains(rendered, "</script><script>alert(1)") {
		t.Fatalf("structured data closed its own script element:\n%s", rendered)
	}
	if !strings.Contains(rendered, `</script`) {
		t.Fatal("expected the closing tag to be escaped as a unicode sequence")
	}
}

func TestRenderShellFailsWhenMarkersAreAbsent(t *testing.T) {
	if _, ok := renderShell("<html><head></head></html>", pageMetadata{Title: "x"}); ok {
		t.Fatal("renderShell claimed success on a shell with no markers")
	}
}

func TestNonIndexableRoutesCarryARobotsDirective(t *testing.T) {
	rendered, _ := renderShell(testShell, pageMetadata{Title: "Account", Indexable: false})
	if !strings.Contains(rendered, `<meta name="robots" content="noindex, nofollow" />`) {
		t.Fatal("a non-indexable route was rendered without a robots directive")
	}

	indexed, _ := renderShell(testShell, pageMetadata{Title: "Product", Indexable: true})
	if strings.Contains(indexed, `name="robots"`) {
		t.Fatal("an indexable route was rendered with a robots directive")
	}
}

func TestTruncateDescriptionCutsOnAWordBoundary(t *testing.T) {
	long := strings.Repeat("word ", 60)
	result := truncateDescription(long, 60)

	if count := len([]rune(result)); count > 61 {
		t.Fatalf("truncated description is %d characters, want at most 61", count)
	}
	if !strings.HasSuffix(result, "…") {
		t.Fatalf("expected an ellipsis on a truncated value, got %q", result)
	}
	if strings.Contains(result, "  ") {
		t.Fatalf("collapsed whitespace expected, got %q", result)
	}

	short := truncateDescription("  Already   short.  ", 160)
	if short != "Already short." {
		t.Fatalf("short description = %q, want %q", short, "Already short.")
	}

	// Byte-slicing would split these characters and emit invalid UTF-8.
	multibyte := truncateDescription(strings.Repeat("оценка ", 40), 20)
	if !utf8.ValidString(multibyte) {
		t.Fatalf("truncation produced invalid UTF-8: %q", multibyte)
	}
	if count := len([]rune(multibyte)); count > 21 {
		t.Fatalf("multibyte truncation is %d characters, want at most 21", count)
	}
}

// Serving the shell from the API moved the document out from behind nginx,
// which used to attach the document policy. The API's own policy is
// "default-src 'none'", which blocks every script and stylesheet the page
// references and renders a blank page. This asserts the document policy admits
// what an HTML page actually needs.
func TestDocumentPolicyAllowsTheAssetsThePageLoads(t *testing.T) {
	for _, directive := range []string{
		"default-src 'self'",
		"script-src 'self'",
		"style-src 'self'",
		"connect-src 'self'",
		"frame-ancestors 'none'",
	} {
		if !strings.Contains(documentContentSecurityPolicy, directive) {
			t.Fatalf("document policy is missing %q:\n%s", directive, documentContentSecurityPolicy)
		}
	}
	if strings.Contains(documentContentSecurityPolicy, "default-src 'none'") {
		t.Fatal("the document policy is the API's JSON policy, which blocks all page assets")
	}
}
