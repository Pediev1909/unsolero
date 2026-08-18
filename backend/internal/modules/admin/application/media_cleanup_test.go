package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
)

type cleanupRepository struct {
	ports.MediaDeletionRepository
	jobs      []ports.MediaDeletion
	completed []string
	failed    []string
}

func (repository *cleanupRepository) ClaimMediaDeletions(context.Context, int, time.Time) ([]ports.MediaDeletion, error) {
	return repository.jobs, nil
}
func (repository *cleanupRepository) CompleteMediaDeletion(_ context.Context, name string, _ time.Time) error {
	repository.completed = append(repository.completed, name)
	return nil
}
func (repository *cleanupRepository) FailMediaDeletion(_ context.Context, name, _ string, _ time.Time) error {
	repository.failed = append(repository.failed, name)
	return nil
}

type cleanupStorage struct {
	ports.ImageStorage
	failName string
	deleted  []string
}

func (storage *cleanupStorage) Delete(_ context.Context, name string) error {
	if name == storage.failName {
		return errors.New("synthetic storage outage")
	}
	storage.deleted = append(storage.deleted, name)
	return nil
}
func (*cleanupStorage) BelongsTo(catalog.ProductID, string) bool { return true }

func TestMediaCleanupCompletesAndRetainsFailedJobs(t *testing.T) {
	repository := &cleanupRepository{jobs: []ports.MediaDeletion{{ObjectName: "ok.png"}, {ObjectName: "retry.png"}}}
	storage := &cleanupStorage{failName: "retry.png"}
	service := NewMediaCleanupService(repository, storage)
	processed, err := service.Process(context.Background(), 10)
	if err == nil || processed != 1 {
		t.Fatalf("Process() = (%d, %v)", processed, err)
	}
	if len(repository.completed) != 1 || repository.completed[0] != "ok.png" ||
		len(repository.failed) != 1 || repository.failed[0] != "retry.png" {
		t.Fatalf("completed=%v failed=%v", repository.completed, repository.failed)
	}
}
