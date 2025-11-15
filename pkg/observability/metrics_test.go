package observability

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/davidjspooner/go-http-server/pkg/metric"
)

type collectingEmitter struct {
	values map[string]float64
}

func (c *collectingEmitter) Emit(meta *metric.MetaData, value float64, labelValues ...string) error {
	return c.EmitZipped(meta, value, zipForTest(meta.LabelKeys, labelValues))
}

func (c *collectingEmitter) EmitZipped(meta *metric.MetaData, value float64, zipped string) error {
	if c.values == nil {
		c.values = make(map[string]float64)
	}
	key := fmt.Sprintf("%s|%s", meta.Name, zipped)
	c.values[key] += value
	return nil
}

func zipForTest(keys, values []string) string {
	if len(keys) != len(values) {
		return ""
	}
	parts := make([]string, len(keys))
	for i := range keys {
		parts[i] = fmt.Sprintf("%s=%q", keys[i], values[i])
	}
	return strings.Join(parts, ",")
}

func snapshotMetrics(t *testing.T) map[string]float64 {
	t.Helper()
	emitter := &collectingEmitter{}
	if err := metric.DefaultCollection().EmitInto(emitter); err != nil {
		t.Fatalf("emit metrics: %v", err)
	}
	return emitter.values
}

func TestRecordCacheMetrics(t *testing.T) {
	ResetForTests()
	RecordCacheHit("terraform", "mirror", CacheRefs)
	RecordCacheMiss("terraform", "mirror", CacheRefs)
	RecordCacheBytes("terraform", "mirror", CacheRefs, "serve", 512)
	RecordCacheBytes("terraform", "mirror", CacheRefs, "store", 256)
	ObserveUpstreamRequest("terraform", "mirror", "registry.terraform.io", http.StatusOK, nil, 50*time.Millisecond)
	ObserveUpstreamRequest("terraform", "mirror", "registry.terraform.io", 0, errors.New("boom"), 10*time.Millisecond)

	metrics := snapshotMetrics(t)
	hitKey := `repoxy_cache_events_total|type="terraform",repo="mirror",cache="refs",result="hit"`
	if got := metrics[hitKey]; got != 1 {
		t.Fatalf("expected 1 cache hit, got %v", got)
	}
	missKey := `repoxy_cache_events_total|type="terraform",repo="mirror",cache="refs",result="miss"`
	if got := metrics[missKey]; got != 1 {
		t.Fatalf("expected 1 cache miss, got %v", got)
	}
	bytesKey := `repoxy_cache_bytes_total|type="terraform",repo="mirror",cache="refs",action="serve"`
	if got := metrics[bytesKey]; got != 512 {
		t.Fatalf("expected 512 bytes served, got %v", got)
	}
	reqKey := `repoxy_upstream_requests_total|type="terraform",repo="mirror",target="registry.terraform.io",status="200"`
	if got := metrics[reqKey]; got != 1 {
		t.Fatalf("expected 1 successful upstream request, got %v", got)
	}
	errKey := `repoxy_upstream_requests_total|type="terraform",repo="mirror",target="registry.terraform.io",status="error"`
	if got := metrics[errKey]; got != 1 {
		t.Fatalf("expected 1 failed upstream request, got %v", got)
	}
}
