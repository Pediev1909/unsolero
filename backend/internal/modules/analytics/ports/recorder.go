package ports

import (
	"context"

	"rigmark/internal/modules/analytics/domain"
)

type EventRecorder interface {
	Record(context.Context, domain.Event) error
}
