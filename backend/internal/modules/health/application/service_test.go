package application

import (
	"context"
	"errors"
	"testing"
)

type fakePinger struct {
	err error
}

func (p fakePinger) Ping(context.Context) error {
	return p.err
}

func TestLive(t *testing.T) {
	service := NewService(fakePinger{}, "test-version")
	report := service.Live()

	if report.Status != "ok" || report.Service != "rigmark-api" || report.Version != "test-version" {
		t.Fatalf("Live() returned unexpected report: %+v", report)
	}
	if report.Checks != nil {
		t.Fatalf("Live() checks = %v, want nil", report.Checks)
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus string
		wantCheck  string
		wantErr    bool
	}{
		{name: "database available", wantStatus: "ok", wantCheck: "ok"},
		{
			name:       "database unavailable",
			pingErr:    errors.New("unavailable"),
			wantStatus: "degraded",
			wantCheck:  "unavailable",
			wantErr:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(fakePinger{err: test.pingErr}, "test")
			report, err := service.Check(context.Background())
			if (err != nil) != test.wantErr {
				t.Fatalf("Check() error = %v, wantErr %v", err, test.wantErr)
			}
			if report.Status != test.wantStatus {
				t.Errorf("Status = %q, want %q", report.Status, test.wantStatus)
			}
			if report.Checks["database"] != test.wantCheck {
				t.Errorf("database check = %q, want %q", report.Checks["database"], test.wantCheck)
			}
		})
	}
}
