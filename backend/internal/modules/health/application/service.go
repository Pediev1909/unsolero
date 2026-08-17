package application

import (
	"context"
	"time"
)

type Pinger interface {
	Ping(context.Context) error
}

type Report struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks,omitempty"`
}

type Service struct {
	database Pinger
	version  string
	timeout  time.Duration
}

func NewService(database Pinger, version string) *Service {
	return &Service{
		database: database,
		version:  version,
		timeout:  2 * time.Second,
	}
}

func (s *Service) Live() Report {
	return Report{
		Status:  "ok",
		Service: "rigmark-api",
		Version: s.version,
	}
}

func (s *Service) Check(ctx context.Context) (Report, error) {
	report := s.Live()
	report.Checks = map[string]string{"database": "ok"}

	checkCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := s.database.Ping(checkCtx); err != nil {
		report.Status = "degraded"
		report.Checks["database"] = "unavailable"
		return report, err
	}

	return report, nil
}
