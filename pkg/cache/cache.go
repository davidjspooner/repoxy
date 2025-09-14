package cache

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/go-http-server/pkg/middleware"
)

const (
	CacheMetaFile = "meta.txt"
	CacheBodyFile = "body.bin"
)

type CacheImpl struct {
	FS storage.WritableFS
}

var _ middleware.CacheImpl = (*CacheImpl)(nil)

func (c *CacheImpl) getCacheDir(req *http.Request) string {
	// Build "host[:port]/path" as a relative cache directory
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	p := req.URL.Path
	if len(p) > 0 && p[0] == '/' {
		p = p[1:]
	}
	if p == "" {
		return host
	}
	// Fast concat using strings.Builder
	var b strings.Builder
	b.Grow(len(host) + 1 + len(p))
	b.WriteString(host)
	b.WriteByte('/')
	b.WriteString(p)
	return b.String()
}

func (c *CacheImpl) canCache(req *http.Request) bool {
	// Only cache GET or HEAD
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return false
	}
	// Disallow caching when query parameters are present
	if req.URL != nil && req.URL.RawQuery != "" {
		return false
	}
	// Reject empty or directory-style paths quickly
	p := req.URL.Path
	lastSlash := strings.LastIndex(p, "/")
	if lastSlash == -1 || lastSlash == len(p)-1 {
		return false
	}
	// Disallow caching requests directly for cache files
	tail := p[lastSlash+1:]
	if tail == CacheMetaFile || tail == CacheBodyFile {
		return false
	}
	return true
}

// readHeaders reads HTTP headers from r until the first blank line or error.
// It returns the headers as an http.Header map.
func (c *CacheImpl) readHeaders(r io.Reader) (http.Header, error) {
	h := make(http.Header)
	var lastKey string
	buf := make([]byte, 4096)
	sr := io.NewSectionReader(r.(io.ReaderAt), 0, 1<<63-1)
	for {
		n, err := sr.Read(buf)
		if n > 0 {
			lines := strings.Split(string(buf[:n]), "\n")
			for i, line := range lines {
				line = strings.TrimRight(line, "\r")
				if i == 0 && strings.HasPrefix(line, "HTTP/") {
					// Skip status line
					continue
				}
				if line == "" {
					// End of headers
					return h, nil
				}
				if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
					// Continuation line
					if lastKey != "" {
						h[lastKey][len(h[lastKey])-1] += " " + strings.TrimSpace(line)
					}
				} else {
					// New header line
					parts := strings.SplitN(line, ":", 2)
					if len(parts) != 2 {
						continue // Malformed header line; skip
					}
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					h.Add(key, value)
					lastKey = key
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return h, nil
}

func (c *CacheImpl) TryHandle(w http.ResponseWriter, req *http.Request) (bool, error) {
	if !c.canCache(req) {
		return false, nil
	}
	path := c.getCacheDir(req)
	ctx := req.Context()

	meta, err := c.FS.Open(ctx, path+"/"+CacheMetaFile)
	if err != nil {
		// Cache miss
		return false, nil
	}
	defer meta.Close()

	body, err := c.FS.Open(ctx, path+"/"+CacheBodyFile)
	if err != nil {
		// Cache miss
		return false, nil
	}
	defer body.Close()

	headers, err := c.readHeaders(meta)
	if err != nil {
		// Corrupt cache entry; treat as miss
		return false, nil
	}
	for k, vv := range headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	//read headers from meta ( same format as an imaginary http.Header.Read )
	_, err = io.Copy(w, body)
	if err != nil {
		// log error has happened ( response has already been sent )
		log.Printf("Error handling cached response: %v", err)
	}

	return true, nil
}

func (c *CacheImpl) CacheResponse(req *http.Request) middleware.CacheResponse {
	if !c.canCache(req) {
		return nil
	}
	// TODO implement cache writer creation logic
	return &cacheResponse{
		cache: c, req: req,
		headers: http.Header{},
	}
}

type cacheResponse struct {
	cache      *CacheImpl
	req        *http.Request
	headers    http.Header
	statusCode int
	committed  bool
	w          *storage.Writer
}

var _ middleware.CacheResponse = (*cacheResponse)(nil)

func (cw *cacheResponse) Write(p []byte) (n int, err error) {
	if cw.committed {
		return 0, io.ErrClosedPipe
	}
	if cw.w == nil {
		// Implicitly write headers if not already done
		cw.WriteHeader(http.StatusOK)
	}
	n, err = cw.w.Write(p)
	return n, err
}

func (cw *cacheResponse) Header() http.Header {
	return http.Header{}
}

func (cw *cacheResponse) WriteHeader(statusCode int) {
	cw.statusCode = statusCode
	path := cw.cache.getCacheDir(cw.req)
	meta := storage.NewWriter(cw.req.Context(), cw.cache.FS, path+"/"+CacheMetaFile, nil) //TODO consider what metadata to use
	defer meta.Close()

	cw.w = storage.NewWriter(cw.req.Context(), cw.cache.FS, path+"/"+CacheBodyFile, nil)

	// Write headers to meta
	if err := cw.headers.Write(meta); err != nil {
		cw.Commit(err)
		return
	}
}

func (cw *cacheResponse) Commit(err error) {
	if cw.committed {
		return
	}
	if cw.w != nil {
		if err != nil {
			log.Printf("Error committing cache body writer with error: %v", err)
			if cerr := cw.w.CloseWithError(err); cerr != nil {
				log.Printf("Error closing cache body writer with error: %v", cerr)
			}
		} else {
			if cerr := cw.w.Close(); cerr != nil {
				log.Printf("Error closing cache body writer: %v", cerr)
			}
		}
		cw.w = nil
	}
	cw.committed = true
}
