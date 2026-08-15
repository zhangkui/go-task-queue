package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-task-queue/internal/model"
	"go-task-queue/internal/service"
	"go-task-queue/internal/store"
)

func NewRouter(s *store.MemoryStore) http.Handler {
	svc := service.NewQueueService(s)
	mux := http.NewServeMux()
	mux.HandleFunc("/task/submit", handleSubmitTask(svc))
	mux.HandleFunc("/task/get", handleGetTask(svc))
	mux.HandleFunc("/task/status", handleGetTaskStatus(svc))
	mux.HandleFunc("/task/list", handleListTasks(svc))
	mux.HandleFunc("/task/execute", handleExecuteTask(svc))
	mux.HandleFunc("/task/retry", handleRetryTask(svc))
	mux.HandleFunc("/task/delete", handleDeleteTask(svc))
	mux.HandleFunc("/health", handleHealth())
	return mux
}

func handleSubmitTask(svc *service.QueueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Payload    string `json:"payload"`
			MaxRetries int    `json:"max_retries"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.ID == "" {
			writeError(w, http.StatusBadRequest, "id is required")
			return
		}
		task, err := svc.SubmitTask(r.Context(), req.ID, req.Name, req.Payload, req.MaxRetries)
		if err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, task)
	}
}

func handleGetTask(svc *service.QueueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "id is required")
			return
		}
		task, err := svc.GetTask(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, task)
	}
}

func handleGetTaskStatus(svc *service.QueueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "id is required")
			return
		}
		status, err := svc.GetTaskStatus(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": string(status)})
	}
}

func handleListTasks(svc *service.QueueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := model.TaskStatus(r.URL.Query().Get("status"))
		tasks := svc.ListTasks(r.Context(), status)
		writeJSON(w, http.StatusOK, tasks)
	}
}

func handleExecuteTask(svc *service.QueueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "id is required")
			return
		}
		if err := svc.ExecuteTask(r.Context(), id); err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleRetryTask(svc *service.QueueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "id is required")
			return
		}
		if err := svc.RetryTask(r.Context(), id); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleDeleteTask(svc *service.QueueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "id is required")
			return
		}
		if err := svc.DeleteTask(r.Context(), id); err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	}
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

var _ = strconv.Atoi
