package application

import (
	"context"
	"time"
)

type Pinger interface {
	Ping(context.Context) error
}

type ReadinessChecker interface {
	Ready(context.Context) error
}

type Dependency struct {
	Name     string
	Critical bool
	Checker  ReadinessChecker
}

type Report struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks,omitempty"`
}

type Service struct {
	database     Pinger
	version      string
	timeout      time.Duration
	dependencies []Dependency
}

func NewService(database Pinger, version string) *Service {
	return NewServiceWithDependencies(database, version, nil)
}

func NewServiceWithDependencies(database Pinger, version string, dependencies []Dependency) *Service {
	return &Service{
		database:     database,
		version:      version,
		timeout:      2 * time.Second,
		dependencies: dependencies,
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
		report.Status = "unavailable"
		report.Checks["database"] = "unavailable"
		return report, err
	}
	for _, dependency := range s.dependencies {
		if dependency.Checker == nil {
			continue
		}
		dependencyCtx, cancelDependency := context.WithTimeout(ctx, s.timeout)
		err := dependency.Checker.Ready(dependencyCtx)
		cancelDependency()
		if err == nil {
			report.Checks[dependency.Name] = "ok"
			continue
		}
		report.Checks[dependency.Name] = "unavailable"
		if dependency.Critical {
			report.Status = "unavailable"
			return report, err
		}
		report.Status = "degraded"
	}

	return report, nil
}
