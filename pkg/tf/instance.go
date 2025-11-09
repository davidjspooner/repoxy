package tf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

type tfInstance struct {
	tofu         bool
	fs           storage.WritableFS
	config       repo.Repo
	pipeline     client.MiddlewarePipeline
	nameMatchers repo.NameMatchers // Matchers for repository names
	refs         *repo.StorageHelper
	packages     *repo.StorageHelper
}

type downloadRequest struct {
	param     *param
	OS        string
	Arch      string
	Filename  string
	IsArchive bool
}

var _ repo.Instance = (*tfInstance)(nil)
var _ client.Authenticator = (*tfInstance)(nil)

func NewInstance(config *repo.Repo, fs storage.WritableFS, refs *repo.StorageHelper, packages *repo.StorageHelper) (*tfInstance, error) {
	if fs == nil {
		return nil, fmt.Errorf("terraform instance missing filesystem")
	}
	if refs == nil {
		return nil, fmt.Errorf("terraform instance missing refs storage")
	}
	if packages == nil {
		return nil, fmt.Errorf("terraform instance missing package storage")
	}
	instance := &tfInstance{
		fs:       fs,
		config:   *config,
		refs:     refs,
		packages: packages,
	}
	instance.nameMatchers.Set(config.Mappings)
	instance.pipeline = append(instance.pipeline, client.WithAuthentication(instance))
	return instance, nil
}

func (d *tfInstance) GetMatchWeight(name []string) int {
	return d.nameMatchers.GetMatchWeight(name)
}

func (d *tfInstance) HandleV1VersionList(param *param, w http.ResponseWriter, r *http.Request) {
	if param == nil || param.namespace == "" || param.name == "" {
		http.Error(w, "missing namespace or name", http.StatusBadRequest)
		return
	}
	if d.serveCachedJSON(d.versionsRelPath(param), w, r) {
		return
	}
	if err := d.fetchAndStoreVersionList(param, w, r); err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch terraform version list", "error", err)
		w.WriteHeader(http.StatusBadGateway)
	}
}

func (d *tfInstance) HandleV1Version(param *param, w http.ResponseWriter, r *http.Request) {
	if param == nil || param.version == "" {
		http.Error(w, "missing version", http.StatusBadRequest)
		return
	}
	if d.serveCachedJSON(d.manifestRelPath(param), w, r) {
		return
	}
	if err := d.fetchAndStoreManifest(param, w, r); err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch terraform manifest", "error", err)
		w.WriteHeader(http.StatusBadGateway)
	}
}

// for actual download of the provider must return a redirect because the API is written that way
func (d *tfInstance) HandleV1VersionDownload(param *param, w http.ResponseWriter, r *http.Request) {
	downloadReq, err := parseDownloadTail(param)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if downloadReq.IsArchive {
		if err := d.servePackageArchive(downloadReq, w, r); err != nil {
			slog.ErrorContext(r.Context(), "failed to serve terraform provider archive", "error", err)
			http.Error(w, "failed to download provider", http.StatusBadGateway)
		}
		return
	}
	if err := d.handleDownloadMetadata(downloadReq, w, r); err != nil {
		slog.ErrorContext(r.Context(), "failed to build terraform download metadata", "error", err)
		http.Error(w, "failed to prepare download metadata", http.StatusBadGateway)
	}
}

func (d *tfInstance) Authenticate(response *http.Response) string {
	// TODO : Implement the logic to authenticate the request according to selected upstream
	slog.DebugContext(response.Request.Context(), "TODO: Implement the logic to authenticate the request according to selected upstream")
	return ""
}

func (d *tfInstance) proxyToUpstream(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	resp, err := d.roundTripUpstream(ctx, r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	d.writeHeadersFromResponse(w, resp)
	w.WriteHeader(resp.StatusCode)
	if resp.Body == nil {
		return nil
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

func (d *tfInstance) roundTripUpstream(ctx context.Context, r *http.Request) (*http.Response, error) {
	u, err := url.Parse(d.config.Upstream.URL)
	if err != nil {
		return nil, err
	}
	u.Path = r.URL.Path
	u.RawQuery = r.URL.RawQuery

	req, err := http.NewRequestWithContext(ctx, r.Method, u.String(), r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header.Clone()
	var c client.Interface = &http.Client{}
	c = d.pipeline.WrapClient(c)
	return c.Do(req)
}

func (d *tfInstance) writeHeadersFromResponse(w http.ResponseWriter, resp *http.Response) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
}

func (d *tfInstance) serveCachedManifest(param *param, w http.ResponseWriter, r *http.Request) bool {
	return d.serveCachedJSON(d.manifestRelPath(param), w, r)
}

func (d *tfInstance) fetchAndStoreManifest(param *param, w http.ResponseWriter, r *http.Request) error {
	return d.fetchAndStoreJSON(d.manifestRelPath(param), w, r)
}

func (d *tfInstance) fetchAndStoreVersionList(param *param, w http.ResponseWriter, r *http.Request) error {
	return d.fetchAndStoreJSON(d.versionsRelPath(param), w, r)
}

func (d *tfInstance) manifestRelPath(param *param) string {
	if param == nil {
		return ""
	}
	filename := fmt.Sprintf("%s.json", param.version)
	return path.Join("providers", param.namespace, param.name, filename)
}

func (d *tfInstance) versionsRelPath(param *param) string {
	if param == nil {
		return ""
	}
	return path.Join("providers", param.namespace, param.name, "versions.json")
}

func (d *tfInstance) serveCachedJSON(relPath string, w http.ResponseWriter, r *http.Request) bool {
	if d.refs == nil || relPath == "" {
		return false
	}
	reader, err := d.refs.Open(r.Context(), relPath)
	if err != nil {
		return false
	}
	defer reader.Close()
	if info, err := d.refs.Stat(r.Context(), relPath); err == nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, reader); err != nil {
		slog.ErrorContext(r.Context(), "failed to stream cached terraform metadata", "error", err, "path", relPath)
	}
	return true
}

func (d *tfInstance) fetchAndStoreJSON(relPath string, w http.ResponseWriter, r *http.Request) error {
	resp, err := d.roundTripUpstream(r.Context(), r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	d.writeHeadersFromResponse(w, resp)
	w.WriteHeader(resp.StatusCode)
	if len(body) > 0 {
		if _, err := w.Write(body); err != nil {
			return err
		}
	}

	if resp.StatusCode == http.StatusOK && len(body) > 0 && d.refs != nil && relPath != "" {
		if _, err := d.refs.Store(r.Context(), relPath, bytes.NewReader(body)); err != nil {
			slog.ErrorContext(r.Context(), "failed to persist terraform metadata", "error", err, "path", relPath)
		}
	}
	return nil
}

func parseDownloadTail(p *param) (*downloadRequest, error) {
	if p == nil {
		return nil, fmt.Errorf("missing provider reference")
	}
	tail := strings.Trim(p.tail, "/")
	if tail == "" {
		return nil, fmt.Errorf("invalid download path")
	}
	parts := strings.Split(tail, "/")
	if len(parts) < 3 || parts[0] != "download" {
		return nil, fmt.Errorf("invalid download path")
	}
	req := &downloadRequest{
		param: p,
		OS:    parts[1],
		Arch:  parts[2],
	}
	if req.OS == "" || req.Arch == "" {
		return nil, fmt.Errorf("missing platform details")
	}
	if len(parts) == 3 {
		return req, nil
	}
	if parts[3] != "archive" {
		return nil, fmt.Errorf("unsupported download path")
	}
	req.IsArchive = true
	if len(parts) > 4 {
		req.Filename = strings.Join(parts[4:], "/")
	}
	return req, nil
}

func (d *tfInstance) handleDownloadMetadata(req *downloadRequest, w http.ResponseWriter, r *http.Request) error {
	resp, err := d.roundTripUpstream(r.Context(), r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		d.writeHeadersFromResponse(w, resp)
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(body)
		return fmt.Errorf("unexpected upstream status: %s", resp.Status)
	}
	payload := make(map[string]any)
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	sourceURL := stringField(payload, "download_url")
	filename := req.Filename
	if filename == "" {
		filename = stringField(payload, "filename")
	}
	if filename == "" {
		filename = filepath.Base(sourceURL)
	}
	if filename == "" {
		return fmt.Errorf("missing filename in upstream response")
	}
	if err := d.ensurePackageCached(r.Context(), req, filename, sourceURL); err != nil {
		return err
	}
	payload["download_url"] = d.localDownloadURL(r, req, filename)
	result, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(result)
	return err
}

func (d *tfInstance) ensurePackageCached(ctx context.Context, req *downloadRequest, filename, sourceURL string) error {
	if d.packages == nil {
		return fmt.Errorf("package storage not configured")
	}
	relPath := d.packageRelPath(req, filename)
	if _, err := d.packages.Stat(ctx, relPath); err == nil {
		return nil
	}
	if sourceURL == "" {
		return fmt.Errorf("missing upstream download url")
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return err
	}
	var c client.Interface = &http.Client{}
	c = d.pipeline.WrapClient(c)
	resp, err := c.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download provider from upstream: %s", resp.Status)
	}
	if _, err := d.packages.Store(ctx, relPath, resp.Body); err != nil {
		return err
	}
	return nil
}

func (d *tfInstance) servePackageArchive(req *downloadRequest, w http.ResponseWriter, r *http.Request) error {
	filename := filepath.Base(req.Filename)
	if filename == "" || filename == "." {
		return fmt.Errorf("missing filename")
	}
	if err := d.ensurePackagePresence(r.Context(), req, filename); err != nil {
		return err
	}
	relPath := d.packageRelPath(req, filename)
	reader, err := d.packages.Open(r.Context(), relPath)
	if err != nil {
		return err
	}
	defer reader.Close()
	if info, err := d.packages.Stat(r.Context(), relPath); err == nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, reader)
	return err
}

func (d *tfInstance) ensurePackagePresence(ctx context.Context, req *downloadRequest, filename string) error {
	relPath := d.packageRelPath(req, filename)
	if _, err := d.packages.Stat(ctx, relPath); err == nil {
		return nil
	}
	payload, err := d.fetchDownloadMetadataMap(ctx, req)
	if err != nil {
		return err
	}
	sourceURL := stringField(payload, "download_url")
	if filename == "" {
		filename = stringField(payload, "filename")
		if filename == "" {
			filename = filepath.Base(sourceURL)
		}
	}
	return d.ensurePackageCached(ctx, req, filename, sourceURL)
}

func (d *tfInstance) fetchDownloadMetadataMap(ctx context.Context, req *downloadRequest) (map[string]any, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid download request")
	}
	path := fmt.Sprintf("/v1/providers/%s/%s/%s/download/%s/%s", req.param.namespace, req.param.name, req.param.version, req.OS, req.Arch)
	fakeReq := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path: path,
		},
		Header: make(http.Header),
	}
	resp, err := d.roundTripUpstream(ctx, fakeReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch upstream metadata: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	payload := make(map[string]any)
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (d *tfInstance) localDownloadURL(r *http.Request, req *downloadRequest, filename string) string {
	scheme := detectScheme(r)
	escaped := url.PathEscape(filename)
	return fmt.Sprintf("%s://%s/v1/providers/%s/%s/%s/download/%s/%s/archive/%s",
		scheme,
		r.Host,
		req.param.namespace,
		req.param.name,
		req.param.version,
		req.OS,
		req.Arch,
		escaped,
	)
}

func (d *tfInstance) packageRelPath(req *downloadRequest, filename string) string {
	safe := filepath.Base(filename)
	if safe == "" || safe == "." {
		safe = "package.zip"
	}
	return path.Join("providers", req.param.namespace, req.param.name, req.param.version, req.OS, req.Arch, safe)
}

func stringField(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
