package domain

import (
	"errors"
	"net/url"
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
)

type BlockType string

const (
	BlockParagraph BlockType = "paragraph"
	BlockHeading   BlockType = "heading"
	BlockUnordered BlockType = "unordered_list"
	BlockOrdered   BlockType = "ordered_list"
	BlockQuote     BlockType = "quote"
	BlockCallout   BlockType = "callout"
)

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
	case ContentTypeArticle, ContentTypeGuide, ContentTypeBuyingGuide, ContentTypeComparison:
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
	default:
		return ErrInvalidContent
	}
	return nil
}
