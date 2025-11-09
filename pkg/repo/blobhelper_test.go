package repo

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/davidjspooner/go-fs/pkg/storage"
	_ "github.com/davidjspooner/go-fs/pkg/storage/mem"
)

const sampleDigest = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd"

func newMemoryFS(t *testing.T) storage.WritableFS {
	t.Helper()
	fsRO, err := storage.OpenFileSystemFromString(context.Background(), "mem://", storage.Config{})
	if err != nil {
		t.Fatalf("failed to open mem fs: %v", err)
	}
	fs, ok := fsRO.(storage.WritableFS)
	if !ok {
		t.Fatalf("mem fs is not writable")
	}
	return fs
}

func TestBlobHelperStoreAndOpen(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fs := newMemoryFS(t)

	helper, err := NewBlobHelper(ctx, "docker", fs)
	if err != nil {
		t.Fatalf("NewBlobHelper failed: %v", err)
	}

	payload := []byte("hello blob")
	if _, err := helper.Store(ctx, sampleDigest, bytes.NewReader(payload)); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	reader, err := helper.Open(ctx, sampleDigest)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if !bytes.Equal(data, payload) {
		t.Fatalf("expected %q, got %q", payload, data)
	}

	info, err := helper.Stat(ctx, sampleDigest)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Size() != int64(len(payload)) {
		t.Fatalf("expected size %d, got %d", len(payload), info.Size())
	}
}

func TestBlobHelperAdopt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fs := newMemoryFS(t)
	helper, err := NewBlobHelper(ctx, "docker", fs)
	if err != nil {
		t.Fatalf("NewBlobHelper failed: %v", err)
	}

	tmp := helper.TmpFS()
	writer := storage.NewWriter(ctx, tmp, "scratch/blob.bin", nil)
	want := []byte("from tmp")
	if _, err := writer.Write(want); err != nil {
		t.Fatalf("write temp data failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("closing temp writer failed: %v", err)
	}

	if _, err := helper.Adopt(ctx, sampleDigest, tmp, "scratch/blob.bin"); err != nil {
		t.Fatalf("Adopt failed: %v", err)
	}

	reader, err := helper.Open(ctx, sampleDigest)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer reader.Close()
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
