package httpapi

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	catalog "rigmark/internal/modules/catalog/domain"
	content "rigmark/internal/modules/content/domain"
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

// An article whose text never reaches the document reads as an empty page to
// every client that does not run the application, including the assistants the
// site publishes an llms.txt for.
func TestEntryBodyReachesTheDocument(t *testing.T) {
	entry := content.Entry{Content: []content.Block{
		{Type: content.BlockHeading, Heading: "How the score is built"},
		{Type: content.BlockParagraph, Text: "Commission is not an input."},
		{Type: content.BlockUnordered, Items: []string{"Recorded source", "Date read"}},
	}}
	entry.Title = "How UNSOLERO ranks software"
	entry.Author = content.Author{Name: "A. Reviewer"}
	entry.PublishedAt = time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC)

	rendered, ok := renderShell(testShell, pageMetadata{
		Title: "t", Indexable: true, PrerenderedBody: renderEntryBody(entry),
	})
	if !ok {
		t.Fatal("renderShell reported failure")
	}
	for _, want := range []string{
		"How the score is built",
		"Commission is not an input.",
		"<li>Recorded source</li>",
		"A. Reviewer",
		`<time datetime="2026-08-19T00:00:00Z">`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("document is missing %q:\n%s", want, rendered)
		}
	}
	// The mount point must still be the single container React takes over.
	if strings.Count(rendered, `id="root"`) != 1 {
		t.Fatalf("expected exactly one root element:\n%s", rendered)
	}
}

// Block text is stored data. It must not be able to introduce markup.
func TestEntryBodyEscapesBlockText(t *testing.T) {
	entry := content.Entry{Content: []content.Block{
		{Type: content.BlockParagraph, Text: `</div><script>alert(1)</script>`},
		{Type: content.BlockUnordered, Items: []string{`<img src=x onerror=alert(1)>`}},
	}}
	entry.Title = `Title <script>alert(1)</script>`

	body := renderEntryBody(entry)
	if strings.Contains(body, "<script>") || strings.Contains(body, "<img") {
		t.Fatalf("stored text produced live markup:\n%s", body)
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Fatalf("expected escaped text in the body:\n%s", body)
	}
}

// A route with no body must leave the shell exactly as it was.
func TestShellIsUnchangedWithoutAPrerenderedBody(t *testing.T) {
	rendered, ok := renderShell(testShell, pageMetadata{Title: "t", Indexable: true})
	if !ok {
		t.Fatal("renderShell reported failure")
	}
	if !strings.Contains(rendered, `<div id="root"></div>`) {
		t.Fatalf("empty mount point was modified:\n%s", rendered)
	}
}

// The live audit reported 125 pages with no <h1> and 126 with no inbound
// internal link. Both are the same defect: a route whose body only exists after
// JavaScript runs is, to anything that does not run it, a page with no heading
// and no anchors. These pin the server-rendered replacement.
func TestRenderProductBodyCarriesHeadingAndCatalogLinks(t *testing.T) {
	body := renderProductBody(catalog.Product{
		Name: "ClickUp Unlimited", Slug: "clickup-unlimited",
		BrandName: "ClickUp", BrandSlug: "clickup",
		CategoryName: "Project management", CategorySlug: "project-management",
		Description: "Project and task tracking on the entry paid tier.",
		Price:       catalog.Money{AmountMinor: 1000, Currency: "USD"},
		Billing:     catalog.Billing{Period: catalog.BillingMonthly, Unit: catalog.PricingUnitPerUser},
	})
	for _, want := range []string{
		"<h1", "ClickUp Unlimited",
		`href="/brands/clickup"`,
		`href="/categories/project-management"`,
		"USD 10.00 <span",
		">per user, monthly billing</span>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("product body is missing %q\ngot: %s", want, body)
		}
	}
	if strings.Contains(body, "per month, entry paid tier") {
		t.Errorf("product body still asserts a basis it does not know\ngot: %s", body)
	}
}

// The basis under the price is the record's, not a constant. Half the software
// catalog stored annual-contract figures under a line that said "per month".
func TestRenderProductBodyStatesTheRecordedBillingBasis(t *testing.T) {
	note := "at 1,000 contacts"
	cases := []struct {
		name    string
		product catalog.Product
		want    string
	}{
		{"annual only", catalog.Product{Name: "Kit Creator", Slug: "kit-creator",
			Price:   catalog.Money{AmountMinor: 2_900, Currency: "USD"},
			Billing: catalog.Billing{Period: catalog.BillingAnnual, Unit: catalog.PricingUnitFlat}},
			"USD 29.00 <span class=\"text-ink/70\">flat rate, billed yearly</span>"},
		{"contact tier", catalog.Product{Name: "Mailchimp Standard", Slug: "mailchimp-standard",
			Price:   catalog.Money{AmountMinor: 2_000, Currency: "USD"},
			Billing: catalog.Billing{Period: catalog.BillingMonthly, Unit: catalog.PricingUnitPerContacts, UnitNote: &note}},
			"USD 20.00 <span class=\"text-ink/70\">at 1,000 contacts, monthly billing</span>"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if body := renderProductBody(testCase.product); !strings.Contains(body, testCase.want) {
				t.Errorf("product body is missing %q\ngot: %s", testCase.want, body)
			}
		})
	}
	// A physical product has no billing cadence, and the default basis a
	// physical row carries must not be printed as if it were one.
	physical := renderProductBody(catalog.Product{Name: "Kettlebell", Slug: "kettlebell", IsPhysical: true,
		Price:   catalog.Money{AmountMinor: 4_900, Currency: "USD"},
		Billing: catalog.Billing{Period: catalog.BillingMonthly, Unit: catalog.PricingUnitFlat}})
	if !strings.Contains(physical, "USD 49.00</p>") || strings.Contains(physical, "billing") {
		t.Errorf("physical product body carries a billing phrase\ngot: %s", physical)
	}
}

func TestRenderCatalogListingBodyLinksEveryProduct(t *testing.T) {
	body := renderCatalogListingBody("CRM", "Customer records and pipelines.",
		[]catalog.Product{
			{Name: "Zoho CRM Standard", Slug: "zoho-crm-standard", BrandName: "Zoho"},
			{Name: "Pipedrive Lite", Slug: "pipedrive-lite", BrandName: "Pipedrive"},
		})
	if !strings.Contains(body, "<h1") || !strings.Contains(body, "CRM") {
		t.Errorf("listing body has no heading\ngot: %s", body)
	}
	for _, slug := range []string{"zoho-crm-standard", "pipedrive-lite"} {
		if !strings.Contains(body, `href="/products/`+slug+`"`) {
			t.Errorf("listing body does not link %q\ngot: %s", slug, body)
		}
	}
}

// A name carrying markup must not become markup. The renderers build HTML by
// concatenation, so escaping is the only thing standing between a catalog value
// and an injected element.
func TestRenderedBodiesEscapeCatalogValues(t *testing.T) {
	body := renderProductBody(catalog.Product{
		Name: `<script>alert(1)</script>`, Slug: "x", BrandName: `"onload="x`, BrandSlug: "b",
	})
	if strings.Contains(body, "<script>") {
		t.Errorf("product name was not escaped\ngot: %s", body)
	}
	listing := renderCatalogListingBody(`<img src=x onerror=y>`, "", nil)
	if strings.Contains(listing, "<img") {
		t.Errorf("listing heading was not escaped\ngot: %s", listing)
	}
}

// The prerendered body is what a crawler and a JavaScript-less reader get, so
// the CTA has to be a working, disclosed link there and not only in React.
func TestEntryBodyRendersCTAAsADisclosedTrackedLink(t *testing.T) {
	entry := content.Entry{Content: []content.Block{
		{
			Type:      content.BlockCTA,
			Heading:   "If automation is why you are leaving",
			Text:      "Their own comparison is the honest place to start.",
			Label:     "See ActiveCampaign against Mailchimp",
			Promotion: "activecampaign-mailchimp-switch",
		},
	}}
	entry.Title = "Mailchimp alternatives"

	body := renderEntryBody(entry)

	// source=promotion is load-bearing. TrackPromotionClick rejects any other
	// source and the handler defaults an absent one to product_detail, so
	// without it every no-JS click resolves to an error instead of a vendor.
	if !strings.Contains(body,
		"/api/affiliate/promotion/activecampaign-mailchimp-switch?source=promotion") {
		t.Fatalf("CTA did not render a tracked promotion path:\n%s", body)
	}
	// An undisclosed paid link in indexed HTML is what search engines penalise.
	for _, want := range []string{`rel="nofollow noopener sponsored"`,
		"See ActiveCampaign against Mailchimp",
		"Their own comparison is the honest place to start."} {
		if !strings.Contains(body, want) {
			t.Fatalf("CTA body is missing %q:\n%s", want, body)
		}
	}
}

// A CTA missing either half is a button with no destination or a destination
// with no button. Neither should reach the document.
func TestEntryBodySkipsIncompleteCTA(t *testing.T) {
	for name, block := range map[string]content.Block{
		"no promotion": {Type: content.BlockCTA, Text: "t", Label: "Go"},
		"no label":     {Type: content.BlockCTA, Text: "t", Promotion: "a-promo"},
	} {
		body := renderEntryBody(content.Entry{Content: []content.Block{block}})
		if strings.Contains(body, "/api/affiliate/promotion/") {
			t.Fatalf("%s produced a link:\n%s", name, body)
		}
	}
}

// The three editorial blocks added for the comparison pages: what a tool does
// well and what it costs, the questions readers actually ask, and a vendor exit
// by product slug. Each has to reach the document, and the offer must send a
// crawler to the product page rather than the affiliate redirect.
func TestEntryBodyRendersProsConsFAQAndOfferBlocks(t *testing.T) {
	entry := content.Entry{Content: []content.Block{
		{Type: content.BlockProsCons, Heading: "Verdict", Pros: []string{"Generous free tier"}, Cons: []string{"Automation starts at the paid tier"}},
		{Type: content.BlockFAQ, Heading: "Questions", Questions: []content.QuestionAnswer{
			{Question: "Is there a free plan?", Answer: "Yes, up to 500 contacts."},
		}},
		{Type: content.BlockOffer, Product: "kit-creator", Heading: "Try Kit", Text: "Fourteen days, no card.", Label: "See Kit"},
	}}
	entry.Title = "Mailchimp against Kit"

	body := renderEntryBody(entry)
	for _, want := range []string{
		"<h3", ">Pros</h3>", "<li>Generous free tier</li>",
		">Cons</h3>", "<li>Automation starts at the paid tier</li>",
		">Is there a free plan?</h3>", "Yes, up to 500 contacts.",
		`href="/products/kit-creator"`, "Fourteen days, no card.", ">See Kit</a>",
		"Affiliate link on the product page. Commission never changes the ranking.",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("entry body is missing %q\ngot: %s", want, body)
		}
	}
	if strings.Contains(body, "/api/affiliate/") {
		t.Errorf("an offer block put the affiliate redirect into the static body:\n%s", body)
	}
}

// Every field of the new blocks is stored text. A "<" in any of them must
// arrive as an entity, not an element.
func TestNewBlocksEscapeStoredText(t *testing.T) {
	entry := content.Entry{Content: []content.Block{
		{Type: content.BlockProsCons, Pros: []string{"<b>bold</b> pro"}, Cons: []string{"<i>con</i>"}},
		{Type: content.BlockFAQ, Questions: []content.QuestionAnswer{{Question: "<q>?", Answer: "<script>alert(1)</script>"}}},
		{Type: content.BlockOffer, Product: "kit-creator", Text: "<img src=x onerror=y>", Label: "<em>go</em>"},
	}}
	body := renderEntryBody(entry)
	for _, live := range []string{"<b>", "<i>", "<q>", "<script>", "<img", "<em>"} {
		if strings.Contains(body, live) {
			t.Errorf("stored text produced live markup %q:\n%s", live, body)
		}
	}
	if strings.Count(body, "&lt;") < 6 {
		t.Errorf("expected every < to be escaped:\n%s", body)
	}
}

func TestFAQStructuredDataExistsOnlyForFAQBlocks(t *testing.T) {
	if data := faqStructuredData([]content.Block{{Type: content.BlockParagraph, Text: "No questions here."}}); data != nil {
		t.Fatalf("FAQPage emitted without a FAQ block: %v", data)
	}
	data := faqStructuredData([]content.Block{
		{Type: content.BlockFAQ, Questions: []content.QuestionAnswer{{Question: "Q1?", Answer: "A1"}}},
		{Type: content.BlockFAQ, Questions: []content.QuestionAnswer{{Question: "Q2?", Answer: "A2"}}},
	})
	if data["@type"] != "FAQPage" {
		t.Fatalf("@type = %v, want FAQPage", data["@type"])
	}
	questions, _ := data["mainEntity"].([]map[string]any)
	if len(questions) != 2 || questions[1]["name"] != "Q2?" {
		t.Fatalf("mainEntity = %v, want both questions in order", data["mainEntity"])
	}
}

// A page gets one JSON-LD block. Several nodes share a single @context under
// @graph, and an absent optional node simply does not appear.
func TestStructuredDataGraphJoinsNodesUnderOneContext(t *testing.T) {
	product := map[string]any{"@context": "https://schema.org", "@type": "Product"}
	graph := structuredDataGraph(product, nil, map[string]any{"@type": "BreadcrumbList"})
	nodes, _ := graph["@graph"].([]map[string]any)
	if len(nodes) != 2 || graph["@context"] != "https://schema.org" {
		t.Fatalf("graph = %v", graph)
	}
	if _, nested := nodes[0]["@context"]; nested {
		t.Fatal("a node kept its own @context inside the graph")
	}
}
