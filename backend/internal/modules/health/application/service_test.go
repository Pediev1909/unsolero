package application

import (
	"context"
	"errors"
	"testing"
)

type pingerStub struct{ err error }

func (stub pingerStub) Ping(context.Context) error { return stub.err }

type checkerStub struct{ err error }

func (stub checkerStub) Ready(context.Context) error { return stub.err }

func TestReadinessDistinguishesCriticalAndAdvisoryDependencies(t *testing.T) {
	service := NewServiceWithDependencies(pingerStub{}, "test", []Dependency{
		{Name: "alerting", Checker: checkerStub{err: errors.New("disabled")}},
	})
	report, err := service.Check(context.Background())
	if err != nil || report.Status != "degraded" || report.Checks["database"] != "ok" || report.Checks["alerting"] != "unavailable" {
		t.Fatalf("report=%#v err=%v", report, err)
	}

	service = NewService(pingerStub{err: errors.New("down")}, "test")
	report, err = service.Check(context.Background())
	if err == nil || report.Status != "unavailable" {
		t.Fatalf("report=%#v err=%v", report, err)
	}
}
