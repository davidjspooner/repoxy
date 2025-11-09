package repo

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/davidjspooner/go-fs/pkg/storage"
	_ "github.com/davidjspooner/go-fs/pkg/storage/mem"
)

func newWritableMemFS(t *testing.T) storage.WritableFS {
	t.Helper()
	fsRO, err := storage.OpenFileSystemFromString(context.Background(), "mem://", storage.Config{})
	if err != nil {
		t.Fatalf("failed to create mem fs: %v", err)
	}
	fs, ok := fsRO.(storage.WritableFS)
	if !ok {
		t.Fatalf("mem fs not writable")
	}
	return fs
}

func TestStorageHelperStoreAndOpen(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	helper, err := NewStorageHelper(newWritableMemFS(t))
	if err != nil {
		t.Fatalf("NewStorageHelper failed: %v", err)
	}
	content := []byte("test-data")
	if _, err := helper.Store(ctx, "a/b/c.txt", bytes.NewReader(content)); err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	rc, err := helper.Open(ctx, "a/b/c.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("expected %q, got %q", content, data)
	}
	if info, err := helper.Stat(ctx, "a/b/c.txt"); err != nil || info.Size() != int64(len(content)) {
		t.Fatalf("unexpected stat: %v size=%d", err, info.Size())
	}
}
