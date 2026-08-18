// Package migrations exposes the immutable migration set to binaries that need
// to verify schema compatibility without applying schema changes.
package migrations

import "embed"

// Files is the release migration manifest embedded at build time.
//
//go:embed *.sql
var Files embed.FS
