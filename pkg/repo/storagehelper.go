package repo

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/go-http-server/pkg/metric"
)

// StorageHelper wraps a WritableFS and provides streaming helpers for arbitrary paths.
type StorageHelper struct {
	fs       storage.WritableFS
	repoType string
	repoName string
}

var (
	storageHelperOps = metric.MustNewCounterVector(&metric.MetaData{
		Name:      "repoxy_storage_operations_total",
		Help:      "Count of storage helper operations",
		LabelKeys: []string{"type", "repo", "op", "result"},
	})
	storageHelperBytes = metric.MustNewCounterVector(&metric.MetaData{
		Name:      "repoxy_storage_bytes_total",
		Help:      "Bytes written via storage helper operations",
		LabelKeys: []string{"type", "repo", "op"},
	})
)

// NewStorageHelper constructs a helper around the provided filesystem.
func NewStorageHelper(fs storage.WritableFS, repoType, repoName string) (*StorageHelper, error) {
	if fs == nil {
		return nil, fmt.Errorf("storage filesystem is nil")
	}
	if repoType == "" {
		repoType = "unknown"
	}
	if repoName == "" {
		repoName = "shared"
	}
	return &StorageHelper{
		fs:       fs,
		repoType: repoType,
		repoName: repoName,
	}, nil
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
		h.recordOp("store", "error")
		return 0, fmt.Errorf("nil reader")
	}
	if err := h.ensureDir(ctx, relPath); err != nil {
		h.recordOp("store", "error")
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
		h.recordOp("store", "error")
		return n, err
	}
	if err := w.Close(); err != nil {
		w = nil
		h.recordOp("store", "error")
		return n, err
	}
	w = nil
	h.recordBytes("store", n)
	h.recordOp("store", "success")
	return n, nil
}

// Open opens a stored file for reading.
func (h *StorageHelper) Open(ctx context.Context, relPath string) (io.ReadCloser, error) {
	rc, err := h.fs.Open(ctx, relPath)
	if err != nil {
		h.recordOp("open", "error")
		return nil, err
	}
	h.recordOp("open", "success")
	return rc, nil
}

// Stat returns FileMetaData for the stored file.
func (h *StorageHelper) Stat(ctx context.Context, relPath string) (storage.FileMetaData, error) {
	info, err := h.fs.Stat(ctx, relPath)
	if err != nil {
		h.recordOp("stat", "error")
		return nil, err
	}
	h.recordOp("stat", "success")
	return info, nil
}

// EnsureSub exposes WritableFS.EnsureSub for callers that need nested helpers.
func (h *StorageHelper) EnsureSub(ctx context.Context, relPath string) (storage.WritableFS, error) {
	sub, err := h.fs.EnsureSub(ctx, relPath)
	if err != nil {
		h.recordOp("ensure_sub", "error")
		return nil, err
	}
	h.recordOp("ensure_sub", "success")
	return sub, nil
}

// FS returns the underlying WritableFS.
func (h *StorageHelper) FS() storage.WritableFS {
	return h.fs
}

func (h *StorageHelper) recordOp(op, result string) {
	if h == nil {
		return
	}
	_ = storageHelperOps.Inc(h.repoType, h.repoName, op, result)
}

func (h *StorageHelper) recordBytes(op string, n int64) {
	if h == nil || n <= 0 {
		return
	}
	_ = storageHelperBytes.IncN(n, h.repoType, h.repoName, op)
}
