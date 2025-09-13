package repo

import (
	"net/http"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/go-http-server/pkg/middleware"
)

type Cache struct {
	FS storage.WritableFS
}

var _ middleware.CacheImpl = (*Cache)(nil)

func (c *Cache) TryHandle(w http.ResponseWriter, req *http.Request) (bool, error) {
	// TODO implement cache lookup and serving logic
	return false, nil
}

func (c *Cache) CacheResponce(req *http.Request) middleware.CacheResponse {
	// TODO implement cache writer creation logic
	return &cacheResponse{cache: c, req: req}
}

type cacheResponse struct {
	cache *Cache
	req   *http.Request
}

var _ middleware.CacheResponse = (*cacheResponse)(nil)

func (cw *cacheResponse) Write(p []byte) (n int, err error) {
	// TODO implement cache writing logic
	return len(p), nil
}

func (cw *cacheResponse) Header() http.Header {
	return http.Header{}
}

func (cw *cacheResponse) WriteHeader(statusCode int) {
	// TODO implement status code handling logic
}

func (cw *cacheResponse) Commit(success bool) {
	// TODO implement commit logic
}
