package imagescan

import (
	"context"
	"errors"
	"testing"

	adminports "rigmark/internal/modules/admin/ports"
)

func TestDevelopmentScannerRequiresMatchingDetectedType(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	if err := (Development{}).Scan(context.Background(), png, "image/png"); err != nil {
		t.Fatalf("valid PNG rejected: %v", err)
	}
	if err := (Development{}).Scan(context.Background(), png, "image/jpeg"); !errors.Is(err, adminports.ErrMediaScanUnavailable) {
		t.Fatalf("mismatched type error = %v", err)
	}
}

func TestUnavailableScannerFailsClosed(t *testing.T) {
	if err := (Unavailable{}).Scan(context.Background(), []byte("data"), "image/png"); !errors.Is(err, adminports.ErrMediaScanUnavailable) {
		t.Fatalf("error = %v", err)
	}
}
