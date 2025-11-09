package repo

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/davidjspooner/go-fs/pkg/storage"
)

// BlobHelper centralizes creation, lookup, and storage of type-level blob files.
// It enforces the canonical sharding layout defined in requirements/framework/storage-heirachy.md.
type BlobHelper struct {
	storage *StorageHelper
	tmpFS   storage.WritableFS
}

// NewBlobHelper prepares the shared blobs/ and tmp/ subdirectories under a repository type root.
func NewBlobHelper(ctx context.Context, typeFS storage.WritableFS) (*BlobHelper, error) {
	if typeFS == nil {
		return nil, fmt.Errorf("type filesystem is nil")
	}
	blobsFS, err := typeFS.EnsureSub(ctx, "blobs")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare blobs subtree: %w", err)
	}
	tmpFS, err := typeFS.EnsureSub(ctx, "tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare tmp subtree: %w", err)
	}
	storageHelper, err := NewStorageHelper(blobsFS)
	if err != nil {
		return nil, err
	}
	return &BlobHelper{
		storage: storageHelper,
		tmpFS:   tmpFS,
	}, nil
}

// blobPath converts a digest (e.g. sha256:abcd) into the sharded relative path under blobs/.
func (b *BlobHelper) blobPath(digest string) (string, error) {
	algo, hex, err := splitDigest(digest)
	if err != nil {
		return "", err
	}
	shard1 := hex[:2]
	shard2 := hex[2:4]
	return path.Join(algo, shard1, shard2, hex), nil
}

func splitDigest(digest string) (algo, hex string, err error) {
	parts := strings.SplitN(digest, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid digest %q", digest)
	}
	algo = strings.ToLower(parts[0])
	hex = strings.ToLower(parts[1])
	if algo == "" || hex == "" {
		return "", "", fmt.Errorf("invalid digest %q", digest)
	}
	if len(hex) < 4 {
		return "", "", fmt.Errorf("digest %q is too short", digest)
	}
	return algo, hex, nil
}

// ensureShard makes sure the sharded directory exists before writing.
func (b *BlobHelper) ensureShard(ctx context.Context, relPath string) error {
	dir := path.Dir(relPath)
	if dir == "." || dir == "/" {
		return nil
	}
	_, err := b.storage.EnsureSub(ctx, dir)
	return err
}

// Store streams a blob into the shared blob tree by digest.
func (b *BlobHelper) Store(ctx context.Context, digest string, r io.Reader) (int64, error) {
	if r == nil {
		return 0, fmt.Errorf("nil reader")
	}
	rel, err := b.blobPath(digest)
	if err != nil {
		return 0, err
	}
	if err := b.ensureShard(ctx, rel); err != nil {
		return 0, fmt.Errorf("failed to ensure blob directory: %w", err)
	}
	return b.storage.Store(ctx, rel, r)
}

// Adopt copies an already-downloaded file (e.g. in a temp directory) into the blob tree.
func (b *BlobHelper) Adopt(ctx context.Context, digest string, src storage.ReadOnlyFS, srcPath string) (int64, error) {
	if src == nil {
		return 0, fmt.Errorf("source filesystem is nil")
	}
	f, err := src.Open(ctx, srcPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return b.Store(ctx, digest, f)
}

// Open returns a reader for the blob content if present.
func (b *BlobHelper) Open(ctx context.Context, digest string) (io.ReadCloser, error) {
	rel, err := b.blobPath(digest)
	if err != nil {
		return nil, err
	}
	return b.storage.Open(ctx, rel)
}

// Stat reports metadata for a stored blob.
func (b *BlobHelper) Stat(ctx context.Context, digest string) (storage.FileMetaData, error) {
	rel, err := b.blobPath(digest)
	if err != nil {
		return nil, err
	}
	return b.storage.Stat(ctx, rel)
}

// TmpFS exposes the tmp/ subtree for callers that need to stage files manually.
func (b *BlobHelper) TmpFS() storage.WritableFS {
	return b.tmpFS
}
