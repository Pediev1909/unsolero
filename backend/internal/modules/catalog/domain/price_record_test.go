package domain

import (
	"testing"
	"time"
)

func usd(minor int64) Money { return Money{AmountMinor: minor, Currency: "USD"} }

func day(day int) time.Time {
	return time.Date(2026, 9, day, 9, 28, 0, 0, time.UTC)
}

func monthlyFlat(annual *int64) *Billing {
	return &Billing{Period: BillingMonthly, Unit: PricingUnitFlat, AnnualPriceMinor: annual}
}

// The billing-basis audit wrote a fact revision for every product, whether its
// price moved or not. A record that lists the same figure twice is not a price
// history, and on forty-six of fifty-three products that is all there would be.
func TestPriceRecordCollapsesARevisionThatRepeatsTheSamePrice(t *testing.T) {
	record := PriceRecord([]PriceRevision{
		{ObservedAt: day(2), Price: usd(2900), Billing: monthlyFlat(nil),
			Note: "Billing basis audit 2026-09-02: monthly only; period=monthly, unit=flat. Compared price unchanged."},
		{ObservedAt: day(1), Price: usd(2900),
			Note: "Entry paid tier at monthly billing, read from the vendor pricing page on 2026-08-18."},
	})

	if len(record) != 1 {
		t.Fatalf("record has %d entries, want the price stated once: %#v", len(record), record)
	}
	// The row answers "since when", so it keeps the earlier date, and the
	// basis the newer revision stated about the same unmoved figure.
	if !record[0].ObservedAt.Equal(day(1)) {
		t.Fatalf("observed at = %v, want the first date the figure was read", record[0].ObservedAt)
	}
	if record[0].Billing == nil || record[0].Billing.Period != BillingMonthly {
		t.Fatalf("billing = %#v, want the stated monthly basis", record[0].Billing)
	}
	if !record[0].IsCurrent {
		t.Fatal("the newest entry is the current price")
	}
	if record[0].Note != "Entry paid tier at monthly billing, read from the vendor pricing page on 2026-08-18." {
		t.Fatalf("note = %q, want the note of the revision that established the figure", record[0].Note)
	}
}

// The six products whose price moved, and Zoho Books before them, are the
// history. A changed figure is never collapsed, and the revision that predates
// the billing columns publishes no basis rather than a guessed one.
func TestPriceRecordKeepsAMovedPriceAndDoesNotInventABasisForIt(t *testing.T) {
	annual := int64(2900)
	record := PriceRecord([]PriceRevision{
		{ObservedAt: day(2), Price: usd(3900), Billing: monthlyFlat(&annual),
			Note: "Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=flat. Compared price moved from 2900 to 3900 minor units."},
		{ObservedAt: day(1), Price: usd(2900),
			Note: "Price read from the vendor pricing page on 2026-08-21 at annual billing."},
	})

	if len(record) != 2 {
		t.Fatalf("record has %d entries, want both figures: %#v", len(record), record)
	}
	if record[0].Price != usd(3900) || record[1].Price != usd(2900) {
		t.Fatalf("record is not newest first: %#v", record)
	}
	if record[1].Billing != nil {
		t.Fatalf("older billing = %#v, want none stated", record[1].Billing)
	}
	if record[1].IsCurrent {
		t.Fatal("only the newest entry is current")
	}
	// The audit note restates the basis the row already prints in words, so it
	// is dropped rather than published as a sentence about nothing.
	if record[0].Note != "" {
		t.Fatalf("note = %q, want the machine restatement dropped", record[0].Note)
	}
}

// A basis that changed while the figure did not is a real correction: the same
// number on annual billing and on monthly billing are different prices.
func TestPriceRecordKeepsARestatedBasisAtTheSamePrice(t *testing.T) {
	record := PriceRecord([]PriceRevision{
		{ObservedAt: day(2), Price: usd(1000), Billing: monthlyFlat(nil)},
		{ObservedAt: day(1), Price: usd(1000),
			Billing: &Billing{Period: BillingAnnual, Unit: PricingUnitFlat}},
	})

	if len(record) != 2 {
		t.Fatalf("record has %d entries, want the basis correction kept: %#v", len(record), record)
	}
}

func TestPriceRecordStopsAtTenEntries(t *testing.T) {
	revisions := make([]PriceRevision, 0, 14)
	for index := 0; index < 14; index++ {
		revisions = append(revisions, PriceRevision{
			ObservedAt: day(28 - index), Price: usd(int64(1000 + index*100)),
		})
	}

	record := PriceRecord(revisions)

	if len(record) != MaximumPriceRecordEntries {
		t.Fatalf("record has %d entries, want %d", len(record), MaximumPriceRecordEntries)
	}
	if record[0].Price != usd(1000) {
		t.Fatalf("first entry = %#v, want the newest revision", record[0])
	}
}

func TestPriceRecordOfNothingIsEmptyRatherThanAnEntry(t *testing.T) {
	if record := PriceRecord(nil); len(record) != 0 {
		t.Fatalf("record = %#v, want no entries", record)
	}
}

func TestPriceRecordNoteKeepsOneSentenceOrNothing(t *testing.T) {
	cases := map[string]struct {
		raw  string
		want string
	}{
		"a reviewer's sentence": {
			raw:  "Monthly-billing list price corrected to $20 from Zoho official US pricing observed 2026-08-26. The $15 figure requires annual billing.",
			want: "Monthly-billing list price corrected to $20 from Zoho official US pricing observed 2026-08-26.",
		},
		"the audit's date prefix": {
			raw:  "Billing basis audit 2026-09-02: re-read on the vendor's page. Compared price unchanged.",
			want: "Re-read on the vendor's page.",
		},
		"the audit restating the basis": {
			raw:  "Billing basis audit 2026-09-02: monthly only; period=monthly, unit=flat. Compared price unchanged.",
			want: "",
		},
		"a figure with a decimal point": {
			raw:  "2.9% + 30¢ per successful domestic card charge. Unchanged.",
			want: "2.9% + 30¢ per successful domestic card charge.",
		},
		"nothing written":  {raw: "   ", want: ""},
		"a review memo":    {raw: longNote(), want: ""},
		"no full stop":     {raw: "Read from the vendor pricing page", want: "Read from the vendor pricing page"},
		"a bare prefix":    {raw: "Billing basis audit", want: "Billing basis audit"},
		"an unstyled note": {raw: "read on 2026-09-02.", want: "Read on 2026-09-02."},
	}
	for name, testCase := range cases {
		if got := priceRecordNote(testCase.raw); got != testCase.want {
			t.Fatalf("%s: note = %q, want %q", name, got, testCase.want)
		}
	}
}

func longNote() string {
	note := make([]byte, maximumPriceNoteLength+1)
	for index := range note {
		note[index] = 'a'
	}
	return string(note)
}
