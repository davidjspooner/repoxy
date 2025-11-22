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

func newCommonStorage(t *testing.T) CommonStorage {
	t.Helper()
	fsRO, err := storage.OpenFileSystemFromString(context.Background(), "mem://", storage.Config{})
	if err != nil {
		t.Fatalf("failed to open mem fs: %v", err)
	}
	fs, ok := fsRO.(storage.WritableFS)
	if !ok {
		t.Fatalf("mem fs is not writable")
	}
	store, err := NewCommonStorageWithLabels(fs, "test", "unit")
	if err != nil {
		t.Fatalf("failed to construct common storage: %v", err)
	}
	return store
}

func TestNewCommonStorageRejectsNil(t *testing.T) {
	t.Parallel()
	if _, err := NewCommonStorage(nil); err == nil {
		t.Fatalf("expected error for nil filesystem")
	}
	if _, err := NewCommonStorageWithLabels(nil, "", ""); err == nil {
		t.Fatalf("expected error for nil filesystem with labels")
	}
}

func TestCommonStorageBlobLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newCommonStorage(t)

	payload := []byte("hello blob")
	n, err := store.PutBlob(ctx, sampleDigest, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("PutBlob failed: %v", err)
	}
	if n != int64(len(payload)) {
		t.Fatalf("expected %d bytes written, got %d", len(payload), n)
	}
	rc, err := store.OpenBlob(ctx, sampleDigest)
	if err != nil {
		t.Fatalf("OpenBlob failed: %v", err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("reading blob failed: %v", err)
	}
	if !bytes.Equal(data, payload) {
		t.Fatalf("expected %q, got %q", payload, data)
	}
	info, err := store.StatBlob(ctx, sampleDigest)
	if err != nil {
		t.Fatalf("StatBlob failed: %v", err)
	}
	if info.Size() != int64(len(payload)) {
		t.Fatalf("expected size %d, got %d", len(payload), info.Size())
	}
	// Second write should be a no-op without error.
	n, err = store.PutBlob(ctx, sampleDigest, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("second PutBlob failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected no bytes written on duplicate, got %d", n)
	}
}

func TestCommonStorageStoreFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newCommonStorage(t)

	content := []byte("test-data")
	if n, err := store.StoreFile(ctx, "a/b/c.txt", bytes.NewReader(content)); err != nil {
		t.Fatalf("StoreFile failed: %v", err)
	} else if n != int64(len(content)) {
		t.Fatalf("expected %d bytes written, got %d", len(content), n)
	}
	rc, err := store.OpenFile(ctx, "a/b/c.txt")
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("expected %q, got %q", content, data)
	}
	info, err := store.StatFile(ctx, "a/b/c.txt")
	if err != nil {
		t.Fatalf("StatFile failed: %v", err)
	}
	if info.Size() != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), info.Size())
	}
}

func TestCommonStorageVersionAndLabels(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newCommonStorage(t)

	loc := Locator{Host: "registry.test", Name: "sample/app"}
	created, err := store.CreateVersion(ctx, loc, &VersionMeta{
		Files: []FileEntry{
			{Name: "manifest.json", BlobKey: sampleDigest, Size: 10, MediaType: "application/json"},
		},
	})
	if err != nil {
		t.Fatalf("CreateVersion failed: %v", err)
	}
	if created.VersionID == "" {
		t.Fatalf("expected version id to be populated")
	}

	loc.VersionID = created.VersionID
	if err := store.SetLabel(ctx, Locator{Host: loc.Host, Name: loc.Name, Label: "latest", VersionID: loc.VersionID}); err != nil {
		t.Fatalf("SetLabel failed: %v", err)
	}
	bindings, err := store.GetLabels(ctx, loc)
	if err != nil {
		t.Fatalf("GetLabels failed: %v", err)
	}
	if bindings.Labels["latest"] != loc.VersionID {
		t.Fatalf("label not set as expected")
	}

	versions, err := store.ListVersions(ctx, loc)
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
}
