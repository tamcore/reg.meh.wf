package hooks

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Auth(t *testing.T) {
	handler := NewHandler(nil, "test-token", 0, 0, slog.Default())

	t.Run("rejects missing auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/hook/registry-event", bytes.NewReader([]byte("{}")))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("rejects wrong token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/hook/registry-event", bytes.NewReader([]byte("{}")))
		req.Header.Set("Authorization", "Token wrong-token")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("rejects wrong method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/hook/registry-event", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rr.Code)
		}
	})
}

func TestHandler_EventParsing(t *testing.T) {
	// We can't test Redis interaction without a real Redis,
	// but we can test that the handler parses events correctly
	// by checking that it doesn't error on valid input (with nil redis it will fail,
	// so we just test the auth + decode path).

	t.Run("rejects invalid json", func(t *testing.T) {
		handler := NewHandler(nil, "tok", 0, 0, slog.Default())
		req := httptest.NewRequest(http.MethodPost, "/v1/hook/registry-event", bytes.NewReader([]byte("not json")))
		req.Header.Set("Authorization", "Token tok")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("accepts empty events", func(t *testing.T) {
		handler := NewHandler(nil, "tok", 0, 0, slog.Default())
		body, _ := json.Marshal(EventEnvelope{Events: []RegistryEvent{}})
		req := httptest.NewRequest(http.MethodPost, "/v1/hook/registry-event", bytes.NewReader(body))
		req.Header.Set("Authorization", "Token tok")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("skips non-push events", func(t *testing.T) {
		handler := NewHandler(nil, "tok", 0, 0, slog.Default())
		body, _ := json.Marshal(EventEnvelope{Events: []RegistryEvent{
			{Action: "pull", Target: EventTarget{Repository: "foo", Tag: "1h"}},
		}})
		req := httptest.NewRequest(http.MethodPost, "/v1/hook/registry-event", bytes.NewReader(body))
		req.Header.Set("Authorization", "Token tok")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("skips events with empty repo or tag", func(t *testing.T) {
		handler := NewHandler(nil, "tok", 0, 0, slog.Default())
		body, _ := json.Marshal(EventEnvelope{Events: []RegistryEvent{
			{Action: "push", Target: EventTarget{Repository: "", Tag: "1h"}},
			{Action: "push", Target: EventTarget{Repository: "foo", Tag: ""}},
		}})
		req := httptest.NewRequest(http.MethodPost, "/v1/hook/registry-event", bytes.NewReader(body))
		req.Header.Set("Authorization", "Token tok")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}
