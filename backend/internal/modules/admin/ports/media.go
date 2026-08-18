package ports

import (
	"context"
	"errors"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
)

var ErrMediaScanUnavailable = errors.New("media scanning is unavailable")
var ErrMediaReconciliationRunning = errors.New("media reconciliation is already running")

// ImageStorage is the application-owned contract implemented by local or
// object-storage adapters. Keys are product-scoped and deterministic.
type ImageStorage interface {
	Save(context.Context, catalog.ProductID, []byte, string) (name string, created bool, err error)
	Open(context.Context, string) ([]byte, string, error)
	Delete(context.Context, string) error
	BelongsTo(catalog.ProductID, string) bool
	Ready(context.Context) error
}

type InventoryObject struct {
	// Identity is provider-local and must only be used transiently or hashed.
	Identity     string
	Name         string
	Expected     bool
	Size         int64
	LastModified time.Time
}

type InventoryPage struct {
	Objects    []InventoryObject
	NextCursor string
}

// ImageInventory is deliberately separate from ImageStorage so an adapter
// cannot accidentally acquire destructive reconciliation behavior merely by
// implementing the upload path.
type ImageInventory interface {
	ListObjects(context.Context, string, int) (InventoryPage, error)
	StatObject(context.Context, string) (InventoryObject, bool, error)
}

// ImageScanner must approve bytes before they reach durable storage. A
// production adapter is expected to call a reviewed malware-scanning service.
type ImageScanner interface {
	Scan(context.Context, []byte, string) error
}

type MediaDeletion struct {
	ObjectName   string
	AttemptCount int
	CreatedAt    time.Time
}

type MediaDeletionRepository interface {
	ScheduleMediaDeletion(context.Context, catalog.ProductID, string) error
	ClaimMediaDeletions(context.Context, int, time.Time) ([]MediaDeletion, error)
	CompleteMediaDeletion(context.Context, string, time.Time) error
	FailMediaDeletion(context.Context, string, string, time.Time) error
}

type MediaObjectState struct {
	ReferenceCount       int
	MatchingProductCount int
	DeletionStatus       string
	DeletionUpdatedAt    *time.Time
	DeletionNextAttempt  *time.Time
}

type MediaReference struct {
	ObjectName string
	ProductIDs []catalog.ProductID
}

type MediaReconciliationResult struct {
	Classification string
	IdentifierHash string
	SafeObjectName *string
	Action         string
	DetailCode     string
}

type MediaReconciliationRun struct {
	ID                    string
	Mode                  string
	BatchSize             int
	ObjectCursor          string
	ReferenceCursor       string
	NextObjectCursor      string
	NextReferenceCursor   string
	ObjectsInspected      int
	ReferencesInspected   int
	Discrepancies         int
	DeletionJobsScheduled int
}

type MediaReconciliationRepository interface {
	BeginMediaReconciliation(context.Context, string, int, string, string, time.Time) (string, error)
	InspectMediaObject(context.Context, catalog.ProductID, string) (MediaObjectState, error)
	ListMediaReferences(context.Context, string, int) ([]MediaReference, string, error)
	ListStaleMediaDeletions(context.Context, time.Time, int) ([]MediaDeletion, error)
	RecordMediaReconciliationResult(context.Context, string, MediaReconciliationResult) error
	FinishMediaReconciliation(context.Context, MediaReconciliationRun, time.Time) error
	FailMediaReconciliation(context.Context, string, string, time.Time) error
	MediaDeletionRepository
}
