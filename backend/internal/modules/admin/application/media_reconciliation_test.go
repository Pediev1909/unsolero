package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"rigmark/internal/adapters/storage/mediaobject"
	"rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
)

const reconciliationProductID catalog.ProductID = "12345678-1234-4234-8234-123456789abc"

type reconciliationStorage struct {
	ports.ImageStorage
	objects []ports.InventoryObject
	stats   map[string]bool
}

func (storage *reconciliationStorage) ListObjects(context.Context, string, int) (ports.InventoryPage, error) {
	return ports.InventoryPage{Objects: storage.objects}, nil
}
func (storage *reconciliationStorage) StatObject(_ context.Context, name string) (ports.InventoryObject, bool, error) {
	exists := storage.stats[name]
	return ports.InventoryObject{Name: name, Identity: name, Expected: exists}, exists, nil
}
func (*reconciliationStorage) Save(context.Context, catalog.ProductID, []byte, string) (string, bool, error) {
	return "", false, errors.New("not used")
}
func (*reconciliationStorage) Open(context.Context, string) ([]byte, string, error) {
	return nil, "", errors.New("not used")
}
func (*reconciliationStorage) Delete(context.Context, string) error     { return errors.New("not used") }
func (*reconciliationStorage) BelongsTo(catalog.ProductID, string) bool { return false }
func (*reconciliationStorage) Ready(context.Context) error              { return nil }

type reconciliationRepository struct {
	ports.MediaReconciliationRepository
	states        map[string]ports.MediaObjectState
	references    []ports.MediaReference
	stale         []ports.MediaDeletion
	results       []ports.MediaReconciliationResult
	scheduled     map[string]int
	finished      ports.MediaReconciliationRun
	beginError    error
	nextRunNumber int
}

func (repository *reconciliationRepository) BeginMediaReconciliation(context.Context, string, int, string, string, time.Time) (string, error) {
	if repository.beginError != nil {
		return "", repository.beginError
	}
	repository.nextRunNumber++
	return "run", nil
}
func (repository *reconciliationRepository) InspectMediaObject(_ context.Context, _ catalog.ProductID, name string) (ports.MediaObjectState, error) {
	return repository.states[name], nil
}
func (repository *reconciliationRepository) ListMediaReferences(context.Context, string, int) ([]ports.MediaReference, string, error) {
	return repository.references, "", nil
}
func (repository *reconciliationRepository) ListStaleMediaDeletions(context.Context, time.Time, int) ([]ports.MediaDeletion, error) {
	return repository.stale, nil
}
func (repository *reconciliationRepository) RecordMediaReconciliationResult(_ context.Context, _ string, result ports.MediaReconciliationResult) error {
	repository.results = append(repository.results, result)
	return nil
}
func (repository *reconciliationRepository) FinishMediaReconciliation(_ context.Context, run ports.MediaReconciliationRun, _ time.Time) error {
	repository.finished = run
	return nil
}
func (*reconciliationRepository) FailMediaReconciliation(context.Context, string, string, time.Time) error {
	return nil
}
func (repository *reconciliationRepository) ScheduleMediaDeletion(_ context.Context, _ catalog.ProductID, name string) error {
	if repository.scheduled == nil {
		repository.scheduled = map[string]int{}
	}
	repository.scheduled[name]++
	now := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	repository.states[name] = ports.MediaObjectState{DeletionStatus: "pending", DeletionNextAttempt: &now}
	return nil
}

func TestMediaReconciliationCrashWindowsDryRunAndApply(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	orphan := reconciliationObject(t, []byte("orphan"), now.Add(-48*time.Hour))
	referenced := reconciliationObject(t, []byte("referenced"), now.Add(-48*time.Hour))
	missing := reconciliationObject(t, []byte("missing"), now.Add(-48*time.Hour))
	stale := reconciliationObject(t, []byte("stale"), now.Add(-48*time.Hour))
	duplicate := reconciliationObject(t, []byte("duplicate"), now.Add(-48*time.Hour))
	unexpected := ports.InventoryObject{Identity: "incoming/unclassified-secret-name", LastModified: now.Add(-48 * time.Hour)}
	staleAt := now.Add(-2 * time.Hour)
	repository := &reconciliationRepository{
		states: map[string]ports.MediaObjectState{
			referenced.Name: {ReferenceCount: 1, MatchingProductCount: 1},
			stale.Name:      {DeletionStatus: "completed", DeletionUpdatedAt: &staleAt},
			duplicate.Name:  {ReferenceCount: 2, MatchingProductCount: 1},
		},
		references: []ports.MediaReference{
			{ObjectName: referenced.Name, ProductIDs: []catalog.ProductID{reconciliationProductID}},
			{ObjectName: missing.Name, ProductIDs: []catalog.ProductID{reconciliationProductID}},
			{ObjectName: duplicate.Name, ProductIDs: []catalog.ProductID{reconciliationProductID, "22345678-1234-4234-8234-123456789abc"}},
		},
	}
	storage := &reconciliationStorage{
		objects: []ports.InventoryObject{orphan, referenced, stale, duplicate, unexpected},
		stats:   map[string]bool{referenced.Name: true, duplicate.Name: true, stale.Name: true},
	}
	service, err := NewMediaReconciliationService(repository, storage)
	if err != nil {
		t.Fatal(err)
	}
	service.now = func() time.Time { return now }
	request := MediaReconciliationRequest{Mode: MediaReconciliationDryRun, BatchSize: 50,
		OrphanGrace: 24 * time.Hour, DeletionLease: 10 * time.Minute}
	result, err := service.Reconcile(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.DeletionJobsScheduled != 0 || len(repository.scheduled) != 0 {
		t.Fatalf("dry run scheduled deletion: result=%+v scheduled=%v", result, repository.scheduled)
	}
	assertReconciliationClassifications(t, repository.results,
		mediaOrphanObject, mediaMissingObject, mediaStaleDeletion,
		mediaDuplicateReference, mediaOwnershipMismatch, mediaUnexpectedObject)

	repository.results = nil
	request.Mode = MediaReconciliationApply
	result, err = service.Reconcile(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.DeletionJobsScheduled != 2 || repository.scheduled[orphan.Name] != 1 || repository.scheduled[stale.Name] != 1 {
		t.Fatalf("apply did not schedule only safe orphan/stale objects: result=%+v scheduled=%v", result, repository.scheduled)
	}
	// Re-running apply sees fresh pending jobs and does not reset their retry
	// state. This is the interrupted create/delete idempotency guarantee.
	result, err = service.Reconcile(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.DeletionJobsScheduled != 0 || repository.scheduled[orphan.Name] != 1 || repository.scheduled[stale.Name] != 1 {
		t.Fatalf("repeated apply was not idempotent: result=%+v scheduled=%v", result, repository.scheduled)
	}
}

func TestMediaReconciliationYoungUploadCrashWindowStaysQuarantined(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	young := reconciliationObject(t, []byte("young"), now.Add(-5*time.Minute))
	repository := &reconciliationRepository{states: map[string]ports.MediaObjectState{}}
	storage := &reconciliationStorage{objects: []ports.InventoryObject{young}, stats: map[string]bool{young.Name: true}}
	service, _ := NewMediaReconciliationService(repository, storage)
	service.now = func() time.Time { return now }
	result, err := service.Reconcile(context.Background(), MediaReconciliationRequest{Mode: MediaReconciliationApply,
		BatchSize: 10, OrphanGrace: time.Hour, DeletionLease: 10 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if result.DeletionJobsScheduled != 0 || len(repository.scheduled) != 0 || repository.results[0].DetailCode != "object.within_safety_grace" {
		t.Fatalf("young unregistered object was not quarantined: result=%+v findings=%+v", result, repository.results)
	}
}

func reconciliationObject(t *testing.T, data []byte, modified time.Time) ports.InventoryObject {
	t.Helper()
	png := append([]byte("\x89PNG\r\n\x1a\n"), data...)
	for len(png) < 16 {
		png = append(png, 0)
	}
	description, err := mediaobject.Describe(reconciliationProductID, png, ".png")
	if err != nil {
		t.Fatal(err)
	}
	return ports.InventoryObject{Identity: description.ObjectKey, Name: description.Name, Expected: true,
		Size: int64(len(png)), LastModified: modified}
}

func assertReconciliationClassifications(t *testing.T, results []ports.MediaReconciliationResult, expected ...string) {
	t.Helper()
	seen := map[string]bool{}
	for _, result := range results {
		seen[result.Classification] = true
		if result.Classification == mediaUnexpectedObject && result.SafeObjectName != nil {
			t.Fatal("unexpected provider key was persisted instead of only its digest")
		}
	}
	for _, classification := range expected {
		if !seen[classification] {
			t.Fatalf("missing classification %q in %+v", classification, results)
		}
	}
}
