package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-task-queue/internal/store"
)

func TestExecuteTaskNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/task/execute?id=missing", nil)
	recorder := httptest.NewRecorder()

	NewRouter(store.NewMemoryStore()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
	if got := recorder.Body.String(); got != "{\"error\":\"task not found\"}\n" {
		t.Fatalf("expected task not found response, got %q", got)
	}
}
