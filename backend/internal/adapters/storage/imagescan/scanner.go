package imagescan

import (
	"context"
	"net/http"

	adminports "rigmark/internal/modules/admin/ports"
)

// Development validates the detected media family. It is deliberately not
// described as malware scanning and central configuration prohibits it in
// production.
type Development struct{}

func (Development) Scan(_ context.Context, data []byte, declaredType string) error {
	if len(data) == 0 || http.DetectContentType(data) != declaredType {
		return adminports.ErrMediaScanUnavailable
	}
	return nil
}

// Unavailable fails uploads closed while allowing the rest of the API to run.
type Unavailable struct{}

func (Unavailable) Scan(context.Context, []byte, string) error {
	return adminports.ErrMediaScanUnavailable
}
