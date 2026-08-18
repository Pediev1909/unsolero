package app

import (
	"errors"
	"fmt"

	"rigmark/internal/adapters/storage/localimages"
	"rigmark/internal/adapters/storage/s3images"
	adminports "rigmark/internal/modules/admin/ports"
	"rigmark/internal/platform/config"
)

func NewImageStorage(assets config.Assets) (adminports.ImageStorage, error) {
	var (
		storage adminports.ImageStorage
		err     error
	)
	switch assets.StorageProvider {
	case "local":
		storage, err = localimages.New(assets.ProductImageDirectory)
	case "s3":
		storage, err = s3images.New(s3images.Config{
			Endpoint: assets.S3Endpoint, AccessKey: assets.S3AccessKey, SecretKey: assets.S3SecretKey,
			Bucket: assets.S3Bucket, Region: assets.S3Region, Secure: assets.S3Secure,
		})
	case "external":
		return nil, errors.New("MEDIA_STORAGE_PROVIDER=external requires a reviewed adapter")
	default:
		return nil, errors.New("unsupported media storage provider")
	}
	if err != nil {
		return nil, fmt.Errorf("configure product image storage: %w", err)
	}
	return storage, nil
}
