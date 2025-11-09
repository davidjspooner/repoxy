package repo

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/davidjspooner/go-fs/pkg/storage"
)

// StorageHelper wraps a WritableFS and provides streaming helpers for arbitrary paths.
type StorageHelper struct {
	fs storage.WritableFS
}

// NewStorageHelper constructs a helper around the provided filesystem.
func NewStorageHelper(fs storage.WritableFS) (*StorageHelper, error) {
	if fs == nil {
		return nil, fmt.Errorf("storage filesystem is nil")
	}
	return &StorageHelper{fs: fs}, nil
}

func (h *StorageHelper) ensureDir(ctx context.Context, relPath string) error {
	dir := path.Dir(relPath)
	if dir == "." || dir == "/" {
		return nil
	}
	_, err := h.fs.EnsureSub(ctx, dir)
	return err
}

// Store streams the reader into the given relative path, creating parent directories as needed.
func (h *StorageHelper) Store(ctx context.Context, relPath string, r io.Reader) (int64, error) {
	if r == nil {
		return 0, fmt.Errorf("nil reader")
	}
	if err := h.ensureDir(ctx, relPath); err != nil {
		return 0, err
	}
	w := storage.NewWriter(ctx, h.fs, relPath, nil)
	defer func() {
		if w != nil {
			_ = w.CloseWithError(fmt.Errorf("aborted"))
		}
	}()
	n, err := io.Copy(w, r)
	if err != nil {
		_ = w.CloseWithError(err)
		w = nil
		return n, err
	}
	if err := w.Close(); err != nil {
		w = nil
		return n, err
	}
	w = nil
	return n, nil
}

// Open opens a stored file for reading.
func (h *StorageHelper) Open(ctx context.Context, relPath string) (io.ReadCloser, error) {
	return h.fs.Open(ctx, relPath)
}

// Stat returns FileMetaData for the stored file.
func (h *StorageHelper) Stat(ctx context.Context, relPath string) (storage.FileMetaData, error) {
	return h.fs.Stat(ctx, relPath)
}

// EnsureSub exposes WritableFS.EnsureSub for callers that need nested helpers.
func (h *StorageHelper) EnsureSub(ctx context.Context, relPath string) (storage.WritableFS, error) {
	return h.fs.EnsureSub(ctx, relPath)
}

// FS returns the underlying WritableFS.
func (h *StorageHelper) FS() storage.WritableFS {
	return h.fs
}
