package tf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"

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
}

var _ repo.Instance = (*tfInstance)(nil)
var _ client.Authenticator = (*tfInstance)(nil)

func NewInstance(config *repo.Repo, fs storage.WritableFS, refs *repo.StorageHelper) (*tfInstance, error) {
	if fs == nil {
		return nil, fmt.Errorf("terraform instance missing filesystem")
	}
	instance := &tfInstance{
		fs:     fs,
		config: *config,
		refs:   refs,
	}
	instance.nameMatchers.Set(config.Mappings)
	instance.pipeline = append(instance.pipeline, client.WithAuthentication(instance))
	return instance, nil
}

func (d *tfInstance) GetMatchWeight(name []string) int {
	return d.nameMatchers.GetMatchWeight(name)
}

func (d *tfInstance) HandleV1VersionList(param *param, w http.ResponseWriter, r *http.Request) {
	if err := d.proxyToUpstream(r.Context(), w, r); err != nil {
		slog.ErrorContext(r.Context(), "failed to proxy terraform version list", "error", err)
		w.WriteHeader(http.StatusBadGateway)
	}
}

func (d *tfInstance) HandleV1Version(param *param, w http.ResponseWriter, r *http.Request) {
	if param == nil || param.version == "" {
		http.Error(w, "missing version", http.StatusBadRequest)
		return
	}
	if d.serveCachedManifest(param, w, r) {
		return
	}
	if err := d.fetchAndStoreManifest(param, w, r); err != nil {
		slog.ErrorContext(r.Context(), "failed to fetch terraform manifest", "error", err)
		w.WriteHeader(http.StatusBadGateway)
	}
}

// for actual download of the provider must return a redirect because the API is written that way
func (d *tfInstance) HandleV1VersionDownload(param *param, w http.ResponseWriter, r *http.Request) {
	// TODO Implement the logic to handle the request for downloading a specific provider version
	slog.DebugContext(r.Context(), "TODO: Implement the logic to handle the request for downloading a specific provider version")
	w.WriteHeader(http.StatusNotImplemented)
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
	if d.refs == nil {
		return false
	}
	reader, err := d.refs.Open(r.Context(), d.manifestRelPath(param))
	if err != nil {
		return false
	}
	defer reader.Close()
	if info, err := d.refs.Stat(r.Context(), d.manifestRelPath(param)); err == nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, reader); err != nil {
		slog.ErrorContext(r.Context(), "failed to stream cached terraform manifest", "error", err)
	}
	return true
}

func (d *tfInstance) fetchAndStoreManifest(param *param, w http.ResponseWriter, r *http.Request) error {
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

	if resp.StatusCode == http.StatusOK && len(body) > 0 && d.refs != nil {
		if _, err := d.refs.Store(r.Context(), d.manifestRelPath(param), bytes.NewReader(body)); err != nil {
			slog.ErrorContext(r.Context(), "failed to persist terraform manifest", "error", err)
		}
	}
	return nil
}

func (d *tfInstance) manifestRelPath(param *param) string {
	filename := fmt.Sprintf("%s.json", param.version)
	return path.Join("providers", param.namespace, param.name, filename)
}
