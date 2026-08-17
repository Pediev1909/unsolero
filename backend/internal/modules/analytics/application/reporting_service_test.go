package application

import (
	"context"
	"testing"

	"rigmark/internal/modules/analytics/domain"
)

type reportingRepositoryStub struct {
	limit int
}

func (repository *reportingRepositoryStub) Report(_ context.Context, limit int) (domain.Report, error) {
	repository.limit = limit
	return domain.Report{Summary: domain.ReportSummary{Users: 3}}, nil
}

func TestReportingServiceUsesBoundedRankingLimit(t *testing.T) {
	repository := &reportingRepositoryStub{}
	report, err := NewReportingService(repository).Report(context.Background())
	if err != nil {
		t.Fatalf("Report(): %v", err)
	}
	if repository.limit != 10 || report.Summary.Users != 3 {
		t.Fatalf("report = %#v, limit = %d", report, repository.limit)
	}
}
