package tf

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleWellKnownTerraform(t *testing.T) {
	t.Parallel()
	handler := &tfType{}
	req := httptest.NewRequest(http.MethodGet, "http://repoxy.test/.well-known/terraform.json", nil)
	req.Host = "repoxy.test"
	req.Header.Set("X-Forwarded-Proto", "https")
	rr := httptest.NewRecorder()

	handler.HandleWellKnownTerraform(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var payload map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	want := "https://repoxy.test/v1/providers/"
	if payload["providers.v1"] != want {
		t.Fatalf("expected providers.v1=%s, got %s", want, payload["providers.v1"])
	}
}
