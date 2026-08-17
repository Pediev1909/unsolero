package localimages

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var safeName = regexp.MustCompile(`^[a-f0-9]{32}\.(jpg|png|webp)$`)

type Store struct {
	directory string
}

func New(directory string) (*Store, error) {
	if directory == "" {
		return nil, errors.New("image upload directory is required")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create image upload directory: %w", err)
	}
	return &Store{directory: directory}, nil
}

func (store *Store) Save(_ context.Context, data []byte, extension string) (string, error) {
	if extension != ".jpg" && extension != ".png" && extension != ".webp" {
		return "", errors.New("unsupported image extension")
	}
	identifier := make([]byte, 16)
	if _, err := rand.Read(identifier); err != nil {
		return "", fmt.Errorf("generate image identifier: %w", err)
	}
	name := hex.EncodeToString(identifier) + extension
	file, err := os.OpenFile(filepath.Join(store.directory, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("create image file: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		_ = os.Remove(filepath.Join(store.directory, name))
		return "", fmt.Errorf("write image file: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(filepath.Join(store.directory, name))
		return "", fmt.Errorf("close image file: %w", err)
	}
	return name, nil
}

func (store *Store) Open(_ context.Context, name string) ([]byte, string, error) {
	if !safeName.MatchString(name) {
		return nil, "", os.ErrNotExist
	}
	data, err := os.ReadFile(filepath.Join(store.directory, name))
	if err != nil {
		return nil, "", err
	}
	mimeType := "image/" + filepath.Ext(name)[1:]
	if filepath.Ext(name) == ".jpg" {
		mimeType = "image/jpeg"
	}
	return data, mimeType, nil
}

func (store *Store) Delete(_ context.Context, name string) error {
	if !safeName.MatchString(name) {
		return os.ErrNotExist
	}
	if err := os.Remove(filepath.Join(store.directory, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete image file: %w", err)
	}
	return nil
}
