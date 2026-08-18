package application

import (
	"context"
	"testing"
	"time"

	"rigmark/internal/modules/analytics/domain"
)

type reportingRepositoryStub struct {
	query domain.ReportQuery
}

func (repository *reportingRepositoryStub) Report(_ context.Context, query domain.ReportQuery) (domain.Report, error) {
	repository.query = query
	return domain.Report{Summary: domain.ReportSummary{Users: 3}}, nil
}

func TestReportingServiceUsesBoundedRankingLimit(t *testing.T) {
	repository := &reportingRepositoryStub{}
	report, err := NewReportingService(repository).Report(context.Background(), domain.ReportQuery{})
	if err != nil {
		t.Fatalf("Report(): %v", err)
	}
	if repository.query.Limit != 10 || report.Summary.Users != 3 || repository.query.To.Sub(repository.query.From) != 30*24*time.Hour {
		t.Fatalf("report = %#v, query = %#v", report, repository.query)
	}
}
