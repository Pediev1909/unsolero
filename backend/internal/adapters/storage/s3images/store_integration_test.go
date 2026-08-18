package s3images

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"rigmark/internal/adapters/storage/mediaobject"
	catalog "rigmark/internal/modules/catalog/domain"
)

const (
	testProductID      catalog.ProductID = "12345678-1234-4234-8234-123456789abc"
	otherTestProductID catalog.ProductID = "22345678-1234-4234-8234-123456789abc"
)

var (
	testPNG  = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	testPNG2 = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHD2")
)

func TestS3StoreLifecycleIsolationAndConcurrentIdempotency(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()
	ctx := context.Background()

	name, created, err := store.Save(ctx, testProductID, testPNG, ".png")
	if err != nil || !created {
		t.Fatalf("Save() = (%q, %v, %v)", name, created, err)
	}
	data, contentType, err := store.Open(ctx, name)
	if err != nil || !bytes.Equal(data, testPNG) || contentType != "image/png" {
		t.Fatalf("Open() = (%q, %q, %v)", data, contentType, err)
	}
	if !store.BelongsTo(testProductID, name) || store.BelongsTo(otherTestProductID, name) {
		t.Fatal("product ownership check failed")
	}
	object, exists, err := store.StatObject(ctx, name)
	if err != nil || !exists || !object.Expected || object.Name != name {
		t.Fatalf("StatObject() = (%+v, %v, %v)", object, exists, err)
	}
	if _, err = store.client.PutObject(ctx, store.bucket, "incoming/unexpected.bin", strings.NewReader("staging"), 7, minio.PutObjectOptions{}); err != nil {
		t.Fatal(err)
	}
	page, err := store.ListObjects(ctx, "", 10)
	if err != nil {
		t.Fatal(err)
	}
	var expectedFound, unexpectedFound bool
	for _, inventoryObject := range page.Objects {
		expectedFound = expectedFound || inventoryObject.Name == name && inventoryObject.Expected
		unexpectedFound = unexpectedFound || inventoryObject.Identity == "incoming/unexpected.bin" && !inventoryObject.Expected
	}
	if !expectedFound || !unexpectedFound {
		t.Fatalf("inventory did not classify expected and unexpected objects: %+v", page.Objects)
	}

	const writers = 12
	var newlyCreated atomic.Int64
	var wait sync.WaitGroup
	for index := 0; index < writers; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			duplicate, wasCreated, saveErr := store.Save(ctx, testProductID, testPNG2, ".png")
			if saveErr != nil || duplicate == "" {
				t.Errorf("concurrent Save() = (%q, %v, %v)", duplicate, wasCreated, saveErr)
			}
			if wasCreated {
				newlyCreated.Add(1)
			}
		}()
	}
	wait.Wait()
	if got := newlyCreated.Load(); got != 1 {
		t.Fatalf("newly created concurrent objects = %d, want 1", got)
	}

	if err := store.Delete(ctx, name); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Open(ctx, name); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Open() after deletion error = %v", err)
	}
}

func TestS3StoreRejectsInvalidObjectsAndTraversal(t *testing.T) {
	store, cleanup := testStore(t)
	defer cleanup()
	for _, fixture := range []struct {
		data      []byte
		extension string
	}{
		{[]byte("<script>alert(1)</script>"), ".png"},
		{testPNG, ".svg"},
		{make([]byte, mediaobject.MaximumImageBytes+1), ".png"},
	} {
		if _, _, err := store.Save(context.Background(), testProductID, fixture.data, fixture.extension); err == nil {
			t.Fatalf("accepted extension=%q size=%d", fixture.extension, len(fixture.data))
		}
	}
	for _, name := range []string{"../secret", "products/other/file.png", "not-an-object.exe"} {
		if _, _, err := store.Open(context.Background(), name); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Open(%q) error = %v", name, err)
		}
	}
}

func TestS3StoreFailsClosedDuringStorageOutage(t *testing.T) {
	store, err := New(Config{Endpoint: "127.0.0.1:1", AccessKey: "test", SecretKey: "test-secret", Bucket: "unavailable", Secure: false})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if _, _, err := store.Save(ctx, testProductID, testPNG, ".png"); err == nil {
		t.Fatal("Save() succeeded while object storage was unavailable")
	}
}

func testStore(t *testing.T) (*Store, func()) {
	t.Helper()
	endpoint := os.Getenv("TEST_S3_ENDPOINT")
	accessKey := os.Getenv("TEST_S3_ACCESS_KEY")
	secretKey := os.Getenv("TEST_S3_SECRET_KEY")
	if endpoint == "" || accessKey == "" || secretKey == "" {
		t.Skip("S3-compatible integration environment is not configured")
	}
	secure, _ := strconv.ParseBool(os.Getenv("TEST_S3_SECURE"))
	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(accessKey, secretKey, ""), Secure: secure})
	if err != nil {
		t.Fatal(err)
	}
	bucket := fmt.Sprintf("unsolero-media-test-%d", time.Now().UnixNano())
	if err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
		t.Fatalf("create test bucket: %v", err)
	}
	store, err := NewWithClient(client, bucket)
	if err != nil {
		t.Fatal(err)
	}
	return store, func() {
		ctx := context.Background()
		for object := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
			if object.Err == nil {
				_ = client.RemoveObject(ctx, bucket, object.Key, minio.RemoveObjectOptions{})
			}
		}
		_ = client.RemoveBucket(ctx, bucket)
	}
}
