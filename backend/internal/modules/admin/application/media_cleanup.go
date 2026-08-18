package application

import (
	"context"
	"errors"
	"time"

	"rigmark/internal/modules/admin/ports"
)

type MediaCleanupService struct {
	repository ports.MediaDeletionRepository
	storage    ports.ImageStorage
	now        func() time.Time
}

func NewMediaCleanupService(repository ports.MediaDeletionRepository, storage ports.ImageStorage) *MediaCleanupService {
	return &MediaCleanupService{repository: repository, storage: storage, now: time.Now}
}

func (service *MediaCleanupService) Process(ctx context.Context, limit int) (int, error) {
	if service.repository == nil || service.storage == nil || limit < 1 || limit > 1000 {
		return 0, ErrInvalidInput
	}
	jobs, err := service.repository.ClaimMediaDeletions(ctx, limit, service.now().UTC())
	if err != nil {
		return 0, err
	}
	processed := 0
	var failures []error
	for _, job := range jobs {
		if err := service.storage.Delete(ctx, job.ObjectName); err != nil {
			failures = append(failures, err)
			if recordErr := service.repository.FailMediaDeletion(ctx, job.ObjectName, "storage.delete_failed", service.now().UTC()); recordErr != nil {
				failures = append(failures, recordErr)
			}
			continue
		}
		if err := service.repository.CompleteMediaDeletion(ctx, job.ObjectName, service.now().UTC()); err != nil {
			failures = append(failures, err)
			continue
		}
		processed++
	}
	return processed, errors.Join(failures...)
}
