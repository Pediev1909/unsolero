package localimages

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"rigmark/internal/adapters/storage/mediaobject"
	adminports "rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
)

const maximumImageBytes = mediaobject.MaximumImageBytes

var (
	legacySafeName = regexp.MustCompile(`^[a-f0-9]{32}\.(jpg|png|webp)$`)
)

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

func (store *Store) Save(ctx context.Context, namespace catalog.ProductID, data []byte, extension string) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	description, err := mediaobject.Describe(namespace, data, extension)
	if err != nil {
		return "", false, err
	}
	name := description.Name
	finalPath := filepath.Join(store.directory, name)

	if existing, err := os.ReadFile(finalPath); err == nil {
		if bytes.Equal(existing, data) {
			return name, false, nil
		}
		return "", false, errors.New("image object key collision")
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, fmt.Errorf("inspect image object: %w", err)
	}

	file, err := os.CreateTemp(store.directory, ".image-upload-*")
	if err != nil {
		return "", false, fmt.Errorf("create temporary image object: %w", err)
	}
	temporaryPath := file.Name()
	defer os.Remove(temporaryPath)
	if err = file.Chmod(0o600); err == nil {
		_, err = file.Write(data)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return "", false, fmt.Errorf("write temporary image object: %w", err)
	}
	if err = os.Link(temporaryPath, finalPath); err != nil {
		if existing, readErr := os.ReadFile(finalPath); readErr == nil && bytes.Equal(existing, data) {
			return name, false, nil
		}
		return "", false, fmt.Errorf("publish image object: %w", err)
	}
	return name, true, nil
}

func (store *Store) Open(_ context.Context, name string) ([]byte, string, error) {
	if !safeObjectName(name) {
		return nil, "", os.ErrNotExist
	}
	path := filepath.Join(store.directory, name)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maximumImageBytes {
		if err != nil {
			return nil, "", err
		}
		return nil, "", os.ErrNotExist
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	extension := filepath.Ext(name)
	if !mediaobject.ValidBytes(data, extension) {
		return nil, "", os.ErrNotExist
	}
	mimeType := mediaobject.ContentType(extension)
	return data, mimeType, nil
}

func (store *Store) Delete(_ context.Context, name string) error {
	if !safeObjectName(name) {
		return os.ErrNotExist
	}
	if err := os.Remove(filepath.Join(store.directory, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete image object: %w", err)
	}
	return nil
}

func safeObjectName(name string) bool {
	_, err := mediaobject.Parse(name)
	return legacySafeName.MatchString(name) || err == nil
}

func (store *Store) BelongsTo(productID catalog.ProductID, name string) bool {
	return mediaobject.BelongsTo(productID, name)
}

func (store *Store) Ready(context.Context) error {
	info, err := os.Stat(store.directory)
	if err != nil || !info.IsDir() {
		return errors.New("local image storage is unavailable")
	}
	return nil
}

func (store *Store) ListObjects(ctx context.Context, cursor string, limit int) (adminports.InventoryPage, error) {
	if limit < 1 || limit > 1000 {
		return adminports.InventoryPage{}, errors.New("invalid media inventory limit")
	}
	entries, err := os.ReadDir(store.directory)
	if err != nil {
		return adminports.InventoryPage{}, fmt.Errorf("list local media inventory: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	result := adminports.InventoryPage{Objects: make([]adminports.InventoryObject, 0, limit)}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return adminports.InventoryPage{}, err
		}
		if entry.Name() <= cursor {
			continue
		}
		if len(result.Objects) == limit {
			result.NextCursor = result.Objects[len(result.Objects)-1].Identity
			break
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return adminports.InventoryPage{}, fmt.Errorf("inspect local media inventory entry: %w", infoErr)
		}
		object := adminports.InventoryObject{Identity: entry.Name(), Size: info.Size(), LastModified: info.ModTime().UTC()}
		if _, parseErr := mediaobject.Parse(entry.Name()); parseErr == nil && info.Mode().IsRegular() {
			object.Name, object.Expected = entry.Name(), true
		}
		result.Objects = append(result.Objects, object)
	}
	return result, nil
}

func (store *Store) StatObject(ctx context.Context, name string) (adminports.InventoryObject, bool, error) {
	if err := ctx.Err(); err != nil {
		return adminports.InventoryObject{}, false, err
	}
	if _, err := mediaobject.Parse(name); err != nil {
		return adminports.InventoryObject{}, false, nil
	}
	info, err := os.Lstat(filepath.Join(store.directory, name))
	if errors.Is(err, os.ErrNotExist) {
		return adminports.InventoryObject{}, false, nil
	}
	if err != nil {
		return adminports.InventoryObject{}, false, fmt.Errorf("stat local media object: %w", err)
	}
	if !info.Mode().IsRegular() {
		return adminports.InventoryObject{}, false, nil
	}
	return adminports.InventoryObject{Identity: name, Name: name, Expected: true, Size: info.Size(), LastModified: info.ModTime().UTC()}, true, nil
}
