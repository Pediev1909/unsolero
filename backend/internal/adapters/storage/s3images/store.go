package s3images

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"rigmark/internal/adapters/storage/mediaobject"
	adminports "rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	Secure    bool
}

type Store struct {
	client *minio.Client
	bucket string
}

func New(config Config) (*Store, error) {
	if strings.TrimSpace(config.Endpoint) == "" || strings.TrimSpace(config.AccessKey) == "" ||
		strings.TrimSpace(config.SecretKey) == "" || strings.TrimSpace(config.Bucket) == "" {
		return nil, errors.New("complete S3-compatible media configuration is required")
	}
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""), Secure: config.Secure, Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("configure S3-compatible media client: %w", err)
	}
	return NewWithClient(client, config.Bucket)
}

func NewWithClient(client *minio.Client, bucket string) (*Store, error) {
	if client == nil || strings.TrimSpace(bucket) == "" {
		return nil, errors.New("S3-compatible client and bucket are required")
	}
	return &Store{client: client, bucket: bucket}, nil
}

func (store *Store) Save(ctx context.Context, productID catalog.ProductID, data []byte, extension string) (string, bool, error) {
	description, err := mediaobject.Describe(productID, data, extension)
	if err != nil {
		return "", false, err
	}
	options := minio.PutObjectOptions{
		ContentType: description.ContentType,
		UserMetadata: map[string]string{
			"product-id": string(productID), "sha256": description.Digest,
		},
		DisableMultipart: true,
	}
	options.SetMatchETagExcept("*")
	_, err = store.client.PutObject(ctx, store.bucket, description.ObjectKey, bytes.NewReader(data), int64(len(data)), options)
	if err != nil {
		response := minio.ToErrorResponse(err)
		if response.Code == "PreconditionFailed" || response.StatusCode == 412 {
			existing, _, openErr := store.Open(ctx, description.Name)
			if openErr == nil && bytes.Equal(existing, data) {
				return description.Name, false, nil
			}
			return "", false, errors.New("image object key collision")
		}
		return "", false, fmt.Errorf("store image object: %w", err)
	}
	return description.Name, true, nil
}

func (store *Store) Open(ctx context.Context, name string) ([]byte, string, error) {
	description, err := mediaobject.Parse(name)
	if err != nil {
		return nil, "", os.ErrNotExist
	}
	object, err := store.client.GetObject(ctx, store.bucket, description.ObjectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("open image object: %w", err)
	}
	defer object.Close()
	data, err := io.ReadAll(io.LimitReader(object, mediaobject.MaximumImageBytes+1))
	if err != nil {
		if response := minio.ToErrorResponse(err); response.Code == "NoSuchKey" || response.StatusCode == 404 {
			return nil, "", os.ErrNotExist
		}
		return nil, "", fmt.Errorf("read image object: %w", err)
	}
	if !mediaobject.ValidBytes(data, description.Extension) {
		return nil, "", errors.New("stored image object failed integrity validation")
	}
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != description.Digest {
		return nil, "", errors.New("stored image object digest mismatch")
	}
	return data, description.ContentType, nil
}

func (store *Store) Delete(ctx context.Context, name string) error {
	description, err := mediaobject.Parse(name)
	if err != nil {
		return os.ErrNotExist
	}
	if err := store.client.RemoveObject(ctx, store.bucket, description.ObjectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete image object: %w", err)
	}
	return nil
}

func (store *Store) BelongsTo(productID catalog.ProductID, name string) bool {
	return mediaobject.BelongsTo(productID, name)
}

func (store *Store) Ready(ctx context.Context) error {
	exists, err := store.client.BucketExists(ctx, store.bucket)
	if err != nil {
		return fmt.Errorf("S3-compatible media storage unavailable: %w", err)
	}
	if !exists {
		return errors.New("S3-compatible media bucket does not exist")
	}
	return nil
}

func (store *Store) ListObjects(ctx context.Context, cursor string, limit int) (adminports.InventoryPage, error) {
	if limit < 1 || limit > 1000 || strings.ContainsAny(cursor, "\r\n") {
		return adminports.InventoryPage{}, errors.New("invalid media inventory request")
	}
	listCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	result := adminports.InventoryPage{Objects: make([]adminports.InventoryObject, 0, limit)}
	objects := store.client.ListObjects(listCtx, store.bucket, minio.ListObjectsOptions{
		Recursive: true, StartAfter: cursor, MaxKeys: limit + 1,
	})
	for object := range objects {
		if object.Err != nil {
			return adminports.InventoryPage{}, fmt.Errorf("list S3-compatible media inventory: %w", object.Err)
		}
		if len(result.Objects) == limit {
			result.NextCursor = result.Objects[len(result.Objects)-1].Identity
			break
		}
		item := adminports.InventoryObject{Identity: object.Key, Size: object.Size, LastModified: object.LastModified.UTC()}
		if description, err := mediaobject.ParseObjectKey(object.Key); err == nil {
			item.Name, item.Expected = description.Name, true
		}
		result.Objects = append(result.Objects, item)
	}
	return result, nil
}

func (store *Store) StatObject(ctx context.Context, name string) (adminports.InventoryObject, bool, error) {
	description, err := mediaobject.Parse(name)
	if err != nil {
		return adminports.InventoryObject{}, false, nil
	}
	info, err := store.client.StatObject(ctx, store.bucket, description.ObjectKey, minio.StatObjectOptions{})
	if err != nil {
		response := minio.ToErrorResponse(err)
		if response.Code == "NoSuchKey" || response.Code == "NoSuchObject" || response.StatusCode == http.StatusNotFound {
			return adminports.InventoryObject{}, false, nil
		}
		return adminports.InventoryObject{}, false, fmt.Errorf("stat S3-compatible media object: %w", err)
	}
	modified := info.LastModified.UTC()
	if modified.Equal(time.Time{}) {
		modified = time.Time{}
	}
	return adminports.InventoryObject{Identity: description.ObjectKey, Name: name, Expected: true,
		Size: info.Size, LastModified: modified}, true, nil
}
