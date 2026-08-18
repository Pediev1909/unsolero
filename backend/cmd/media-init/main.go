package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"rigmark/internal/platform/config"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	if cfg.Environment != "development" && cfg.Environment != "staging" {
		return errors.New("media bucket initialization is restricted to development and staging")
	}
	if cfg.Assets.StorageProvider != "s3" {
		return errors.New("media bucket initialization requires MEDIA_STORAGE_PROVIDER=s3")
	}
	client, err := minio.New(cfg.Assets.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Assets.S3AccessKey, cfg.Assets.S3SecretKey, ""),
		Secure: cfg.Assets.S3Secure, Region: cfg.Assets.S3Region,
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	var lastErr error
	for attempt := 0; attempt < 30; attempt++ {
		exists, existsErr := client.BucketExists(ctx, cfg.Assets.S3Bucket)
		if existsErr == nil && exists {
			fmt.Println("media_bucket=ready")
			return nil
		}
		if existsErr == nil {
			existsErr = client.MakeBucket(ctx, cfg.Assets.S3Bucket, minio.MakeBucketOptions{Region: cfg.Assets.S3Region})
			if existsErr == nil {
				fmt.Println("media_bucket=created")
				return nil
			}
		}
		lastErr = existsErr
		select {
		case <-ctx.Done():
			return fmt.Errorf("initialize staging media bucket: %w", ctx.Err())
		case <-time.After(time.Second):
		}
	}
	return fmt.Errorf("initialize staging media bucket: %w", lastErr)
}
