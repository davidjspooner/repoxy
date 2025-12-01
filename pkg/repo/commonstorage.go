package repo

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/go-http-server/pkg/metric"
)

const (
	metadataRootDir  = "metadata"
	metadataIndexDir = "index"
	blobsRootDir     = "blobs"
	labelsFileName   = "labels.json"
	versionsDirName  = "versions"

	labelKind   = "registry.labels"
	versionKind = "registry.version"

	gregorianUnixOffset = 122192928000000000 // 100-ns intervals between 1582-10-15 and 1970-01-01
)

// Locator identifies a logical object within the registry abstraction.
type Locator struct {
	Host      string
	Name      string
	Label     string
	VersionID string
	FileName  string
}

// FileEntry describes one file belonging to a version.
type FileEntry struct {
	Name      string `json:"name"`
	BlobKey   string `json:"blobKey"`
	Size      int64  `json:"size"`
	MediaType string `json:"mediaType"`
}

// VersionMeta mirrors versions/<uuid>.json on disk.
type VersionMeta struct {
	Kind      string      `json:"kind"`
	VersionID string      `json:"versionId"`
	Host      string      `json:"host"`
	Name      string      `json:"name"`
	CreatedAt time.Time   `json:"createdAt"`
	Files     []FileEntry `json:"files"`
	Manifest  string      `json:"manifest,omitempty"`
}

// LabelBindings mirrors labels.json on disk.
type LabelBindings struct {
	Kind      string            `json:"kind"`
	Host      string            `json:"host"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// VersionSummary is a lightweight view for listing.
type VersionSummary struct {
	VersionID string    `json:"versionId"`
	CreatedAt time.Time `json:"createdAt"`
}

// CommonStorage abstracts higher-level helpers on top of a WritableFS.
type CommonStorage interface {
	ListHosts(ctx context.Context) ([]string, error)
	ListNamesForHost(ctx context.Context, host string) ([]string, error)
	ResolveLabel(ctx context.Context, loc Locator) (Locator, error)
	GetLabels(ctx context.Context, loc Locator) (*LabelBindings, error)
	ListVersions(ctx context.Context, loc Locator) ([]VersionSummary, error)
	GetVersionMeta(ctx context.Context, loc Locator) (*VersionMeta, error)
	OpenBlob(ctx context.Context, blobKey string) (io.ReadCloser, error)
	StatBlob(ctx context.Context, blobKey string) (storage.FileMetaData, error)
	PutBlob(ctx context.Context, blobKey string, r io.Reader) (int64, error)
	CreateVersion(ctx context.Context, loc Locator, meta *VersionMeta) (Locator, error)
	SetLabel(ctx context.Context, loc Locator) error
	DeleteLabel(ctx context.Context, loc Locator) error
	DeleteVersion(ctx context.Context, loc Locator) error
	StoreFile(ctx context.Context, rel string, r io.Reader) (int64, error)
	OpenFile(ctx context.Context, rel string) (io.ReadCloser, error)
	StatFile(ctx context.Context, rel string) (storage.FileMetaData, error)
	EnsureSub(ctx context.Context, rel string) (storage.WritableFS, error)
}

// CommonStorageImpl implements CommonStorage using a WritableFS.
type CommonStorageImpl struct {
	fs         storage.WritableFS
	mu         sync.RWMutex
	pathMu     sync.Mutex
	metadataFS storage.WritableFS
	blobsFS    storage.WritableFS
	metricType string
	metricRepo string
}

var (
	commonStorageOps = metric.MustNewCounterVector(&metric.MetaData{
		Name:      "repoxy_storage_operations_total",
		Help:      "Count of common storage operations",
		LabelKeys: []string{"type", "repo", "op", "result"},
	})
	commonStorageBytes = metric.MustNewCounterVector(&metric.MetaData{
		Name:      "repoxy_storage_bytes_total",
		Help:      "Bytes written via common storage operations",
		LabelKeys: []string{"type", "repo", "op"},
	})
)

// NewCommonStorage constructs a CommonStorage around a WritableFS.
func NewCommonStorage(fs storage.WritableFS) (CommonStorage, error) {
	return NewCommonStorageWithLabels(fs, "", "")
}

// NewCommonStorageWithLabels constructs the helper with explicit metric labels.
func NewCommonStorageWithLabels(fs storage.WritableFS, metricType, metricRepo string) (CommonStorage, error) {
	if fs == nil {
		return nil, fmt.Errorf("common storage requires a writable filesystem")
	}
	if metricType == "" {
		metricType = "unknown"
	}
	if metricRepo == "" {
		metricRepo = "shared"
	}
	return &CommonStorageImpl{
		fs:         fs,
		metricType: metricType,
		metricRepo: metricRepo,
	}, nil
}

// ListHosts returns all known hosts in ascending lexicographic order.
func (s *CommonStorageImpl) ListHosts(ctx context.Context) ([]string, error) {
	metaFS, err := s.metadataIndexFS(ctx)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := metaFS.ReadDir(ctx, "")
	if err != nil {
		if isNotFoundError(err) {
			return []string{}, nil
		}
		return nil, err
	}
	hosts := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			hosts = append(hosts, entry.Name())
		}
	}
	sort.Strings(hosts)
	return hosts, nil
}

// ListNamesForHost enumerates repository names beneath the provided host.
func (s *CommonStorageImpl) ListNamesForHost(ctx context.Context, host string) ([]string, error) {
	host, err := sanitizeHost(host)
	if err != nil {
		return nil, err
	}
	metaFS, err := s.metadataIndexFS(ctx)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, err := metaFS.Stat(ctx, host); err != nil {
		if isNotFoundError(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var names []string
	if err := s.collectNames(ctx, metaFS, host, "", &names); err != nil {
		if isNotFoundError(err) {
			return []string{}, nil
		}
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

// ResolveLabel resolves a label to a concrete version ID.
func (s *CommonStorageImpl) ResolveLabel(ctx context.Context, loc Locator) (Locator, error) {
	if loc.Label == "" {
		return loc, fmt.Errorf("label is required")
	}
	bindings, err := s.GetLabels(ctx, loc)
	if err != nil {
		return loc, err
	}
	versionID, ok := bindings.Labels[loc.Label]
	if !ok || versionID == "" {
		return loc, fmt.Errorf("label %q not found for %s/%s", loc.Label, bindings.Host, bindings.Name)
	}
	loc.VersionID = versionID
	return loc, nil
}

// GetLabels loads label bindings for the provided host/name.
func (s *CommonStorageImpl) GetLabels(ctx context.Context, loc Locator) (*LabelBindings, error) {
	host, name, err := sanitizeLocator(loc)
	if err != nil {
		return nil, err
	}
	metaFS, err := s.metadataIndexFS(ctx)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	labels, err := s.readLabels(ctx, metaFS, host, name)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

// ListVersions returns summaries for every recorded version under host/name.
func (s *CommonStorageImpl) ListVersions(ctx context.Context, loc Locator) ([]VersionSummary, error) {
	host, name, err := sanitizeLocator(loc)
	if err != nil {
		return nil, err
	}
	metaFS, err := s.metadataIndexFS(ctx)
	if err != nil {
		return nil, err
	}
	versionDir := path.Join(host, name, versionsDirName)

	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := metaFS.ReadDir(ctx, versionDir)
	if err != nil {
		if isNotFoundError(err) {
			return []VersionSummary{}, nil
		}
		return nil, err
	}
	summaries := make([]VersionSummary, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		meta, err := s.readVersionMeta(ctx, metaFS, path.Join(versionDir, name))
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, VersionSummary{
			VersionID: meta.VersionID,
			CreatedAt: meta.CreatedAt,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].CreatedAt.Equal(summaries[j].CreatedAt) {
			return summaries[i].VersionID < summaries[j].VersionID
		}
		return summaries[i].CreatedAt.After(summaries[j].CreatedAt)
	})
	return summaries, nil
}

// GetVersionMeta loads the metadata for a specific version.
func (s *CommonStorageImpl) GetVersionMeta(ctx context.Context, loc Locator) (*VersionMeta, error) {
	host, name, err := sanitizeLocator(loc)
	if err != nil {
		return nil, err
	}
	versionID, err := sanitizeVersionID(loc.VersionID)
	if err != nil {
		return nil, err
	}
	metaFS, err := s.metadataIndexFS(ctx)
	if err != nil {
		return nil, err
	}
	rel := path.Join(host, name, versionsDirName, fmt.Sprintf("%s.json", versionID))

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.readVersionMeta(ctx, metaFS, rel)
}

// OpenBlob returns a reader for the blob associated with blobKey.
func (s *CommonStorageImpl) OpenBlob(ctx context.Context, blobKey string) (io.ReadCloser, error) {
	rel, err := blobRelativePath(blobKey)
	if err != nil {
		return nil, err
	}
	blobsFS, err := s.blobsRoot(ctx)
	if err != nil {
		return nil, err
	}
	return blobsFS.Open(ctx, rel)
}

// StatBlob reports metadata for the blob associated with blobKey.
func (s *CommonStorageImpl) StatBlob(ctx context.Context, blobKey string) (storage.FileMetaData, error) {
	rel, err := blobRelativePath(blobKey)
	if err != nil {
		return nil, err
	}
	blobsFS, err := s.blobsRoot(ctx)
	if err != nil {
		return nil, err
	}
	return blobsFS.Stat(ctx, rel)
}

// PutBlob stores blob content if it does not already exist and returns bytes written.
func (s *CommonStorageImpl) PutBlob(ctx context.Context, blobKey string, r io.Reader) (int64, error) {
	if r == nil {
		s.recordOp("put_blob", "error")
		return 0, fmt.Errorf("nil reader")
	}
	rel, err := blobRelativePath(blobKey)
	if err != nil {
		s.recordOp("put_blob", "error")
		return 0, err
	}
	blobsFS, err := s.blobsRoot(ctx)
	// Acquire stats first to avoid unnecessary writes.
	if err != nil {
		s.recordOp("put_blob", "error")
		return 0, err
	}
	if _, err := blobsFS.Stat(ctx, rel); err == nil {
		s.recordOp("put_blob", "success")
		s.recordBytes("put_blob", 0)
		return 0, nil
	} else if !isNotFoundError(err) {
		s.recordOp("put_blob", "error")
		return 0, err
	}
	if err := ensureParentDir(ctx, blobsFS, rel); err != nil {
		s.recordOp("put_blob", "error")
		return 0, err
	}
	w := storage.NewWriter(ctx, blobsFS, rel, nil)
	defer func() {
		if w != nil {
			_ = w.CloseWithError(fmt.Errorf("aborted"))
		}
	}()
	n, err := io.Copy(w, r)
	if err != nil {
		_ = w.CloseWithError(err)
		w = nil
		s.recordBytes("put_blob", 0)
		s.recordOp("put_blob", "error")
		return n, err
	}
	if err := w.Close(); err != nil {
		w = nil
		s.recordBytes("put_blob", 0)
		s.recordOp("put_blob", "error")
		return n, err
	}
	w = nil
	s.recordBytes("put_blob", n)
	s.recordOp("put_blob", "success")
	return n, nil
}

// CreateVersion writes a new immutable version metadata file and returns the locator with VersionID populated.
func (s *CommonStorageImpl) CreateVersion(ctx context.Context, loc Locator, meta *VersionMeta) (Locator, error) {
	if meta == nil {
		return loc, fmt.Errorf("version metadata is required")
	}
	host, name, err := sanitizeLocator(loc)
	if err != nil {
		return loc, err
	}
	metaFS, err := s.metadataIndexFS(ctx)
	if err != nil {
		return loc, err
	}

	versionID := meta.VersionID
	if versionID == "" {
		versionID, err = newVersionID()
		if err != nil {
			return loc, fmt.Errorf("generate version id: %w", err)
		}
	}
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = time.Now().UTC()
	}
	meta.Kind = versionKind
	meta.Host = host
	meta.Name = name
	meta.VersionID = versionID

	s.mu.Lock()
	defer s.mu.Unlock()

	nameFS, err := s.ensureNameDir(ctx, metaFS, host, name)
	if err != nil {
		return loc, err
	}
	versionsFS, err := nameFS.EnsureSub(ctx, versionsDirName)
	if err != nil {
		return loc, err
	}
	target := fmt.Sprintf("%s.json", versionID)
	if err := writeJSONAtomic(ctx, versionsFS, target, meta); err != nil {
		return loc, err
	}
	loc.VersionID = versionID
	loc.Host = host
	loc.Name = name
	return loc, nil
}

// SetLabel binds or updates a label for the given version.
func (s *CommonStorageImpl) SetLabel(ctx context.Context, loc Locator) error {
	host, name, err := sanitizeLocator(loc)
	if err != nil {
		return err
	}
	if loc.Label == "" {
		return fmt.Errorf("label is required")
	}
	versionID, err := sanitizeVersionID(loc.VersionID)
	if err != nil {
		return err
	}
	metaFS, err := s.metadataIndexFS(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	labelDoc, err := s.readLabelsOrDefault(ctx, metaFS, host, name)
	if err != nil {
		return err
	}
	if labelDoc.Labels == nil {
		labelDoc.Labels = map[string]string{}
	}
	labelDoc.Kind = labelKind
	labelDoc.Host = host
	labelDoc.Name = name
	labelDoc.Labels[loc.Label] = versionID
	labelDoc.UpdatedAt = time.Now().UTC()

	return s.writeLabels(ctx, metaFS, host, name, labelDoc)
}

// DeleteLabel removes a label binding if it exists.
func (s *CommonStorageImpl) DeleteLabel(ctx context.Context, loc Locator) error {
	host, name, err := sanitizeLocator(loc)
	if err != nil {
		return err
	}
	if loc.Label == "" {
		return fmt.Errorf("label is required")
	}
	metaFS, err := s.metadataIndexFS(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	labelDoc, err := s.readLabels(ctx, metaFS, host, name)
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return err
	}
	if _, exists := labelDoc.Labels[loc.Label]; !exists {
		return nil
	}
	delete(labelDoc.Labels, loc.Label)
	labelDoc.UpdatedAt = time.Now().UTC()
	return s.writeLabels(ctx, metaFS, host, name, labelDoc)
}

// DeleteVersion removes a version metadata file, leaving blobs for a later GC pass.
func (s *CommonStorageImpl) DeleteVersion(ctx context.Context, loc Locator) error {
	host, name, err := sanitizeLocator(loc)
	if err != nil {
		return err
	}
	versionID, err := sanitizeVersionID(loc.VersionID)
	if err != nil {
		return err
	}
	metaFS, err := s.metadataIndexFS(ctx)
	if err != nil {
		return err
	}
	target := path.Join(host, name, versionsDirName, fmt.Sprintf("%s.json", versionID))

	s.mu.Lock()
	defer s.mu.Unlock()

	err = metaFS.Delete(ctx, target, nil)
	if err != nil && !isNotFoundError(err) {
		return err
	}
	return nil
}

// StoreFile writes a file relative to the CommonStorage root, creating parent directories as needed.
func (s *CommonStorageImpl) StoreFile(ctx context.Context, rel string, r io.Reader) (int64, error) {
	if r == nil {
		s.recordOp("store", "error")
		return 0, fmt.Errorf("nil reader")
	}
	if err := ensureParentDir(ctx, s.fs, rel); err != nil {
		s.recordOp("store", "error")
		return 0, err
	}
	w := storage.NewWriter(ctx, s.fs, rel, nil)
	defer func() {
		if w != nil {
			_ = w.CloseWithError(fmt.Errorf("aborted"))
		}
	}()
	n, err := io.Copy(w, r)
	if err != nil {
		_ = w.CloseWithError(err)
		w = nil
		s.recordOp("store", "error")
		return n, err
	}
	if err := w.Close(); err != nil {
		w = nil
		s.recordOp("store", "error")
		return n, err
	}
	w = nil
	s.recordBytes("store", n)
	s.recordOp("store", "success")
	return n, nil
}

// OpenFile opens a file relative to the CommonStorage root.
func (s *CommonStorageImpl) OpenFile(ctx context.Context, rel string) (io.ReadCloser, error) {
	rc, err := s.fs.Open(ctx, rel)
	if err != nil {
		s.recordOp("open", "error")
		return nil, err
	}
	s.recordOp("open", "success")
	return rc, nil
}

// StatFile returns metadata for a file relative to the CommonStorage root.
func (s *CommonStorageImpl) StatFile(ctx context.Context, rel string) (storage.FileMetaData, error) {
	info, err := s.fs.Stat(ctx, rel)
	if err != nil {
		s.recordOp("stat", "error")
		return nil, err
	}
	s.recordOp("stat", "success")
	return info, nil
}

// EnsureSub returns a WritableFS for the given subtree relative to the CommonStorage root.
func (s *CommonStorageImpl) EnsureSub(ctx context.Context, rel string) (storage.WritableFS, error) {
	sub, err := s.fs.EnsureSub(ctx, rel)
	if err != nil {
		s.recordOp("ensure_sub", "error")
		return nil, err
	}
	s.recordOp("ensure_sub", "success")
	return sub, nil
}

func (s *CommonStorageImpl) recordOp(op, result string) {
	if s == nil {
		return
	}
	_ = commonStorageOps.Inc(s.metricType, s.metricRepo, op, result)
}

func (s *CommonStorageImpl) recordBytes(op string, n int64) {
	if s == nil || n <= 0 {
		return
	}
	_ = commonStorageBytes.IncN(n, s.metricType, s.metricRepo, op)
}

// metadataIndexFS returns the writable FS rooted at metadata/index.
func (s *CommonStorageImpl) metadataIndexFS(ctx context.Context) (storage.WritableFS, error) {
	s.pathMu.Lock()
	defer s.pathMu.Unlock()
	if s.metadataFS != nil {
		return s.metadataFS, nil
	}
	metaFS, err := s.fs.EnsureSub(ctx, metadataRootDir)
	if err != nil {
		return nil, fmt.Errorf("ensure metadata root: %w", err)
	}
	indexFS, err := metaFS.EnsureSub(ctx, metadataIndexDir)
	if err != nil {
		return nil, fmt.Errorf("ensure metadata index: %w", err)
	}
	s.metadataFS = indexFS
	return s.metadataFS, nil
}

// blobsRoot returns the writable FS rooted at blobs/.
func (s *CommonStorageImpl) blobsRoot(ctx context.Context) (storage.WritableFS, error) {
	s.pathMu.Lock()
	defer s.pathMu.Unlock()
	if s.blobsFS != nil {
		return s.blobsFS, nil
	}
	blobsFS, err := s.fs.EnsureSub(ctx, blobsRootDir)
	if err != nil {
		return nil, fmt.Errorf("ensure blobs root: %w", err)
	}
	s.blobsFS = blobsFS
	return s.blobsFS, nil
}

// ensureNameDir creates (if needed) the directory for host/name and returns its FS.
func (s *CommonStorageImpl) ensureNameDir(ctx context.Context, metaFS storage.WritableFS, host, name string) (storage.WritableFS, error) {
	rel := path.Join(host, name)
	nameFS, err := metaFS.EnsureSub(ctx, rel)
	if err != nil {
		return nil, err
	}
	return nameFS, nil
}

func (s *CommonStorageImpl) readLabels(ctx context.Context, metaFS storage.WritableFS, host, name string) (*LabelBindings, error) {
	rel := path.Join(host, name, labelsFileName)
	var bindings LabelBindings
	if err := readJSONFile(ctx, metaFS, rel, &bindings); err != nil {
		return nil, err
	}
	if bindings.Labels == nil {
		bindings.Labels = map[string]string{}
	}
	return &bindings, nil
}

func (s *CommonStorageImpl) readLabelsOrDefault(ctx context.Context, metaFS storage.WritableFS, host, name string) (*LabelBindings, error) {
	bindings, err := s.readLabels(ctx, metaFS, host, name)
	if err == nil {
		return bindings, nil
	}
	if !isNotFoundError(err) {
		msg := strings.ToLower(err.Error())
		if !strings.Contains(msg, "cannot open file") {
			return nil, err
		}
	}
	// Treat missing file as an empty set of bindings.
	if _, err := s.ensureNameDir(ctx, metaFS, host, name); err != nil {
		return nil, err
	}
	return &LabelBindings{
		Kind:   labelKind,
		Host:   host,
		Name:   name,
		Labels: map[string]string{},
	}, nil
}

func (s *CommonStorageImpl) writeLabels(ctx context.Context, metaFS storage.WritableFS, host, name string, bindings *LabelBindings) error {
	nameFS, err := s.ensureNameDir(ctx, metaFS, host, name)
	if err != nil {
		return err
	}
	return writeJSONAtomic(ctx, nameFS, labelsFileName, bindings)
}

func (s *CommonStorageImpl) readVersionMeta(ctx context.Context, metaFS storage.WritableFS, rel string) (*VersionMeta, error) {
	var meta VersionMeta
	if err := readJSONFile(ctx, metaFS, rel, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *CommonStorageImpl) collectNames(ctx context.Context, metaFS storage.WritableFS, relPath, prefix string, out *[]string) error {
	entries, err := metaFS.ReadDir(ctx, relPath)
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return err
	}
	hasRepoFiles := false
	subdirs := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			if name == versionsDirName {
				hasRepoFiles = true
				continue
			}
			subdirs = append(subdirs, name)
			continue
		}
		if name == labelsFileName {
			hasRepoFiles = true
		}
	}
	if prefix != "" && hasRepoFiles {
		*out = append(*out, prefix)
	}
	sort.Strings(subdirs)
	for _, dirName := range subdirs {
		childRel := path.Join(relPath, dirName)
		var childPrefix string
		if prefix == "" {
			childPrefix = dirName
		} else {
			childPrefix = path.Join(prefix, dirName)
		}
		if err := s.collectNames(ctx, metaFS, childRel, childPrefix, out); err != nil {
			return err
		}
	}
	return nil
}

func blobRelativePath(digest string) (string, error) {
	algo, hexValue, err := splitDigest(digest)
	if err != nil {
		return "", err
	}
	return path.Join(algo, hexValue[:2], hexValue[2:4], hexValue), nil
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

func ensureParentDir(ctx context.Context, fs storage.WritableFS, rel string) error {
	dir := path.Dir(rel)
	if dir == "." || dir == "/" {
		return nil
	}
	_, err := fs.EnsureSub(ctx, dir)
	return err
}

func sanitizeLocator(loc Locator) (string, string, error) {
	host, err := sanitizeHost(loc.Host)
	if err != nil {
		return "", "", err
	}
	name, err := sanitizeName(loc.Name)
	if err != nil {
		return "", "", err
	}
	return host, name, nil
}

func sanitizeHost(host string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", fmt.Errorf("host is required")
	}
	if strings.Contains(host, "/") {
		return "", fmt.Errorf("host %q must not contain '/'", host)
	}
	if strings.Contains(host, "..") {
		return "", fmt.Errorf("host %q cannot contain '..'", host)
	}
	return host, nil
}

func sanitizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = strings.Trim(name, "/")
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	parts := strings.Split(name, "/")
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			return "", fmt.Errorf("invalid path segment %q", p)
		}
	}
	return strings.Join(parts, "/"), nil
}

func sanitizeVersionID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("version id is required")
	}
	if strings.Contains(id, "/") {
		return "", fmt.Errorf("version id %q must not contain '/'", id)
	}
	return id, nil
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, os.ErrNotExist) {
		return true
	}
	var serr *storage.Error
	if errors.As(err, &serr) {
		if errors.Is(serr.Inner, fs.ErrNotExist) || errors.Is(serr.Inner, os.ErrNotExist) {
			return true
		}
		switch serr.Code {
		case "ESTAT", "EREADDIR", "EOPEN":
			msg := strings.ToLower(serr.Message)
			if strings.Contains(msg, "not found") || strings.Contains(msg, "not a directory") {
				return true
			}
		}
	}
	return false
}

func writeJSONAtomic(ctx context.Context, fs storage.WritableFS, rel string, data any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp-%d", rel, time.Now().UnixNano())
	if err := fs.CreateFrom(ctx, tmp, &buf, nil); err != nil {
		return err
	}
	if err := fs.Rename(ctx, tmp, rel); err != nil {
		_ = fs.Delete(ctx, tmp, &storage.DeleteOptions{})
		return err
	}
	return nil
}

func readJSONFile(ctx context.Context, fs storage.ReadOnlyFS, rel string, dst any) error {
	rc, err := fs.Open(ctx, rel)
	if err != nil {
		return err
	}
	defer rc.Close()
	dec := json.NewDecoder(rc)
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func newVersionID() (string, error) {
	now := time.Now().UTC()
	timestamp := uint64(gregorianUnixOffset + now.UnixNano()/100)
	timeLow := uint32(timestamp & 0xffffffff)
	timeMid := uint16((timestamp >> 32) & 0xffff)
	timeHi := uint16((timestamp >> 48) & 0x0fff)
	timeHi |= 0x1000

	var clockBytes [2]byte
	if _, err := rand.Read(clockBytes[:]); err != nil {
		return "", err
	}
	clockSeq := binary.BigEndian.Uint16(clockBytes[:])
	clockSeq = (clockSeq & 0x3fff) | 0x8000

	var node [6]byte
	if _, err := rand.Read(node[:]); err != nil {
		return "", err
	}
	node[0] |= 0x01

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%02x%02x%02x%02x%02x%02x",
		timeLow, timeMid, timeHi, clockSeq,
		node[0], node[1], node[2], node[3], node[4], node[5],
	), nil
}
