package observability

import (
	"strconv"
	"sync"
	"time"

	"github.com/davidjspooner/go-http-server/pkg/metric"
)

// Cache label values used across repositories.
const (
	CacheRefs     = "refs"
	CachePackages = "packages"
	CacheBlobs    = "blobs"
)

var (
	cacheEvents       *metric.CounterVector
	cacheBytes        *metric.CounterVector
	upstreamRequests  *metric.CounterVector
	upstreamDurations *metric.HistogramVector
	metricsMu         sync.Mutex
	registeredMetrics []metric.Metric
)

func init() {
	resetMetricsLocked()
}

func resetMetricsLocked() {
	cacheEvents = metric.MustNewCounterVector(&metric.MetaData{
		Name:      "repoxy_cache_events_total",
		Help:      "Cache lookup results grouped by repository and cache name",
		LabelKeys: []string{"type", "repo", "cache", "result"},
	})
	cacheBytes = metric.MustNewCounterVector(&metric.MetaData{
		Name:      "repoxy_cache_bytes_total",
		Help:      "Bytes written to or served from caches",
		LabelKeys: []string{"type", "repo", "cache", "action"},
	})
	upstreamRequests = metric.MustNewCounterVector(&metric.MetaData{
		Name:      "repoxy_upstream_requests_total",
		Help:      "Upstream round trips initiated by repositories",
		LabelKeys: []string{"type", "repo", "target", "status"},
	})
	upstreamDurations = metric.DefaultCollection().MustNewHistogramVector(&metric.MetaData{
		Name:      "repoxy_upstream_request_duration_seconds",
		Help:      "Latency of upstream requests",
		LabelKeys: []string{"type", "repo", "target", "status"},
	}, nil)
	registeredMetrics = []metric.Metric{
		cacheEvents,
		cacheBytes,
		upstreamRequests,
		upstreamDurations,
	}
}

// ResetForTests reinitializes all observability metrics. It should only be
// invoked from tests to ensure deterministic counter values.
func ResetForTests() {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	if len(registeredMetrics) > 0 {
		metric.Remove(registeredMetrics...)
	}
	resetMetricsLocked()
}

// RecordCacheHit increments the cache hit counter for the provided repository and cache.
func RecordCacheHit(repoType, repoName, cache string) {
	recordCacheEvent(repoType, repoName, cache, "hit")
}

// RecordCacheMiss increments the cache miss counter for the provided repository and cache.
func RecordCacheMiss(repoType, repoName, cache string) {
	recordCacheEvent(repoType, repoName, cache, "miss")
}

// RecordCacheError increments the cache error counter for the provided repository and cache.
func RecordCacheError(repoType, repoName, cache string) {
	recordCacheEvent(repoType, repoName, cache, "error")
}

func recordCacheEvent(repoType, repoName, cache, result string) {
	if cacheEvents == nil {
		return
	}
	_ = cacheEvents.Inc(normalize(repoType, "unknown"), normalize(repoName, "shared"), normalize(cache, "unknown"), normalize(result, "unknown"))
}

// RecordCacheBytes tracks bytes flowing to or from caches for the given action (serve/store).
func RecordCacheBytes(repoType, repoName, cache, action string, n int64) {
	if cacheBytes == nil || n <= 0 {
		return
	}
	_ = cacheBytes.IncN(n, normalize(repoType, "unknown"), normalize(repoName, "shared"), normalize(cache, "unknown"), normalize(action, "unknown"))
}

// ObserveUpstreamRequest records an upstream request result and latency.
func ObserveUpstreamRequest(repoType, repoName, target string, statusCode int, err error, elapsed time.Duration) {
	if upstreamRequests == nil {
		return
	}
	status := "error"
	if err == nil {
		status = strconv.Itoa(statusCode)
	}
	rt := normalize(repoType, "unknown")
	rn := normalize(repoName, "shared")
	tg := normalize(target, "unknown")
	_ = upstreamRequests.Inc(rt, rn, tg, status)
	if err == nil && upstreamDurations != nil {
		upstreamDurations.Observe(elapsed.Seconds(), rt, rn, tg, status)
	}
}

func normalize(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
