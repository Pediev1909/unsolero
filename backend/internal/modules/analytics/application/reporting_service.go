package application

import (
	"context"

	"rigmark/internal/modules/analytics/domain"
	"rigmark/internal/modules/analytics/ports"
)

const defaultReportRankingLimit = 10

type ReportingService struct {
	repository ports.ReportingRepository
}

func NewReportingService(repository ports.ReportingRepository) *ReportingService {
	return &ReportingService{repository: repository}
}

func (service *ReportingService) Report(ctx context.Context) (domain.Report, error) {
	return service.repository.Report(ctx, defaultReportRankingLimit)
}
