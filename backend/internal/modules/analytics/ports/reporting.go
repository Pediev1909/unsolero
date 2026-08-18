package ports

import (
	"context"

	"rigmark/internal/modules/analytics/domain"
)

type ReportingRepository interface {
	Report(context.Context, domain.ReportQuery) (domain.Report, error)
}
