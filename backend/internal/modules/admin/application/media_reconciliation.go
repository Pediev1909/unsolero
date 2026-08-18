package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"rigmark/internal/adapters/storage/mediaobject"
	"rigmark/internal/modules/admin/ports"
)

const (
	MediaReconciliationDryRun = "dry_run"
	MediaReconciliationApply  = "apply"

	mediaOrphanObject       = "orphan_object"
	mediaMissingObject      = "missing_object"
	mediaDuplicateReference = "duplicate_reference"
	mediaOwnershipMismatch  = "ownership_mismatch"
	mediaUnexpectedObject   = "unexpected_namespace"
	mediaStaleDeletion      = "stale_deletion"
	mediaUnclassified       = "unclassified"
)

type MediaReconciliationRequest struct {
	Mode            string
	BatchSize       int
	ObjectCursor    string
	ReferenceCursor string
	OrphanGrace     time.Duration
	DeletionLease   time.Duration
}

type MediaReconciliationService struct {
	repository ports.MediaReconciliationRepository
	inventory  ports.ImageInventory
	now        func() time.Time
}

func NewMediaReconciliationService(repository ports.MediaReconciliationRepository, storage ports.ImageStorage) (*MediaReconciliationService, error) {
	inventory, ok := storage.(ports.ImageInventory)
	if repository == nil || !ok {
		return nil, errors.New("media storage does not support bounded inventory")
	}
	return &MediaReconciliationService{repository: repository, inventory: inventory, now: time.Now}, nil
}

func (service *MediaReconciliationService) Reconcile(ctx context.Context, request MediaReconciliationRequest) (ports.MediaReconciliationRun, error) {
	if (request.Mode != MediaReconciliationDryRun && request.Mode != MediaReconciliationApply) ||
		request.BatchSize < 1 || request.BatchSize > 500 || request.OrphanGrace < time.Hour ||
		request.DeletionLease < time.Minute || request.DeletionLease > 24*time.Hour {
		return ports.MediaReconciliationRun{}, ErrInvalidInput
	}
	now := service.now().UTC()
	runID, err := service.repository.BeginMediaReconciliation(ctx, request.Mode, request.BatchSize,
		request.ObjectCursor, request.ReferenceCursor, now)
	if err != nil {
		return ports.MediaReconciliationRun{}, err
	}
	run := ports.MediaReconciliationRun{ID: runID, Mode: request.Mode, BatchSize: request.BatchSize,
		ObjectCursor: request.ObjectCursor, ReferenceCursor: request.ReferenceCursor}
	fail := func(cause error) (ports.MediaReconciliationRun, error) {
		_ = service.repository.FailMediaReconciliation(context.WithoutCancel(ctx), runID, "reconciliation.failed", service.now().UTC())
		return ports.MediaReconciliationRun{}, cause
	}

	page, err := service.inventory.ListObjects(ctx, request.ObjectCursor, request.BatchSize)
	if err != nil {
		return fail(fmt.Errorf("list media objects: %w", err))
	}
	run.NextObjectCursor = page.NextCursor
	for _, object := range page.Objects {
		run.ObjectsInspected++
		if err := service.inspectStoredObject(ctx, &run, object, request, now); err != nil {
			return fail(err)
		}
	}

	references, nextReferenceCursor, err := service.repository.ListMediaReferences(ctx, request.ReferenceCursor, request.BatchSize)
	if err != nil {
		return fail(fmt.Errorf("list media references: %w", err))
	}
	run.NextReferenceCursor = nextReferenceCursor
	for _, reference := range references {
		run.ReferencesInspected++
		if err := service.inspectReference(ctx, &run, reference); err != nil {
			return fail(err)
		}
	}

	stale, err := service.repository.ListStaleMediaDeletions(ctx, now.Add(-request.DeletionLease), request.BatchSize)
	if err != nil {
		return fail(fmt.Errorf("list stale media deletions: %w", err))
	}
	for _, deletion := range stale {
		object, exists, statErr := service.inventory.StatObject(ctx, deletion.ObjectName)
		if statErr != nil {
			return fail(fmt.Errorf("inspect stale media deletion: %w", statErr))
		}
		if !exists {
			continue
		}
		description, parseErr := mediaobject.Parse(deletion.ObjectName)
		if parseErr != nil || !object.Expected {
			if err := service.record(ctx, &run, mediaUnclassified, object.Identity, nil, "none", "deletion.invalid_object_name"); err != nil {
				return fail(err)
			}
			continue
		}
		state, stateErr := service.repository.InspectMediaObject(ctx, description.ProductID, deletion.ObjectName)
		if stateErr != nil {
			return fail(stateErr)
		}
		action := "none"
		if request.Mode == MediaReconciliationApply && state.ReferenceCount == 0 {
			if scheduleErr := service.repository.ScheduleMediaDeletion(ctx, description.ProductID, deletion.ObjectName); scheduleErr != nil {
				return fail(scheduleErr)
			}
			action = "deletion_requeued"
			run.DeletionJobsScheduled++
		}
		if err := service.record(ctx, &run, mediaStaleDeletion, object.Identity, &deletion.ObjectName, action, "deletion.stale_state"); err != nil {
			return fail(err)
		}
	}

	if err := service.repository.FinishMediaReconciliation(ctx, run, service.now().UTC()); err != nil {
		return fail(err)
	}
	return run, nil
}

func (service *MediaReconciliationService) inspectStoredObject(ctx context.Context, run *ports.MediaReconciliationRun,
	object ports.InventoryObject, request MediaReconciliationRequest, now time.Time) error {
	if !object.Expected {
		return service.record(ctx, run, mediaUnexpectedObject, object.Identity, nil, "none", "object.outside_product_namespace")
	}
	description, err := mediaobject.Parse(object.Name)
	if err != nil {
		return service.record(ctx, run, mediaUnclassified, object.Identity, nil, "none", "object.invalid_expected_name")
	}
	state, err := service.repository.InspectMediaObject(ctx, description.ProductID, object.Name)
	if err != nil {
		return fmt.Errorf("inspect media object state: %w", err)
	}
	if state.ReferenceCount > 1 {
		if err := service.record(ctx, run, mediaDuplicateReference, object.Identity, &object.Name, "none", "database.multiple_references"); err != nil {
			return err
		}
	}
	if state.ReferenceCount > state.MatchingProductCount {
		if err := service.record(ctx, run, mediaOwnershipMismatch, object.Identity, &object.Name, "none", "database.product_scope_mismatch"); err != nil {
			return err
		}
	}
	if state.ReferenceCount != 0 {
		return nil
	}
	classification, detail := mediaOrphanObject, "object.unreferenced"
	if state.DeletionStatus != "" {
		stale := state.DeletionStatus == "dead" || state.DeletionStatus == "completed"
		if state.DeletionStatus == "processing" && state.DeletionUpdatedAt != nil {
			stale = !state.DeletionUpdatedAt.After(now.Add(-request.DeletionLease))
		}
		if state.DeletionStatus == "pending" && state.DeletionNextAttempt != nil {
			stale = !state.DeletionNextAttempt.After(now.Add(-request.DeletionLease))
		}
		if !stale {
			return nil
		}
		classification, detail = mediaStaleDeletion, "deletion.object_still_present"
	}
	action := "none"
	aged := !object.LastModified.IsZero() && !object.LastModified.After(now.Add(-request.OrphanGrace))
	if request.Mode == MediaReconciliationApply && aged {
		if err := service.repository.ScheduleMediaDeletion(ctx, description.ProductID, object.Name); err != nil {
			return fmt.Errorf("schedule reconciled media deletion: %w", err)
		}
		action = "deletion_scheduled"
		run.DeletionJobsScheduled++
	} else if request.Mode == MediaReconciliationApply {
		detail = "object.within_safety_grace"
	}
	return service.record(ctx, run, classification, object.Identity, &object.Name, action, detail)
}

func (service *MediaReconciliationService) inspectReference(ctx context.Context, run *ports.MediaReconciliationRun, reference ports.MediaReference) error {
	description, err := mediaobject.Parse(reference.ObjectName)
	if err != nil {
		return service.record(ctx, run, mediaUnclassified, reference.ObjectName, nil, "none", "database.invalid_object_reference")
	}
	if len(reference.ProductIDs) > 1 {
		if err := service.record(ctx, run, mediaDuplicateReference, description.ObjectKey, &reference.ObjectName, "none", "database.multiple_products_reference_object"); err != nil {
			return err
		}
	}
	for _, productID := range reference.ProductIDs {
		if productID != description.ProductID {
			if err := service.record(ctx, run, mediaOwnershipMismatch, description.ObjectKey, &reference.ObjectName, "none", "database.reference_scope_mismatch"); err != nil {
				return err
			}
			break
		}
	}
	_, exists, err := service.inventory.StatObject(ctx, reference.ObjectName)
	if err != nil {
		return fmt.Errorf("stat referenced media object: %w", err)
	}
	if !exists {
		return service.record(ctx, run, mediaMissingObject, description.ObjectKey, &reference.ObjectName, "none", "storage.object_missing")
	}
	return nil
}

func (service *MediaReconciliationService) record(ctx context.Context, run *ports.MediaReconciliationRun,
	classification, identity string, safeName *string, action, detail string) error {
	digest := sha256.Sum256([]byte(identity))
	result := ports.MediaReconciliationResult{Classification: classification, IdentifierHash: hex.EncodeToString(digest[:]),
		SafeObjectName: safeName, Action: action, DetailCode: detail}
	if err := service.repository.RecordMediaReconciliationResult(ctx, run.ID, result); err != nil {
		return fmt.Errorf("record media reconciliation result: %w", err)
	}
	run.Discrepancies++
	return nil
}
