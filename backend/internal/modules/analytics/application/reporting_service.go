package application

import (
	"context"
	"errors"
	"time"

	"rigmark/internal/modules/analytics/domain"
	"rigmark/internal/modules/analytics/ports"
)

const defaultReportRankingLimit = 10
const maximumReportWindow = 366 * 24 * time.Hour

var ErrInvalidReportQuery = errors.New("invalid analytics report query")

type ReportingService struct {
	repository ports.ReportingRepository
}

func NewReportingService(repository ports.ReportingRepository) *ReportingService {
	return &ReportingService{repository: repository}
}

func (service *ReportingService) Report(ctx context.Context, query domain.ReportQuery) (domain.Report, error) {
	now := time.Now().UTC()
	if query.To.IsZero() {
		query.To = now
	}
	if query.From.IsZero() {
		query.From = query.To.Add(-30 * 24 * time.Hour)
	}
	if query.Limit == 0 {
		query.Limit = defaultReportRankingLimit
	}
	if !query.From.Before(query.To) || query.To.After(now.Add(time.Minute)) || query.To.Sub(query.From) > maximumReportWindow || query.Limit < 1 || query.Limit > 50 {
		return domain.Report{}, ErrInvalidReportQuery
	}
	return service.repository.Report(ctx, query)
}
