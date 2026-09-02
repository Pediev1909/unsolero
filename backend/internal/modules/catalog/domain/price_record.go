package domain

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// MaximumPriceRecordEntries caps how far back a published price record goes.
// Ten distinct figures is more history than any competitor publishes and more
// than a reader will read; the rest stays in the revision table.
const MaximumPriceRecordEntries = 10

// maximumPriceNoteLength is the longest note the record will carry. A note is
// there to say why a figure moved in one line. Anything longer is a review
// memo, and printing half of it would misquote it, so it is left out.
const maximumPriceNoteLength = 200

// PriceRevision is one dated price claim, read straight from an immutable fact
// revision. Billing is nil when that revision restated no basis: revisions
// published before the billing columns existed carry none, and a later
// revision may leave the basis untouched. Nil means "not stated here", never
// "no basis".
type PriceRevision struct {
	ObservedAt time.Time
	Price      Money
	Billing    *Billing
	// Note is the reviewer's note on the revision, unedited. PriceRecord
	// decides how much of it is fit to publish.
	Note string
}

// PriceObservation is one row of the public price record: what the price was,
// on what basis, and since when.
type PriceObservation struct {
	ObservedAt time.Time
	Price      Money
	Billing    *Billing
	Note       string
	IsCurrent  bool
}

// PriceRecord turns a product's fact revisions into the dated price history
// the product page publishes. revisions must arrive newest first.
//
// Consecutive revisions that repeat the same claim are collapsed, because a
// revision is written whenever any fact changes and the billing-basis audit of
// 2026-09-02 wrote one for all fifty-three products at once. A row repeating
// the number above it is noise; only a figure that moved is history. Of the
// audit's fifty-three revisions, seven products have a record after this.
//
// A collapsed run keeps the *earliest* date it was read, so a row answers
// "what has this cost, and since when" rather than "when did we last confirm
// it" — the second question is already answered beside the price. It keeps the
// most recently stated basis: an older revision that stated none makes no
// competing claim about a figure that never changed.
//
// Nothing here is inferred. A revision with no stated basis and no later one
// to inherit from publishes a nil basis rather than the basis that looks
// likely, and a note is published only where a reviewer wrote one.
func PriceRecord(revisions []PriceRevision) []PriceObservation {
	record := make([]PriceObservation, 0, len(revisions))
	for _, revision := range revisions {
		if len(record) > 0 && repeatsPriceClaim(record[len(record)-1], revision) {
			previous := &record[len(record)-1]
			previous.ObservedAt = revision.ObservedAt
			previous.Note = priceRecordNote(revision.Note)
			if previous.Billing == nil {
				// The newer revision left the basis untouched, so the basis in
				// force is the one this older revision stated.
				previous.Billing = revision.Billing
			}
			continue
		}
		if len(record) == MaximumPriceRecordEntries {
			break
		}
		record = append(record, PriceObservation{
			ObservedAt: revision.ObservedAt,
			Price:      revision.Price,
			Billing:    revision.Billing,
			Note:       priceRecordNote(revision.Note),
		})
	}
	if len(record) > 0 {
		record[0].IsCurrent = true
	}
	return record
}

// repeatsPriceClaim reports whether an older revision says the same thing
// about the price as the entry already kept. A basis the older revision did
// not state is not a different basis.
func repeatsPriceClaim(kept PriceObservation, older PriceRevision) bool {
	if kept.Price != older.Price {
		return false
	}
	return older.Billing == nil || sameBilling(kept.Billing, older.Billing)
}

func sameBilling(left, right *Billing) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Period == right.Period && left.Unit == right.Unit &&
		sameOptionalString(left.UnitNote, right.UnitNote) &&
		sameOptionalInt64(left.AnnualPriceMinor, right.AnnualPriceMinor)
}

func sameOptionalString(left, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func sameOptionalInt64(left, right *int64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

// auditNotePrefix opens every note the billing-basis audit wrote. The audit's
// own date is already the row's date, so the prefix is stripped rather than
// printed fifty-three times.
const auditNotePrefix = "Billing basis audit "

// priceRecordNote reduces a reviewer's note to the one sentence a reader of the
// record needs, or to nothing.
//
// The notes it is given were written for two audiences. "Monthly-billing list
// price corrected to $20 from Zoho official US pricing observed 2026-08-26" is
// for a reader. "period=monthly, unit=flat, note=per account" is the audit
// restating, for the record, the basis this row already prints in words. The
// second kind is dropped: a note that repeats the row is worse than no note,
// and inventing a friendlier one is not an option.
func priceRecordNote(raw string) string {
	note := strings.TrimSpace(raw)
	if rest, found := strings.CutPrefix(note, auditNotePrefix); found {
		const dated = len("2026-09-02: ")
		if len(rest) > dated && rest[len("2026-09-02"):dated] == ": " {
			note = strings.TrimSpace(rest[dated:])
		}
	}
	note = firstSentence(note)
	if note == "" || utf8.RuneCountInString(note) > maximumPriceNoteLength {
		return ""
	}
	if strings.Contains(note, "period=") || strings.Contains(note, "unit=") {
		return ""
	}
	return upperFirstRune(note)
}

// firstSentence cuts at the first full stop that ends a word rather than one
// inside a figure: "2.9% + 30¢ per transaction. Unchanged." keeps its price.
func firstSentence(note string) string {
	for index, character := range note {
		if character != '.' {
			continue
		}
		next := index + 1
		if next == len(note) {
			return note
		}
		if note[next] == ' ' || note[next] == '\n' {
			return strings.TrimSpace(note[:next])
		}
	}
	return note
}

func upperFirstRune(value string) string {
	first, size := utf8.DecodeRuneInString(value)
	if first == utf8.RuneError {
		return value
	}
	return string(unicode.ToUpper(first)) + value[size:]
}
