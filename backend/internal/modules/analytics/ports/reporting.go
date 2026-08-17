package ports

import (
	"context"

	"rigmark/internal/modules/analytics/domain"
)

type ReportingRepository interface {
	Report(context.Context, int) (domain.Report, error)
}
