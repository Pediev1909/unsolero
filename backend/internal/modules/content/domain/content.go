package domain

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
)

type ContentType string

const (
	ContentTypeArticle     ContentType = "article"
	ContentTypeGuide       ContentType = "guide"
	ContentTypeBuyingGuide ContentType = "buying_guide"
	ContentTypeComparison  ContentType = "comparison"
	// ContentTypeStack is a whole set of tools chosen for one kind of business
	// and one budget: what to run, what was deliberately left out and why, and
	// the total per month. It is the shape the recommendation builder produces,
	// published as an indexable page where /build itself is noindex.
	ContentTypeStack ContentType = "stack"
)

type BlockType string

const (
	BlockParagraph BlockType = "paragraph"
	BlockHeading   BlockType = "heading"
	BlockUnordered BlockType = "unordered_list"
	BlockOrdered   BlockType = "ordered_list"
	BlockQuote     BlockType = "quote"
	BlockCallout   BlockType = "callout"
	// BlockCTA carries an affiliate destination inside an article.
	//
	// It names a promotion by slug and never a URL. An editor who could type a
	// href into a content block would be able to publish an untracked link, a
	// link to anywhere, or somebody else's affiliate code, and the body is
	// rendered into the served HTML where that would be invisible until it had
	// already been crawled. A slug can only resolve to a row in
	// commerce.affiliate_promotions, which already enforces https, an active
	// merchant, a freshness window and a disclosure label — so the block
	// chooses which approved destination to show and nothing else.
	BlockCTA BlockType = "cta"
	// BlockProsCons is a two-column list of what a tool does well and what it
	// costs you. Both sides are required: a list of strengths with no
	// trade-offs is an advertisement, and this site does not publish those.
	BlockProsCons BlockType = "pros_cons"
	// BlockFAQ carries question-and-answer pairs. The questions are meant to
	// be the ones readers actually ask — search suggestions, Reddit threads —
	// and the server emits them as schema.org/FAQPage.
	BlockFAQ BlockType = "faq"
	// BlockOffer places a vendor exit for a catalog product inside an article,
	// by product slug. The destination is resolved when the page renders, from
	// the product's live offer, through the same freshness and liveness checks
	// the product page applies. So the block can only point at what the
	// catalog already serves, and an editor still cannot type a URL.
	BlockOffer BlockType = "offer"
)

// A promotion slug, matching the column constraint in
// commerce.affiliate_promotions. Anchored, so a block cannot smuggle a path
// separator or a scheme past the redirect handler.
var promotionSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Author struct {
	ID        string
	Name      string
	Slug      string
	Bio       string
	AvatarURL *string
}

type Block struct {
	Type        BlockType `json:"type"`
	Heading     string    `json:"heading,omitempty"`
	Text        string    `json:"text,omitempty"`
	Items       []string  `json:"items,omitempty"`
	Attribution string    `json:"attribution,omitempty"`
	// Promotion and Label belong to BlockCTA. Promotion is the slug of an
	// affiliate promotion; Label is the text on the control.
	Promotion string `json:"promotion,omitempty"`
	Label     string `json:"label,omitempty"`
	// Pros and Cons belong to BlockProsCons.
	Pros []string `json:"pros,omitempty"`
	Cons []string `json:"cons,omitempty"`
	// Questions belongs to BlockFAQ.
	Questions []QuestionAnswer `json:"questions,omitempty"`
	// Product belongs to BlockOffer: the slug of a catalog product. Label is
	// shared with BlockCTA and is optional here; the control defaults to the
	// merchant's name when it is empty.
	Product string `json:"product,omitempty"`
}

// QuestionAnswer is one entry of a BlockFAQ.
type QuestionAnswer struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type CategoryReference struct {
	ID          catalog.CategoryID
	Name        string
	Slug        string
	Description string
}

type Summary struct {
	ID           string
	Type         ContentType
	Title        string
	Slug         string
	Description  string
	HeroImageURL string
	HeroImageAlt string
	AuthorName   string
	PublishedAt  time.Time
	UpdatedAt    time.Time
	Path         string
	// Covered names a card after the products the piece actually compares.
	// Thirteen comparisons shared one illustration and six guides shared
	// another, so a grid of them read as a template rather than a library. The
	// products differ on every piece, which makes the card differ for free.
	Covered []CoveredProduct
}

type CoveredProduct struct {
	Name       string
	PriceMinor int64
	Currency   string
}

type Entry struct {
	Summary
	Author            Author
	Content           []Block
	ProductIDs        []catalog.ProductID
	RelatedProducts   []catalog.Product
	RelatedCategories []CategoryReference
	RelatedEntries    []Summary
	SEOTitle          string
	SEODescription    string
	CanonicalURL      string
}

type SitemapEntry struct {
	Path       string
	ModifiedAt time.Time
}

var ErrInvalidContent = errors.New("invalid editorial content")

func (contentType ContentType) Valid() bool {
	switch contentType {
	case ContentTypeArticle, ContentTypeGuide, ContentTypeBuyingGuide, ContentTypeComparison, ContentTypeStack:
		return true
	default:
		return false
	}
}

func (contentType ContentType) Path(slug string) string {
	switch contentType {
	case ContentTypeArticle:
		return "/articles/" + slug
	case ContentTypeGuide, ContentTypeBuyingGuide:
		return "/guides/" + slug
	case ContentTypeComparison:
		return "/compare/" + slug
	case ContentTypeStack:
		return "/stacks/" + slug
	default:
		return ""
	}
}

func (entry Entry) Validate() error {
	if !entry.Type.Valid() || entry.Path != entry.Type.Path(entry.Slug) ||
		len(strings.TrimSpace(entry.Title)) < 10 || len(entry.Content) == 0 ||
		len(entry.SEOTitle) < 10 || len(entry.SEODescription) < 40 ||
		entry.PublishedAt.IsZero() || entry.UpdatedAt.IsZero() {
		return ErrInvalidContent
	}
	for _, block := range entry.Content {
		if err := block.Validate(); err != nil {
			return err
		}
	}
	if entry.CanonicalURL != "" {
		parsed, err := url.Parse(entry.CanonicalURL)
		localHTTP := parsed.Scheme == "http" && (parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1")
		if err != nil || parsed.Host == "" || parsed.Scheme != "https" && !localHTTP {
			return ErrInvalidContent
		}
	}
	return nil
}

func (block Block) Validate() error {
	textLength := len(strings.TrimSpace(block.Text))
	switch block.Type {
	case BlockParagraph:
		if textLength < 1 || textLength > 5_000 || block.Heading != "" || len(block.Items) > 0 {
			return ErrInvalidContent
		}
	case BlockHeading:
		if len(strings.TrimSpace(block.Heading)) < 2 || len(block.Heading) > 180 || block.Text != "" || len(block.Items) > 0 {
			return ErrInvalidContent
		}
	case BlockUnordered, BlockOrdered:
		if len(block.Items) < 1 || len(block.Items) > 30 || block.Text != "" {
			return ErrInvalidContent
		}
		for _, item := range block.Items {
			if len(strings.TrimSpace(item)) < 1 || len(item) > 1_000 {
				return ErrInvalidContent
			}
		}
	case BlockQuote:
		if textLength < 1 || textLength > 2_000 || len(block.Attribution) > 160 {
			return ErrInvalidContent
		}
	case BlockCallout:
		if textLength < 1 || textLength > 2_000 || len(strings.TrimSpace(block.Heading)) < 2 || len(block.Heading) > 120 {
			return ErrInvalidContent
		}
	case BlockCTA:
		// Text is required, not optional. A button in the middle of an article
		// that a reader is trusting to be impartial has to say what it is and
		// why it is there; the disclosure is the content, not decoration.
		label := len(strings.TrimSpace(block.Label))
		if !promotionSlugPattern.MatchString(block.Promotion) || len(block.Promotion) > 120 ||
			label < 2 || label > 60 || textLength < 1 || textLength > 600 ||
			len(block.Heading) > 120 || len(block.Items) > 0 || block.Attribution != "" {
			return ErrInvalidContent
		}
	case BlockProsCons:
		if len(block.Pros) < 1 || len(block.Pros) > 8 || len(block.Cons) < 1 || len(block.Cons) > 8 ||
			len(block.Heading) > 120 || textLength > 600 || len(block.Items) > 0 || block.Attribution != "" {
			return ErrInvalidContent
		}
		for _, item := range append(append([]string{}, block.Pros...), block.Cons...) {
			if len(strings.TrimSpace(item)) < 1 || len(item) > 240 {
				return ErrInvalidContent
			}
		}
	case BlockFAQ:
		if len(block.Questions) < 1 || len(block.Questions) > 10 || len(block.Heading) > 120 ||
			textLength > 600 || len(block.Items) > 0 || block.Attribution != "" {
			return ErrInvalidContent
		}
		for _, pair := range block.Questions {
			if len(strings.TrimSpace(pair.Question)) < 5 || len(pair.Question) > 200 ||
				len(strings.TrimSpace(pair.Answer)) < 1 || len(pair.Answer) > 1_200 {
				return ErrInvalidContent
			}
		}
	case BlockOffer:
		// The same anchored slug pattern as a promotion: a product slug that
		// carried a path separator would be a different URL, not a product.
		if !promotionSlugPattern.MatchString(block.Product) || len(block.Product) > 120 ||
			len(strings.TrimSpace(block.Label)) > 60 || textLength > 600 || len(block.Heading) > 120 ||
			len(block.Items) > 0 || block.Attribution != "" {
			return ErrInvalidContent
		}
	default:
		return ErrInvalidContent
	}
	// Fields that belong to one block type are rejected on every other, so a
	// stray field cannot ride along unrendered and surface later.
	if block.Type != BlockCTA && block.Promotion != "" {
		return ErrInvalidContent
	}
	if block.Type != BlockCTA && block.Type != BlockOffer && block.Label != "" {
		return ErrInvalidContent
	}
	if block.Type != BlockProsCons && (len(block.Pros) > 0 || len(block.Cons) > 0) {
		return ErrInvalidContent
	}
	if block.Type != BlockFAQ && len(block.Questions) > 0 {
		return ErrInvalidContent
	}
	if block.Type != BlockOffer && block.Product != "" {
		return ErrInvalidContent
	}
	return nil
}
